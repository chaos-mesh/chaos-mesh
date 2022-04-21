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

package main

import (
	stdlog "log"
	"os"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/addoldobjs"
	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/autoconvert"
	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/create"
	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/migrate"
	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/registerscheme"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
)

var rootCmd = &cobra.Command{
	Use:   "chaos-multiversion-helper",
	Short: "chaos-multiversion-helper is a command to help developer to create a new version of chaos api",
}

func main() {
	rootLogger, err := log.NewDefaultZapLogger()
	if err != nil {
		stdlog.Fatal("failed to create root logger", err)
	}
	rootLogger = rootLogger.WithName("chaos-multiversion-helper")

	rootCmd.AddCommand(create.NewCreateCmd(rootLogger.WithName("create")))
	rootCmd.AddCommand(migrate.NewMigrateCmd(rootLogger.WithName("migrate")))
	rootCmd.AddCommand(autoconvert.NewConvertCmd(rootLogger.WithName("autoconvert")))
	rootCmd.AddCommand(addoldobjs.NewAddOldObjsCmd(rootLogger.WithName("addoldobjs")))
	rootCmd.AddCommand(registerscheme.NewRegisterSchemeCmd(rootLogger.WithName("registerscheme")))

	if err := rootCmd.Execute(); err != nil {
		rootLogger.Error(err, "execute command")
		os.Exit(1)
	}

	rootLogger.Info("execute command successfully")
}
