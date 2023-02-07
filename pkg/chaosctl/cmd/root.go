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
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/recover"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
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
  chaosctl logs

  # forcedly recover chaos from pods
  chaosctl recover networkchaos pod1 -n test`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootLogger, err := log.NewDefaultZapLogger()
	if err != nil {
		cm.PrettyPrint("failed to initialize logger: ", 0, cm.Red)
		cm.PrettyPrint(err.Error(), 1, cm.Red)
		os.Exit(1)
	}

	rootCmd.PersistentFlags().StringVarP(&managerNamespace, "manager-namespace", "N", "chaos-mesh", "the namespace chaos-controller-manager in")
	rootCmd.PersistentFlags().StringVarP(&managerSvc, "manager-svc", "s", "chaos-mesh-controller-manager", "the service to chaos-controller-manager")
	err = rootCmd.RegisterFlagCompletionFunc("manager-namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// TODO: list namespaces without ctrlserver
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		cm.PrettyPrint("failed to register completion function for flag 'manager-namespace': ", 0, cm.Red)
		cm.PrettyPrint(err.Error(), 1, cm.Red)
		os.Exit(1)
	}
	err = rootCmd.RegisterFlagCompletionFunc("manager-svc", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// TODO: list svc without ctrlserver
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		cm.PrettyPrint("failed to register completion function for flag 'manager-svc': ", 0, cm.Red)
		cm.PrettyPrint(err.Error(), 1, cm.Red)
		os.Exit(1)
	}

	logsCmd, err := NewLogsCmd()
	if err != nil {
		cm.PrettyPrint("failed to initialize cmd: ", 0, cm.Red)
		cm.PrettyPrint("log command: "+err.Error(), 1, cm.Red)
		os.Exit(1)
	}

	rootCmd.AddCommand(logsCmd)

	debugCommand, err := NewDebugCommand(rootLogger.WithName("cmd-debug"), map[string]debug.Debug{
		networkChaos: debug.NetworkDebug,
		ioChaos:      debug.IODebug,
		stressChaos:  debug.StressDebug,
		httpChaos:    debug.HTTPDebug,
	})
	if err != nil {
		cm.PrettyPrint("failed to initialize cmd: ", 0, cm.Red)
		cm.PrettyPrint("debug command: "+err.Error(), 1, cm.Red)
		os.Exit(1)
	}

	recoverCommand, err := NewRecoverCommand(rootLogger.WithName("cmd-recover"), map[string]recover.RecovererBuilder{
		httpChaos:    recover.HTTPRecoverer,
		ioChaos:      recover.IORecoverer,
		stressChaos:  recover.StressRecoverer,
		networkChaos: recover.NetworkRecoverer,
	})
	if err != nil {
		cm.PrettyPrint("failed to initialize cmd: ", 0, cm.Red)
		cm.PrettyPrint("recover command: "+err.Error(), 1, cm.Red)
		os.Exit(1)
	}

	physicalMachineCommand, err := NewPhysicalMachineCommand()
	if err != nil {
		cm.PrettyPrint("failed to initialize cmd: ", 0, cm.Red)
		cm.PrettyPrint("physicalmachine command: "+err.Error(), 1, cm.Red)
		os.Exit(1)
	}

	rootCmd.AddCommand(debugCommand)
	rootCmd.AddCommand(recoverCommand)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(forwardCmd)
	rootCmd.AddCommand(physicalMachineCommand)

	if err := rootCmd.Execute(); err != nil {
		cm.PrettyPrint("failed to execute cmd: ", 0, cm.Red)
		cm.PrettyPrint(err.Error(), 1, cm.Red)
		os.Exit(1)
	}
}

func noCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
