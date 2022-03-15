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
	implCode := boilerplate + implImport

	testCode := boilerplate + testImport
	initImpl := ""
	scheduleImpl := ""

	workflowGenerator := newWorkflowCodeGenerator(nil)
	workflowTestGenerator := newWorkflowTestCodeGenerator(nil)

	scheduleGenerator := newScheduleCodeGenerator(nil)

	frontendGenerator := newFrontendCodeGenerator(nil)

	filepath.Walk("./api/v1alpha1", func(path string, info os.FileInfo, err error) error {
		log := log.WithValues("file", path)

		if err != nil {
			log.Error(err, "fail to walk in directory")
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
			log.Error(err, "fail to parse file")
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
						if strings.Contains(comment.Text, "+chaos-mesh:base") {
							if baseType.Name.Name != "Workflow" {
								implCode += generateImpl(baseType.Name.Name, oneShotExp, false, enableUpdate)
								initImpl += generateInit(baseType.Name.Name, false)
							}
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
							implCode += generateImpl(baseType.Name.Name, oneShotExp, true, enableUpdate)
							initImpl += generateInit(baseType.Name.Name, true)
							testCode += generateTest(baseType.Name.Name)
							workflowGenerator.AppendTypes(baseType.Name.Name)
							workflowTestGenerator.AppendTypes(baseType.Name.Name)
							frontendGenerator.AppendTypes(baseType.Name.Name)
						}
						scheduleImpl += generateScheduleRegister(baseType.Name.Name)
						scheduleGenerator.AppendTypes(baseType.Name.Name)
						continue out
					}
				}
			}
		}

		return nil
	})

	implCode += fmt.Sprintf(`
func init() {
%s
%s
}
`, initImpl, scheduleImpl)
	file, err := os.Create("./api/v1alpha1/zz_generated.chaosmesh.go")
	if err != nil {
		log.Error(err, "fail to create file")
		os.Exit(1)
	}
	fmt.Fprint(file, implCode)

	testCode += testInit
	file, err = os.Create("./api/v1alpha1/zz_generated.chaosmesh_test.go")
	if err != nil {
		log.Error(err, "fail to create file")
		os.Exit(1)
	}
	fmt.Fprint(file, testCode)

	file, err = os.Create("./api/v1alpha1/zz_generated.workflow.chaosmesh.go")
	if err != nil {
		log.Error(err, "fail to create file")
		os.Exit(1)
	}
	fmt.Fprint(file, workflowGenerator.Render())

	file, err = os.Create("./api/v1alpha1/zz_generated.workflow.chaosmesh_test.go")
	if err != nil {
		log.Error(err, "fail to create file")
		os.Exit(1)
	}
	fmt.Fprint(file, workflowTestGenerator.Render())

	file, err = os.Create("./api/v1alpha1/zz_generated.schedule.chaosmesh.go")
	if err != nil {
		log.Error(err, "fail to create file")
		os.Exit(1)
	}
	fmt.Fprint(file, scheduleGenerator.Render())

	file, err = os.Create("./ui/app/src/api/zz_generated.frontend.chaos-mesh.ts")
	if err != nil {
		log.Error(err, "fail to create file")
		os.Exit(1)
	}
	fmt.Fprint(file, frontendGenerator.Render())
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
