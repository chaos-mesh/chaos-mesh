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

package autoconvert

import (
	"go/ast"
	"testing"
)

func TestPrintExprToNewVersion(t *testing.T) {
	type testCase struct {
		Name   string
		Expr   ast.Expr
		Result string

		version string
		hub     string
	}

	cases := []testCase{
		{
			Name: "builtin types",
			Expr: &ast.Ident{
				Name: "int",
			},
			Result:  "int",
			version: "v1alpha1",
			hub:     "v1alpha2",
		},
		{
			Name: "custom type",
			Expr: &ast.Ident{
				Name: "PodChaosSpec",
			},
			Result:  "v1alpha2.PodChaosSpec",
			version: "v1alpha1",
			hub:     "v1alpha2",
		},
		{
			Name: "import from other package",
			Expr: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "ctrl",
				},
				Sel: &ast.Ident{
					Name: "Manager",
				},
			},
			Result:  "ctrl.Manager",
			version: "v1alpha1",
			hub:     "v1alpha2",
		},
		{
			Name: "star builtin types",
			Expr: &ast.StarExpr{
				X: &ast.Ident{
					Name: "int",
				},
			},
			Result:  "*int",
			version: "v1alpha1",
			hub:     "v1alpha2",
		},
		{
			Name: "star custom types",
			Expr: &ast.StarExpr{
				X: &ast.Ident{
					Name: "PodChaosSpec",
				},
			},
			Result:  "*v1alpha2.PodChaosSpec",
			version: "v1alpha1",
			hub:     "v1alpha2",
		},
		{
			Name: "star array",
			Expr: &ast.StarExpr{
				X: &ast.ArrayType{
					Elt: &ast.Ident{
						Name: "PodChaos",
					},
				},
			},
			Result:  "*[]v1alpha2.PodChaos",
			version: "v1alpha1",
			hub:     "v1alpha2",
		},
		{
			Name: "star map to array",
			Expr: &ast.StarExpr{
				X: &ast.MapType{
					Key: &ast.Ident{
						Name: "PodChaos",
					},
					Value: &ast.Ident{
						Name: "BlockChaos",
					},
				},
			},
			Result:  "*map[v1alpha2.PodChaos]v1alpha2.BlockChaos",
			version: "v1alpha1",
			hub:     "v1alpha2",
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			impl := convertImpl{
				version: c.version,
				hub:     c.hub,
			}

			result, err := impl.printExprToNewVersion(c.Expr)
			if err != nil {
				t.Fatal(err.Error())
			}
			if result != c.Result {
				t.Fatalf("expect: %s, got: %s", c.Result, result)
			}
		})
	}
}
