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

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
)

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
	err := cm.SetupKlog()
	if err != nil {
		log.Fatal("failed to setup klog", err)
	}
	rootLogger, flushFunc, err := cm.NewStderrLogger()
	if err != nil {
		log.Fatal("failed to initialize logger", err)
	}
	if flushFunc != nil {
		defer flushFunc()
	}
	cm.SetupGlobalLogger(rootLogger.WithName("global-logger"))

	logsCmd, err := NewLogsCmd(rootLogger.WithName("cmd-logs"))
	if err != nil {
		rootLogger.Error(err, "failed to initialize cmd",
			"cmd", "logs",
			"errorVerbose", fmt.Sprintf("%+v", err),
		)
		os.Exit(1)
	}
	rootCmd.AddCommand(logsCmd)

	debugCommand, err := NewDebugCommand(rootLogger.WithName("cmd-debug"))
	if err != nil {
		rootLogger.Error(err, "failed to initialize cmd",
			"cmd", "debug",
			"errorVerbose", fmt.Sprintf("%+v", err),
		)
		os.Exit(1)
	}

	rootCmd.AddCommand(debugCommand)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(forwardCmd)
	if err := rootCmd.Execute(); err != nil {
		rootLogger.Error(err, "failed to execute cmd",
			"errorVerbose", fmt.Sprintf("%+v", err),
		)
	}
}

func noCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
