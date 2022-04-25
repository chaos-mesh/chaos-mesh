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

package registerscheme

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/common"
)

func NewRegisterSchemeCmd(log logr.Logger) *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "registerscheme --version <version>",
		Short: "Automatically add the scheme to `cmd/chaos-controller-manager/provider/controller.go`",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(log, version)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "the version of scheme to add")
	err := cmd.MarkFlagRequired("version")
	if err != nil {
		log.Error(errors.WithStack(err), "fail to mark 'version' as required")
		panic("unreachable")
	}

	return cmd
}

func run(log logr.Logger, version string) error {
	err := common.ValidateAPI(version)
	if err != nil {
		return err
	}

	fileSet := token.NewFileSet()
	filePath := "cmd/chaos-controller-manager/provider/controller.go"
	fileAst, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	shouldAddImport := true
	for _, imp := range fileAst.Imports {
		if imp.Path.Value == common.Quote(common.ChaosMeshAPIPrefix+version) {
			shouldAddImport = false
		}
	}

	ast.Inspect(fileAst, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			if decl.Name.Name == "init" {
				decl.Body.List = append(decl.Body.List, &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "_",
						},
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: version,
								},
								Sel: &ast.Ident{
									Name: "AddToScheme",
								},
							},
							Args: []ast.Expr{
								&ast.Ident{
									Name: "scheme",
								},
							},
						},
					},
				})
				return false
			}
		case *ast.GenDecl:
			if shouldAddImport && decl.Tok == token.IMPORT {
				decl.Specs = append(decl.Specs, &ast.ImportSpec{
					Path: &ast.BasicLit{
						Value: common.Quote(common.ChaosMeshAPIPrefix + version),
					},
				})
			}
		default:
			return true
		}

		return false
	})

	newFile, err := os.Create(filePath)
	if err != nil {
		return errors.Wrapf(err, "open file %s", filePath)
	}
	defer newFile.Close()
	return printer.Fprint(newFile, fileSet, fileAst)
}
