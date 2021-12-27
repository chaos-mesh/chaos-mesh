// Copyright 2021 Chaos Mesh Authors.
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

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/httpchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/iochaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/networkchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/stresschaos"
)

var managerNamespace, managerSvc string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chaosctl [command] [options]",
	Short: "Interacting with chaos mesh",
	Long: `
Interacting with chaos mesh

  # show debug info
  chaosctl debug networkchaos

  # show logs of all chaos-mesh components
  chaosctl logs`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootLogger, flushFunc, err := cm.NewStderrLogger()
	if err != nil {
		cm.PrettyPrint("failed to initialize logger: ", 0, cm.Red)
		cm.PrettyPrint(err.Error(), 1, cm.Red)
		os.Exit(1)
	}
	if flushFunc != nil {
		defer flushFunc()
	}
	cm.SetupGlobalLogger(rootLogger.WithName("global-logger"))

	logsCmd, err := NewLogsCmd(rootLogger.WithName("cmd-logs"))
	if err != nil {
		cm.PrettyPrint("failed to initialize cmd: ", 0, cm.Red)
		cm.PrettyPrint("log command: "+err.Error(), 1, cm.Red)
		os.Exit(1)
	}
	rootCmd.AddCommand(logsCmd)

	debugCommand, err := NewDebugCommand(rootLogger.WithName("cmd-debug"), map[string]Debugger{
		networkChaos: networkchaos.Debug,
		ioChaos:      iochaos.Debug,
		stressChaos:  stresschaos.Debug,
		httpChaos:    httpchaos.Debug,
	})
	if err != nil {
		cm.PrettyPrint("failed to initialize cmd: ", 0, cm.Red)
		cm.PrettyPrint("debug command: "+err.Error(), 1, cm.Red)
		os.Exit(1)
	}

	rootCmd.Flags().StringVarP(&managerNamespace, "manager-namespace", "n", "chaos-testing", "the namespace chaos-controller-manager in")
	rootCmd.Flags().StringVarP(&managerSvc, "manager-svc", "s", "chaos-mesh-controller-manager", "the service to chaos-controller-manager")
	rootCmd.AddCommand(debugCommand)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(forwardCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(NewQueryCmd(rootLogger.WithName("query")))
	if err := rootCmd.Execute(); err != nil {
		cm.PrettyPrint("failed to execute cmd: ", 0, cm.Red)
		cm.PrettyPrint(err.Error(), 1, cm.Red)
		os.Exit(1)
	}
}

func noCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
