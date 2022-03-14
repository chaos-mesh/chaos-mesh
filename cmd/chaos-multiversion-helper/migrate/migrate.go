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

package migrate

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"

	doublestar "github.com/bmatcuk/doublestar/v4"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/common"
)

func NewMigrateCmd(log logr.Logger) *cobra.Command {
	var from, to string

	cmd := &cobra.Command{
		Use:   "migrate --from <old-version> --to <new-version>",
		Short: "migrate command iterate over all golang source codes (except a whitelist) and migrate them to the new version",
		Run: func(cmd *cobra.Command, args []string) {
			err := run(from, to)
			if err != nil {
				log.Error(err, "migrate source codes", "from", from, "to", to)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "old version of chaos api")
	cmd.Flags().StringVar(&to, "to", "", "new version of chaos api")

	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}

var whiteListPattern = []string{
	".cache/**/*",
	"api/**/*",
}

func isWhiteListed(path string) bool {
	for _, pattern := range whiteListPattern {
		match, _ := doublestar.PathMatch(pattern, path)
		// ignore error, because the only possible error is bad pattern

		if match {
			return true
		}
	}

	return false
}

func run(from, to string) error {
	allGoFiles, err := doublestar.Glob(os.DirFS("."), "**/*.go")
	if err != nil {
		return errors.WithStack(err)
	}

	for _, file := range allGoFiles {
		if isWhiteListed(file) {
			continue
		}

		err := migrateFile(file, from, to)
		if err != nil {
			return err
		}
	}

	return nil
}

func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

func migrateFile(path string, from string, to string) error {
	fileSet := token.NewFileSet()

	fileAst, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		return errors.WithStack(err)
	}

	needMigrate := false
	for _, imp := range fileAst.Imports {
		if imp.Path.Value == quote(common.ChaosMeshAPIPrefix+from) {
			if imp.Name == nil {
				imp.Path.Value = quote(common.ChaosMeshAPIPrefix + to)

				needMigrate = true
			}
		}
	}
	if needMigrate {
		// do migration for old package name
		ast.Inspect(fileAst, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.SelectorExpr:
				ident, ok := node.X.(*ast.Ident)
				if !ok {
					break
				}

				if ident.Name == from {
					ident.Name = to
				}
			}

			return true
		})

		file, err := os.Create(path)
		if err != nil {
			return errors.WithStack(err)
		}
		defer file.Close()
		printer.Fprint(file, fileSet, fileAst)
	}
	return nil
}
