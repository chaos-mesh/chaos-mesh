// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/pingcap/errors"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	log = zap.Logger(true)
)

type metadata struct {
	Type string
}

func main() {
	generatedCode := `// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

	`

	initImpl := ""
	filepath.Walk("./api/v1alpha1", func(path string, info os.FileInfo, err error) error {
		log := log.WithValues("file", path)

		if err != nil {
			log.Error(err, "fail to walk in directory")
			return err
		}
		if info.IsDir() {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			log.Error(err, "fail to parse file")
		}

		cmap := ast.NewCommentMap(fset, file, file.Comments)

	out:
		for node, commentGroups := range cmap {
			for _, commentGroup := range commentGroups {
				var err error
				for _, comment := range commentGroup.List {
					if strings.Contains(comment.Text, "+chaos-mesh:base") {
						log.Info("build")
						baseDecl, ok := node.(*ast.GenDecl)
						if !ok {
							err = errors.Errorf("node is not a *ast.GenDecl")
							log.Error(err, "fail to get type")
							continue
						}

						if baseDecl.Tok != token.TYPE {
							err = errors.Errorf("node.Tok is not token.TYPE")
							log.Error(err, "fail to get type")
							continue
						}

						baseType, ok := baseDecl.Specs[0].(*ast.TypeSpec)
						if !ok {
							err = errors.Errorf("node is not a *ast.TypeSpec")
							log.Error(err, "fail to get type")
							continue
						}

						generatedCode += generateImpl(baseType.Name.Name)
						initImpl += generateInit(baseType.Name.Name)
						break out
					}
				}
			}
		}

		return nil
	})

	generatedCode += fmt.Sprintf(`
func init() {
%s
}
`, initImpl)
	file, err := os.Create("./api/v1alpha1/zz_generated.chaosmesh.go")
	if err != nil {
		log.Error(err, "fail to create file")
	}
	fmt.Fprint(file, generatedCode)

	return
}
