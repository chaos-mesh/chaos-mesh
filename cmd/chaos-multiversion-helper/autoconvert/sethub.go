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
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/common"
)

func setHub(hub string) error {
	fileSet := token.NewFileSet()

	apiDirectory := "api" + "/" + hub
	sources, err := ioutil.ReadDir(apiDirectory)
	if err != nil {
		return err
	}

	hubTypes := []string{}
	for _, file := range sources {
		// skip files which are not golang source code
		if !strings.HasSuffix(file.Name(), "go") {
			continue
		}

		filePath := apiDirectory + "/" + file.Name()
		fileAst, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		cmap := ast.NewCommentMap(fileSet, fileAst, fileAst.Comments)
		for node, commentGroups := range cmap {
			node, ok := node.(*ast.GenDecl)
			if !ok || node.Tok != token.TYPE {
				continue
			}

			for _, commentGroup := range commentGroups {
				isObjectRoot := false

				for _, comment := range commentGroup.List {
					if strings.Contains(comment.Text, "+kubebuilder:object:root=true") {
						isObjectRoot = true
					}
				}

				if isObjectRoot {
					hubTypes = append(hubTypes, node.Specs[0].(*ast.TypeSpec).Name.Name)
				}
			}
		}
	}

	hubFilePath := apiDirectory + "/" + "zz_generated.hub.chaosmesh.go"
	hubFile, err := os.Create(hubFilePath)
	if err != nil {
		return err
	}
	defer hubFile.Close()

	hubFile.WriteString(common.Boilerplate + "\n")
	hubFile.WriteString("package " + hub + "\n\n")
	for _, typ := range hubTypes {
		hubFile.WriteString("func (*" + typ + ") Hub() {}\n\n")
	}
	return nil
}
