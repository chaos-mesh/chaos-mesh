// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/zapr"
	"github.com/pingcap/errors"
	"go.uber.org/zap"
)

var (
	zapLogger, _ = zap.NewDevelopment()
	log          = zapr.NewLogger(zapLogger)
)

type metadata struct {
	Type         string
	OneShotExp   string
	IsExperiment bool
	EnableUpdate bool
}

func main() {
	if err := generate("."); err != nil {
		log.Error(err, "generate chaos API")
		os.Exit(1)
	}
}

func generate(root string) error {
	var implCode strings.Builder
	implCode.WriteString(boilerplate + implImport)

	var testCode strings.Builder
	testCode.WriteString(boilerplate + testImport)
	var initImpl strings.Builder
	var scheduleImpl strings.Builder

	workflowGenerator := newWorkflowCodeGenerator(nil)
	workflowTestGenerator := newWorkflowTestCodeGenerator(nil)

	scheduleGenerator := newScheduleCodeGenerator(nil)

	frontendGenerator := newFrontendCodeGenerator(nil)

	apiDir := filepath.Join(root, "api", "v1alpha1")
	err := filepath.Walk(apiDir, func(path string, info os.FileInfo, err error) error {
		log := log.WithValues("file", path)

		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		if strings.HasPrefix(info.Name(), "zz_generated") {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		cmap := ast.NewCommentMap(fset, file, file.Comments)

	out:
		for node, commentGroups := range cmap {
			for _, commentGroup := range commentGroups {
				var oneShotExp string
				var enableUpdate bool

				for _, comment := range commentGroup.List {
					if strings.Contains(comment.Text, "+chaos-mesh:webhook:enableUpdate") {
						enableUpdate = true
						break
					}
				}

				for _, comment := range commentGroup.List {
					if strings.Contains(comment.Text, "+chaos-mesh:base") {
						baseType, err := getType(fset, node, comment)
						if err != nil {
							return err
						}
						if baseType.Name.Name != "Workflow" {
							implCode.WriteString(generateImpl(baseType.Name.Name, oneShotExp, false, enableUpdate))
							initImpl.WriteString(generateInit(baseType.Name.Name, false))
						}
						continue out
					}
				}

				for _, comment := range commentGroup.List {
					if strings.Contains(comment.Text, "+chaos-mesh:oneshot") {
						oneShotExp = strings.TrimPrefix(comment.Text, "// +chaos-mesh:oneshot=")
						log.Info("decode oneshot expression", "expression", oneShotExp)
					}
				}
				for _, comment := range commentGroup.List {
					if strings.Contains(comment.Text, "+chaos-mesh:experiment") {
						baseType, err := getType(fset, node, comment)
						if err != nil {
							return err
						}

						if baseType.Name.Name != "Workflow" {
							implCode.WriteString(generateImpl(baseType.Name.Name, oneShotExp, true, enableUpdate))
							initImpl.WriteString(generateInit(baseType.Name.Name, true))
							testCode.WriteString(generateTest(baseType.Name.Name))
							workflowGenerator.AppendTypes(baseType.Name.Name)
							workflowTestGenerator.AppendTypes(baseType.Name.Name)
							frontendGenerator.AppendTypes(baseType.Name.Name)
						}
						scheduleImpl.WriteString(generateScheduleRegister(baseType.Name.Name))
						scheduleGenerator.AppendTypes(baseType.Name.Name)
						continue out
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "read chaos API declarations")
	}

	implCode.WriteString(fmt.Sprintf(`
func init() {
%s
%s
}
`, initImpl.String(), scheduleImpl.String()))
	testCode.WriteString(testInit)
	for _, output := range []struct {
		path    string
		content string
	}{
		{filepath.Join(apiDir, "zz_generated.chaosmesh.go"), implCode.String()},
		{filepath.Join(apiDir, "zz_generated.chaosmesh_test.go"), testCode.String()},
		{filepath.Join(apiDir, "zz_generated.workflow.chaosmesh.go"), workflowGenerator.Render()},
		{filepath.Join(apiDir, "zz_generated.workflow.chaosmesh_test.go"), workflowTestGenerator.Render()},
		{filepath.Join(apiDir, "zz_generated.schedule.chaosmesh.go"), scheduleGenerator.Render()},
		{filepath.Join(root, "ui", "app", "src", "api", "zz_generated.frontend.chaos-mesh.ts"), frontendGenerator.Render()},
	} {
		if err := os.WriteFile(output.path, []byte(output.content), 0o666); err != nil {
			return errors.Wrapf(err, "write %s", output.path)
		}
	}
	return nil
}

func getType(fset *token.FileSet, node ast.Node, comment *ast.Comment) (*ast.TypeSpec, error) {
	log.Info("build", "pos", fset.Position(comment.Pos()))
	decl, ok := node.(*ast.GenDecl)
	if !ok {
		err := errors.New("node is not a *ast.GenDecl")
		log.Error(err, "fail to get type")
		return nil, err
	}

	if decl.Tok != token.TYPE {
		err := errors.New("node.Tok is not token.TYPE")
		log.Error(err, "fail to get type")
		return nil, err
	}

	baseType, ok := decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		err := errors.New("node is not a *ast.TypeSpec")
		log.Error(err, "fail to get type")
		return nil, err
	}
	return baseType, nil
}
