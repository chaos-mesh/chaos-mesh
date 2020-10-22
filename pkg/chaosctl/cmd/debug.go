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
	"strings"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/pkg/debug/networkchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/debug/stresschaos"
)

type DebugOptions struct {
	ChaosName   string
	Namespace   string
	BuilderArgs []string
}

// debugCmd represents the debug command

func init() {
	o := &DebugOptions{}

	var debugCmd = &cobra.Command{
		Use:   `debug (CHAOSTYPE) [-c CHAOSNAME] [-n NAMESPACE]`,
		Short: `Print the debug information for certain chaos`,
		Long: `Print the debug information for certain chaos.
Currently only support networkchaos and stresschaos.

Examples:
  # Return debug information from all networkchaos
  chaosctl debug networkchaos

  # Return debug information from certain networkchaos
  chaosctl debug networkchaos -c web-show-network-delay -n chaos-testing`,
		Run: func(cmd *cobra.Command, args []string) {
			o.BuilderArgs = args
			if err := o.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.AddCommand(debugCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// debugCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	debugCmd.Flags().StringVarP(&o.ChaosName, "chaosname", "c", "", "certain chaos name")
	debugCmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "default", "namespace to find chaos")
}

func (o *DebugOptions) Run() error {
	if len(o.BuilderArgs) == 0 {
		return fmt.Errorf("You must specify the type of chaos")
	}
	if len(o.BuilderArgs) > 1 {
		return fmt.Errorf("You must specify only one type of chaos")
	}

	switch strings.ToLower(o.BuilderArgs[0]) {
	case "networkchaos":
		if err := networkchaos.Debug(o.ChaosName, o.Namespace); err != nil {
			return err
		}
	case "stresschaos":
		if err := stresschaos.Debug(o.ChaosName, o.Namespace); err != nil {
			return err
		}
	default:
		return fmt.Errorf("chaos not supported")
	}
	return nil
}
