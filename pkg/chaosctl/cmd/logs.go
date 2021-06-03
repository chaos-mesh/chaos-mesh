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
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

type logsOptions struct {
	logger logr.Logger
	tail   int64
	node   string
}

func NewLogsCmd(logger logr.Logger) (*cobra.Command, error) {
	o := &logsOptions{
		logger: logger,
	}

	logsCmd := &cobra.Command{
		Use:   `logs [-t LINE]`,
		Short: `Print logs of controller-manager, chaos-daemon and chaos-dashboard`,
		Long: `Print logs of controller-manager, chaos-daemon and chaos-dashboard, to provide debug information.

Examples:
  # Default print all log of all chaosmesh components
  chaosctl logs

  # Print 100 log lines for chaosmesh components in node NODENAME
  chaosctl logs -t 100 -n NODENAME`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run(args)
		},
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: noCompletions,
	}

	logsCmd.Flags().Int64VarP(&o.tail, "tail", "t", -1, "Number of lines of recent log")
	logsCmd.Flags().StringVarP(&o.node, "node", "n", "", "Number of lines of recent log")
	err := logsCmd.RegisterFlagCompletionFunc("node", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		clientset, err := cm.InitClientSet()
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}
		return listNodes(toComplete, clientset.KubeCli)
	})
	if err != nil {
		return nil, err
	}
	return logsCmd, nil
}

// Run logs
func (o *logsOptions) Run(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, err := cm.InitClientSet()
	if err != nil {
		o.logger.V(4).Info("failed to initialize clientset", "err", err)
		return err
	}

	componentsNeeded := []string{"controller-manager", "chaos-daemon", "chaos-dashboard"}
	for _, name := range componentsNeeded {
		selectorSpec := v1alpha1.PodSelectorSpec{
			LabelSelectors: map[string]string{"app.kubernetes.io/component": name},
		}
		if o.node != "" {
			selectorSpec.Nodes = []string{o.node}
		}

		// TODO: just use kubernetes native label selector
		components, err := pod.SelectPods(ctx, c.CtrlCli, nil, selectorSpec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, false)
		if err != nil {
			return errors.Wrapf(err, "failed to SelectPods for component %s", name)
		}
		o.logger.V(4).Info("select pods for component", "component", name, "pods", components)
		for _, comp := range components {
			cm.PrettyPrint(fmt.Sprintf("[%s]", comp.Name), 0, cm.Cyan)
			comLog, err := cm.Log(comp, o.tail, c.KubeCli)
			if err != nil {
				cm.PrettyPrint(err.Error(), 1, cm.Red)
			} else {
				cm.PrettyPrint(comLog, 1, cm.NoColor)
			}
		}
	}
	return nil
}

func listNodes(toComplete string, c *kubernetes.Clientset) ([]string, cobra.ShellCompDirective) {
	nodes, err := c.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	var ret []string
	for _, ns := range nodes.Items {
		if strings.HasPrefix(ns.Name, toComplete) {
			ret = append(ret, ns.Name)
		}
	}
	return ret, cobra.ShellCompDirectiveNoFileComp
}
