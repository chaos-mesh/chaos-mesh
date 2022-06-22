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
		Short: "Iterate over all Golang source codes (except a whitelist) and migrate them to the new version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(log, from, to)
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "old version of chaos api")
	cmd.Flags().StringVar(&to, "to", "", "new version of chaos api")

	err := cmd.MarkFlagRequired("from")
	if err != nil {
		log.Error(errors.WithStack(err), "fail to mark 'from' as required")
		panic("unreachable")
	}
	err = cmd.MarkFlagRequired("to")
	if err != nil {
		log.Error(errors.WithStack(err), "fail to mark 'to' as required")
		panic("unreachable")
	}

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

func run(log logr.Logger, from, to string) error {
	err := common.ValidateAPI(from)
	if err != nil {
		return err
	}

	err = common.ValidateAPI(to)
	if err != nil {
		return err
	}

	allGoFiles, err := doublestar.Glob(os.DirFS("."), "**/*.go")
	if err != nil {
		return errors.WithStack(err)
	}

	for _, file := range allGoFiles {
		if isWhiteListed(file) {
			continue
		}

		err := migrateFile(log.WithValues("file", file, "from", from, "to", to), file, from, to)
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateFile(log logr.Logger, path string, from string, to string) error {
	fileSet := token.NewFileSet()

	fileAst, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		return errors.WithStack(err)
	}

	if migrateAst(log, fileAst, from, to) {
		file, err := os.Create(path)
		if err != nil {
			return errors.WithStack(err)
		}
		defer file.Close()
		printer.Fprint(file, fileSet, fileAst)
	}

	return nil
}

// migrateAst migrates the ast, and returns whether this ast has been modified
func migrateAst(log logr.Logger, fileAst *ast.File, from string, to string) bool {
	needMigrate := false
	fromName := ""
	for _, imp := range fileAst.Imports {
		if imp.Path.Value == common.Quote(common.ChaosMeshAPIPrefix+from) {
			if imp.Name == nil {
				fromName = from
			} else {
				fromName = imp.Name.Name
			}

			imp.Path.Value = common.Quote(common.ChaosMeshAPIPrefix + to)
			imp.Name = &ast.Ident{
				Name: formatChaosMeshAPI(to),
			}
			needMigrate = true
		}
	}
	if needMigrate {
		log.Info("migrate file")
		// do migration for old package name
		ast.Inspect(fileAst, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.SelectorExpr:
				ident, ok := node.X.(*ast.Ident)
				if !ok {
					break
				}

				if ident.Name == fromName {
					ident.Name = formatChaosMeshAPI(to)
				}
			}

			return true
		})

		return true
	}

	return false
}

func formatChaosMeshAPI(to string) string {
	return "chaosmeshapi" + to
}
