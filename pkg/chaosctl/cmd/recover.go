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

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/recover"
	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

type RecoverOptions struct {
	namespace string
}

func NewRecoverCommand(logger logr.Logger, builders map[string]recover.RecoverBuilder) (*cobra.Command, error) {
	o := &RecoverOptions{namespace: "default"}

	recoverCmd := &cobra.Command{
		Use:   `recover (CHAOSTYPE) (PODs) [-n NAMESPACE]`,
		Short: `Recover certain chaos from certain pods`,
		Long: `Recover certain chaos from certain pods.
Currently support networkchaos, stresschaos, iochaos and httpchaos.

Examples:
  # Recover network chaos from pods in namespace default
  chaosctl debug networkchaos

  # Recover network chaos from certain pod in 
  chaosctl debug networkchaos PODs -n NAMESPACE`,
		ValidArgsFunction: noCompletions,
	}

	for chaosType, builder := range builders {
		recoverCmd.AddCommand(recoverResourceCommand(o, chaosType, builder))
	}

	recoverCmd.PersistentFlags().StringVarP(&o.namespace, "namespace", "n", "default", "namespace to find chaos")
	err := recoverCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	return recoverCmd, err
}

func recoverResourceCommand(option *RecoverOptions, chaosType string, builder recover.RecoverBuilder) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf(`%s (PODs) [-n NAMESPACE]`, chaosType),
		Short: fmt.Sprintf(`Recover %s from certain pods`, chaosType),
		Long:  fmt.Sprintf(`Recover %s from certain pods`, chaosType),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, cancel, err := common.CreateClient(context.TODO(), managerNamespace, managerSvc)
			if err != nil {
				return err
			}
			defer cancel()
			return option.Run(builder(client), client, args)
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
			return option.List(client)
		},
	}
}

// Run recover
func (o *RecoverOptions) Run(recover recover.Recover, client *ctrlclient.CtrlClient, args []string) error {
	var names []string
	var err error
	if len(args) > 0 {
		names = args
	} else {
		names, err = o.selectPods(client)
		if err != nil {
			return errors.Wrap(err, "select pods")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, name := range names {
		err = recover.Recover(ctx, o.namespace, name)
		if err != nil {
			return err
		}
	}
	return nil
}

// List pods to recover
func (o *RecoverOptions) List(client *ctrlclient.CtrlClient) ([]string, cobra.ShellCompDirective) {
	names, err := o.selectPods(client)
	if err != nil {
		common.PrettyPrint(errors.Wrap(err, "select pods").Error(), 0, common.Red)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func (o *RecoverOptions) selectPods(client *ctrlclient.CtrlClient) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	namespacedNames, err := client.SelectPods(ctx, v1alpha1.PodSelectorSpec{
		GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
			Namespaces: []string{o.namespace},
		},
	})

	if err != nil {
		return nil, err
	}

	var names []string
	for _, namespacedName := range namespacedNames {
		names = append(names, namespacedName.Name)
	}

	return names, nil
}
