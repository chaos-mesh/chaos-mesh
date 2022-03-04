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
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug"
)

type DebugOptions struct {
	namespace string
}

const (
	networkChaos = "networkchaos"
	stressChaos  = "stresschaos"
	ioChaos      = "iochaos"
	httpChaos    = "httpchaos"
)

func NewDebugCommand(logger logr.Logger, debugs map[string]debug.Debug) (*cobra.Command, error) {
	o := &DebugOptions{}

	debugCmd := &cobra.Command{
		Use:   `debug (CHAOSTYPE) [-c CHAOSNAME] [-n NAMESPACE]`,
		Short: `Print the debug information for certain chaos`,
		Long: `Print the debug information for certain chaos.
Currently support networkchaos, stresschaos, iochaos and httpchaos.

Examples:
  # Return debug information from all networkchaos in default namespace
  chaosctl debug networkchaos

  # Return debug information from certain networkchaos
  chaosctl debug networkchaos CHAOSNAME -n NAMESPACE`,
		ValidArgsFunction: noCompletions,
	}

	for chaosType, debug := range debugs {
		debugCmd.AddCommand(debugResourceCommand(o, chaosType, debug))
	}

	debugCmd.PersistentFlags().StringVarP(&o.namespace, "namespace", "n", "default", "namespace to find chaos")
	err := debugCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, cancel, err := common.CreateClient(context.TODO(), managerNamespace, managerSvc)
		if err != nil {
			logger.Error(err, "fail to create client")
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		defer cancel()

		completion, err := client.ListNamespace(context.TODO())
		if err != nil {
			logger.Error(err, "fail to complete resource")
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completion, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	})
	return debugCmd, err
}

func debugResourceCommand(option *DebugOptions, chaosType string, debug debug.Debug) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf(`%s (CHAOSNAME) [-n NAMESPACE]`, chaosType),
		Short: fmt.Sprintf(`Print the debug information for certain %s`, chaosType),
		Long:  fmt.Sprintf(`Print the debug information for certain %s`, chaosType),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, cancel, err := common.CreateClient(context.TODO(), managerNamespace, managerSvc)
			if err != nil {
				return err
			}
			defer cancel()
			return option.Run(debug(client), args)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}
			client, cancel, err := common.CreateClient(context.TODO(), managerNamespace, managerSvc)
			if err != nil {
				common.PrettyPrint(errors.Wrap(err, "create client").Error(), 0, common.Red)
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			defer cancel()
			return option.List(debug(client))
		},
	}
}

// Run debug
func (o *DebugOptions) Run(debugger debug.Debugger, args []string) error {
	if len(args) > 1 {
		return errors.New("only one chaos could be specified")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chaosName := ""
	if len(args) == 1 {
		chaosName = args[0]
	}

	var result []*common.ChaosResult
	var err error

	result, err = debugger.Collect(ctx, o.namespace, chaosName)
	if err != nil {
		return err
	}

	common.PrintResult(result)
	return nil
}

// Run debug
func (o *DebugOptions) List(debugger debug.Debugger) ([]string, cobra.ShellCompDirective) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chaos, err := debugger.List(ctx, o.namespace)
	if err != nil {
		common.PrettyPrint(errors.Wrap(err, "list chaos").Error(), 0, common.Red)
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return chaos, cobra.ShellCompDirectiveNoFileComp
}
