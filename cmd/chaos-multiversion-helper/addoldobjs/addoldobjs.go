// Copyright 2022 Chaos Mesh Authors.
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

package addoldobjs

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/common"
)

func NewAddOldObjsCmd(log logr.Logger) *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "addoldobjs --version <version>",
		Short: "addoldobjs command automatically add the old version objs to `cmd/chaos-controller-manager/provider/convert.go`",
		Run: func(cmd *cobra.Command, args []string) {
			err := run(log, version)
			if err != nil {
				log.Error(err, "add old objects")
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "the version to iterate and add to convert")
	cmd.MarkFlagRequired("version")

	return cmd
}

func run(log logr.Logger, version string) error {
	types, err := getOldTypes(version)
	if err != nil {
		return err
	}

	fileSet := token.NewFileSet()
	filePath := "cmd/chaos-controller-manager/provider/convert.go"
	fileAst, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)

	if err != nil {
		return errors.WithStack(err)
	}
	ast.Inspect(fileAst, func(n ast.Node) bool {
		node, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}
		if node.Tok == token.IMPORT {
			imported := false

			for _, spec := range node.Specs {
				if spec.(*ast.ImportSpec).Path.Value == common.ChaosMeshAPIPrefix+"/"+version {
					imported = true
				}
			}

			if !imported {
				node.Specs = append(node.Specs, &ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: `"` + common.ChaosMeshAPIPrefix + version + `"`,
					},
				})
			}

			return false
		} else if node.Tok == token.VAR {
			if valueSpec, ok := node.Specs[0].(*ast.ValueSpec); ok && valueSpec.Names[0].Name == "oldObjs" {
				sliceLit, ok := valueSpec.Values[0].(*ast.CompositeLit)
				if !ok {
					err := errors.New("oldObjs is not a slice")
					log.Error(err, "oldObjs is not a slice")
				}

				included := map[string]struct{}{}
				for _, elt := range sliceLit.Elts {
					referenceExpr, ok := elt.(*ast.UnaryExpr)
					if !ok {
						continue
					}

					if referenceExpr.Op == token.AND {
						if structLit, ok := referenceExpr.X.(*ast.CompositeLit); ok {
							if structLit.Type == nil {
								continue
							}

							typeName := structLit.Type.(*ast.SelectorExpr).Sel.Name
							included[typeName] = struct{}{}
						}
					}
				}
				for _, typ := range types {
					if _, ok := included[typ]; !ok {
						sliceLit.Elts = append(sliceLit.Elts, &ast.UnaryExpr{
							Op: token.AND,
							X: &ast.CompositeLit{
								Type: &ast.SelectorExpr{
									X:   &ast.Ident{Name: "v1alpha1"},
									Sel: &ast.Ident{Name: typ},
								},
							},
						})
					}
				}
			}
		} else {
			return true
		}

		return false
	})

	newFile, err := os.OpenFile(filePath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer newFile.Close()
	return printer.Fprint(newFile, fileSet, fileAst)
}

func getOldTypes(version string) ([]string, error) {
	fileSet := token.NewFileSet()

	apiDirectory := "api" + "/" + version
	sources, err := ioutil.ReadDir(apiDirectory)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	types := []string{}
	for _, file := range sources {
		// skip files which are not golang source code
		if !strings.HasSuffix(file.Name(), "go") {
			continue
		}

		filePath := apiDirectory + "/" + file.Name()
		fileAst, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// read the comment map to decide which types need to be converted
		cmap := ast.NewCommentMap(fileSet, fileAst, fileAst.Comments)
		for node, commentGroups := range cmap {
			node, ok := node.(*ast.GenDecl)
			if !ok || node.Tok != token.TYPE {
				continue
			}

			typeName := node.Specs[0].(*ast.TypeSpec).Name.Name
			for _, commentGroup := range commentGroups {
				isObjectRoot := false

				for _, comment := range commentGroup.List {
					if strings.Contains(comment.Text, "+kubebuilder:object:root=true") {
						isObjectRoot = true
					}
				}

				if isObjectRoot {
					types = append(types, typeName)
				}
			}
		}
	}

	return types, nil
}
