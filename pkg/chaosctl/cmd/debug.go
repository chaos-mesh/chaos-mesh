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
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/iochaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/networkchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/stresschaos"
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

	// Need to separately support chaos-level completion, so split each chaos apart
	networkCmd := &cobra.Command{
		Use:   `networkchaos (CHAOSNAME) [-n NAMESPACE]`,
		Short: `Print the debug information for certain network chaos`,
		Long:  `Print the debug information for certain network chaos`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientset, err := common.InitClientSet()
			if err != nil {
				return err
			}
			return o.Run(networkChaos, args, clientset)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return validArgsChaos(networkChaos, o.namespace, args, toComplete)
		},
	}

	stressCmd := &cobra.Command{
		Use:   `stresschaos (CHAOSNAME) [-n NAMESPACE]`,
		Short: `Print the debug information for certain stress chaos`,
		Long:  `Print the debug information for certain stress chaos`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientset, err := common.InitClientSet()
			if err != nil {
				return err
			}
			return o.Run(stressChaos, args, clientset)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return validArgsChaos(stressChaos, o.namespace, args, toComplete)
		},
	}

	ioCmd := &cobra.Command{
		Use:   `iochaos (CHAOSNAME) [-n NAMESPACE]`,
		Short: `Print the debug information for certain io chaos`,
		Long:  `Print the debug information for certain io chaos`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientset, err := common.InitClientSet()
			if err != nil {
				return err
			}
			return o.Run(ioChaos, args, clientset)

		},
		SilenceErrors: true,
		SilenceUsage:  true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return validArgsChaos(ioChaos, o.namespace, args, toComplete)
		},
	}

	httpCmd := &cobra.Command{
		Use:   `httpchaos (CHAOSNAME) [-n NAMESPACE]`,
		Short: `Print the debug information for certain http chaos`,
		Long:  `Print the debug information for certain http chaos`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientset, err := common.InitClientSet()
			if err != nil {
				return err
			}
			return o.Run(httpChaos, args, clientset)

		},
		SilenceErrors: true,
		SilenceUsage:  true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return validArgsChaos(httpChaos, o.namespace, args, toComplete)
		},
	}

	debugCmd.AddCommand(networkCmd)
	debugCmd.AddCommand(stressCmd)
	debugCmd.AddCommand(ioCmd)
	debugCmd.AddCommand(httpCmd)

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

// Run debug
func (o *DebugOptions) Run(chaosType string, args []string, c *common.ClientSet) error {
	if len(args) > 1 {
		return fmt.Errorf("only one chaos could be specified")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chaosName := ""
	if len(args) == 1 {
		chaosName = args[0]
	}

	chaosList, chaosNameList, err := common.GetChaosList(ctx, chaosType, chaosName, o.namespace, c.CtrlCli)
	if err != nil {
		return err
	}
	var result []common.ChaosResult

	for i, chaos := range chaosList {
		var chaosResult common.ChaosResult
		chaosResult.Name = chaosNameList[i]

		var err error
		switch chaosType {
		case networkChaos:
			err = networkchaos.Debug(ctx, chaos, c, &chaosResult)
		case stressChaos:
			err = stresschaos.Debug(ctx, chaos, c, &chaosResult)
		case ioChaos:
			err = iochaos.Debug(ctx, chaos, c, &chaosResult)
		default:
			return fmt.Errorf("chaos type not supported")
		}
		result = append(result, chaosResult)
		if err != nil {
			common.PrintResult(result)
			return err
		}
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

	list, err := listChaos(context.TODO(), httpChaos, namespace, toComplete, client)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return list, cobra.ShellCompDirectiveDefault
}
