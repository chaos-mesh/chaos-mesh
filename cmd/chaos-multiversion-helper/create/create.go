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
	"io/fs"
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
		Short: "Create a new version of chaos api",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(log, from, to, asStorageVersion)
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
	err := os.Mkdir(newAPIDirectory, 0755)
	if err != nil {
		return errors.Wrapf(err, "create directory %s", newAPIDirectory)
	}

	oldFiles, err := ioutil.ReadDir(oldAPIDirectory)
	if err != nil {
		return errors.Wrapf(err, "read directory %s", oldAPIDirectory)
	}

	ctx := createFileContext{
		oldAPIDirectory,
		newAPIDirectory,
		to,
		asStorageVersion,
		fileSet,
		log,
	}
	for _, file := range oldFiles {
		log.Info("copy api", "oldFile", file.Name())
		ctx.log = log.WithValues("fileName", file.Name())
		err := createFile(ctx, file)
		if err != nil {
			return err
		}
	}
	log.Info("create new api successfully")
	return nil
}

type createFileContext struct {
	oldAPIDirectory string
	newAPIDirectory string
	to              string

	asStorageVersion bool
	fileSet          *token.FileSet

	log logr.Logger
}

func createFile(ctx createFileContext, file fs.FileInfo) error {
	// skip files which are not golang source code
	if !strings.HasSuffix(file.Name(), "go") {
		return nil
	}

	oldFilePath := ctx.oldAPIDirectory + "/" + file.Name()
	fileAst, err := parser.ParseFile(ctx.fileSet, oldFilePath, nil, parser.ParseComments)
	if err != nil {
		return errors.Wrapf(err, "parse file %s", oldFilePath)
	}

	// modify the package name to the new version
	fileAst.Name.Name = ctx.to

	if ctx.asStorageVersion {
		// automatically add the storage version annotation
		cmap := ast.NewCommentMap(ctx.fileSet, fileAst, fileAst.Comments)
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
		// traditional string processing tool is preferred :)
		sedProcess := exec.Command("sed", "-i", "/+kubebuilder:storageversion/d", oldFilePath)
		err = sedProcess.Run()
		if err != nil {
			return errors.Wrapf(err, "remove storage version for %s", oldFilePath)
		}
	}

	newFilePath := ctx.newAPIDirectory + "/" + file.Name()
	newFile, err := os.Create(newFilePath)
	if err != nil {
		return errors.Wrapf(err, "create file %s", newFilePath)
	}
	defer newFile.Close()
	printer.Fprint(newFile, ctx.fileSet, fileAst)

	return nil
}
