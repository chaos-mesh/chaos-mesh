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

package create

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCreateCmd(log logr.Logger) *cobra.Command {
	var from, to string
	var asStorageVersion bool

	var cmd = &cobra.Command{
		Use:   "create --from <old-version> --to <new-version>",
		Short: "create command create a new version of chaos api",
		Run: func(cmd *cobra.Command, args []string) {
			err := run(log, from, to, asStorageVersion)
			if err != nil {
				log.Error(err, "create new version")
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "old version of chaos api")
	cmd.Flags().StringVar(&to, "to", "", "new version of chaos api")
	cmd.Flags().BoolVar(&asStorageVersion, "as-storage-version", true, "new version of chaos api")

	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}

func run(log logr.Logger, from, to string, asStorageVersion bool) error {
	fileSet := token.NewFileSet()

	oldAPIDirectory := "api/" + from
	newAPIDirectory := "api/" + to
	os.Mkdir(newAPIDirectory, 0755)

	oldFiles, err := ioutil.ReadDir(oldAPIDirectory)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, file := range oldFiles {
		// skip files which are not golang source code
		if !strings.HasSuffix(file.Name(), "go") {
			continue
		}

		oldFilePath := oldAPIDirectory + "/" + file.Name()
		fileAst, err := parser.ParseFile(fileSet, oldFilePath, nil, parser.ParseComments)
		if err != nil {
			return errors.WithStack(err)
		}

		// modify the package name to the new version
		fileAst.Name.Name = to

		if asStorageVersion {
			// automatically add the storage version annotation
			cmap := ast.NewCommentMap(fileSet, fileAst, fileAst.Comments)
			for node, commentGroups := range cmap {
				node, ok := node.(*ast.GenDecl)
				if !ok || node.Tok != token.TYPE {
					continue
				}

				for _, commentGroup := range commentGroups {
					isObjectRoot := false
					isStorageVersion := false

					for _, comment := range commentGroup.List {
						if strings.Contains(comment.Text, "+kubebuilder:object:root=true") {
							isObjectRoot = true
						}
						if strings.Contains(comment.Text, "+kubebuilder:storageversion") {
							isStorageVersion = true
						}
					}

					if isObjectRoot && !isStorageVersion {
						commentGroup.List = append(commentGroup.List, &ast.Comment{
							Text:  "// +kubebuilder:storageversion",
							Slash: node.Pos() - 1,
						})
					}
				}
			}

			// `ast` package is not suitable for removing a comment, so the
			// traditional string processing tool is prefered :)
			sedProcess := exec.Command("sed", "-i", "/+kubebuilder:storageversion/d", oldFilePath)
			sedProcess.Start()

			err = sedProcess.Wait()
			if err != nil {
				return errors.WithStack(err)
			}
		}

		newFile, err := os.Create(newAPIDirectory + "/" + file.Name())
		if err != nil {
			return errors.WithStack(err)
		}
		defer newFile.Close()
		printer.Fprint(newFile, fileSet, fileAst)
	}
	return nil
}
