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
	"log"
	"os"

	doublestar "github.com/bmatcuk/doublestar/v4"

	"github.com/spf13/cobra"
)

var MigrateCmd = &cobra.Command{
	Use:   "migrate --from <old-version> --to <new-version>",
	Short: "migrate command iterate over all golang source codes (except a whitelist) and migrate them to the new version",
	Run: func(cmd *cobra.Command, args []string) {
		err := run()
		if err != nil {
			log.Fatal(err)
		}
	},
}

var from, to string

var whiteListPattern = []string{
	".cache/**/*",
	"api/**/*",
}

const chaosMeshAPIPrefix = "github.com/chaos-mesh/chaos-mesh/api/"

func init() {
	MigrateCmd.Flags().StringVar(&from, "from", "", "old version of chaos api")
	MigrateCmd.Flags().StringVar(&to, "to", "", "new version of chaos api")

	MigrateCmd.MarkFlagRequired("from")
	MigrateCmd.MarkFlagRequired("to")
}

func isWhiteListed(path string) bool {
	for _, pattern := range whiteListPattern {
		match, err := doublestar.PathMatch(pattern, path)
		if err != nil {
			log.Fatal(err)
		}

		if match {
			return true
		}
	}

	return false
}

func run() error {
	allGoFiles, err := doublestar.Glob(os.DirFS("."), "**/*.go")
	if err != nil {
		return err
	}

	for _, file := range allGoFiles {
		if isWhiteListed(file) {
			continue
		}

		err := migrateFile(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

func migrateFile(path string) error {
	fileSet := token.NewFileSet()

	fileAst, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	needMigrate := false
	for _, imp := range fileAst.Imports {
		if imp.Path.Value == quote(chaosMeshAPIPrefix+from) {
			if imp.Name == nil {
				imp.Path.Value = quote(chaosMeshAPIPrefix + to)

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
			return err
		}
		printer.Fprint(file, fileSet, fileAst)
	}
	return nil
}
