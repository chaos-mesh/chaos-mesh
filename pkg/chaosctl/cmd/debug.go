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
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/httpchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/iochaos"
	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

type DebugOptions struct {
	logger    logr.Logger
	namespace string
}

const (
	networkChaos = "networkchaos"
	stressChaos  = "stresschaos"
	ioChaos      = "iochaos"
	httpChaos    = "httpchaos"
)

func NewDebugCommand(logger logr.Logger) (*cobra.Command, error) {
	o := &DebugOptions{
		logger: logger,
	}

	debugCmd := &cobra.Command{
		Use:   `debug (CHAOSTYPE) [-c CHAOSNAME] [-n NAMESPACE]`,
		Short: `Print the debug information for certain chaos`,
		Long: `Print the debug information for certain chaos.
Currently support networkchaos, stresschaos and iochaos.

Examples:
  # Return debug information from all networkchaos in default namespace
  chaosctl debug networkchaos

  # Return debug information from certain networkchaos
  chaosctl debug networkchaos CHAOSNAME -n NAMESPACE`,
		ValidArgsFunction: noCompletions,
	}

	debugCmd.AddCommand(debugResouceCommand(o, networkChaos))
	debugCmd.AddCommand(debugResouceCommand(o, stressChaos))
	debugCmd.AddCommand(debugResouceCommand(o, ioChaos))
	debugCmd.AddCommand(debugResouceCommand(o, httpChaos))

	debugCmd.PersistentFlags().StringVarP(&o.namespace, "namespace", "n", "default", "namespace to find chaos")
	err := debugCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, cancel, err := createClient(context.TODO(), managerNamespace)
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

func debugResouceCommand(option *DebugOptions, chaosType string) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf(`%s (CHAOSNAME) [-n NAMESPACE]`, chaosType),
		Short: fmt.Sprintf(`Print the debug information for certain %s`, chaosType),
		Long:  fmt.Sprintf(`Print the debug information for certain %s`, chaosType),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrlclient.DisableRuntimeErrorHandler()
			client, cancel, err := createClient(context.TODO(), managerNamespace)
			if err != nil {
				return err
			}
			defer cancel()
			return option.Run(chaosType, args, client)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return validArgsChaos(chaosType, option.namespace, args, toComplete)
		},
	}
}

// Run debug
func (o *DebugOptions) Run(chaosType string, args []string, client *ctrlclient.CtrlClient) error {
	if len(args) > 1 {
		return fmt.Errorf("only one chaos could be specified")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chaosName := ""
	if len(args) == 1 {
		chaosName = args[0]
	}

	var result []*common.ChaosResult
	var err error

	switch chaosType {
	case ioChaos:
		result, err = iochaos.Debug(ctx, o.namespace, chaosName, client)
	case httpChaos:
		result, err = httpchaos.Debug(ctx, o.namespace, chaosName, client)
	default:
		err = fmt.Errorf("chaos type not supported")
	}

	if err != nil {
		return err
	}

	common.PrintResult(result)
	return nil
}

func listChaos(ctx context.Context, chaosType, namespace, toComplete string, c *ctrlclient.CtrlClient) ([]string, error) {
	return c.ListArguments(ctx, []string{fmt.Sprintf("%s:%s", ctrlclient.NamespaceKey, namespace), chaosType}, "name", toComplete)
}

func validArgsChaos(chaosType, namespace string, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, cancel, err := createClient(context.TODO(), managerNamespace)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer cancel()

	list, err := listChaos(context.TODO(), chaosType, namespace, toComplete, client)
	if err != nil || len(list) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return list, cobra.ShellCompDirectiveDefault
}
