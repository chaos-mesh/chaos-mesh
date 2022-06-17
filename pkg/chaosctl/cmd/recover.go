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
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/recover"
	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/label"
)

type RecoverOptions struct {
	namespace string
	labels    *[]string
}

func NewRecoverCommand(logger logr.Logger, builders map[string]recover.RecovererBuilder) (*cobra.Command, error) {
	o := &RecoverOptions{namespace: "default"}

	recoverCmd := &cobra.Command{
		Use:   `recover (CHAOSTYPE) POD[,POD[,POD...]] [-n NAMESPACE]`,
		Short: `Recover certain chaos from certain pods`,
		Long: `Recover certain chaos from certain pods.
Currently unimplemented.

Examples:
  # Recover network chaos from pods in namespace default
  chaosctl recover networkchaos

  # Recover network chaos from certain pods in certain namespace
  chaosctl recover networkchaos pod1 pod2 pod3 -n NAMESPACE
  
  # Recover network chaos from pods with label key=value
  chaosctl recover networkchaos -l key=value`,
		ValidArgsFunction: noCompletions,
	}

	for chaosType, builder := range builders {
		recoverCmd.AddCommand(recoverResourceCommand(o, chaosType, builder))
	}

	recoverCmd.PersistentFlags().StringVarP(&o.namespace, "namespace", "n", "default", "namespace to find pods")
	o.labels = recoverCmd.PersistentFlags().StringSliceP("label", "l", nil, "labels to select pods")
	err := recoverCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, cancel, err := common.CreateClient(context.TODO(), managerNamespace, managerSvc)
		if err != nil {
			logger.Error(err, "create client")
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		defer cancel()

		completion, err := client.ListNamespace(context.TODO())
		if err != nil {
			logger.Error(err, "complete resource")
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completion, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	})
	if err != nil {
		return nil, errors.Wrap(err, "register completion func for flag `namespace`")
	}
	err = recoverCmd.RegisterFlagCompletionFunc("label", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		return nil, errors.Wrap(err, "register completion func for flag `label`")
	}
	return recoverCmd, nil
}

func recoverResourceCommand(option *RecoverOptions, chaosType string, builder recover.RecovererBuilder) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf(`%s POD[,POD[,POD...]] [-n NAMESPACE]`, chaosType),
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
func (o *RecoverOptions) Run(recover recover.Recoverer, client *ctrlclient.CtrlClient, args []string) error {
	pods, err := o.selectPods(client, args)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, pod := range pods {
		err = recover.Recover(ctx, pod)
		if err != nil {
			return err
		}
	}
	return nil
}

// List pods to recover
func (o *RecoverOptions) List(client *ctrlclient.CtrlClient) ([]string, cobra.ShellCompDirective) {
	pods, err := o.selectPods(client, []string{})
	if err != nil {
		common.PrettyPrint(errors.Wrap(err, "select pods").Error(), 0, common.Red)
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var names []string
	for _, pod := range pods {
		names = append(names, pod.Name)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func (o *RecoverOptions) selectPods(client *ctrlclient.CtrlClient, names []string) ([]*recover.PartialPod, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	selector := v1alpha1.PodSelectorSpec{
		GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
			Namespaces: []string{o.namespace},
		},
	}

	if len(names) != 0 {
		selector.Pods = map[string][]string{o.namespace: names}
	}

	if o.labels != nil && len(*o.labels) > 0 {
		labels, err := label.ParseLabel(strings.Join(*o.labels, ","))
		if err != nil {
			return nil, errors.Wrap(err, "parse labels")
		}
		if len(labels) != 0 {
			selector.LabelSelectors = labels
		}
	}

	return recover.SelectPods(ctx, client, selector)
}
