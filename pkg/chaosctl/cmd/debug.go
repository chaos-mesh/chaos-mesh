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
	"log"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cm "github.com/chaos-mesh/chaos-mesh/pkg/debug/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/debug/iochaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/debug/networkchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/debug/stresschaos"
)

type DebugOptions struct {
	ChaosName string
	Namespace string
}

func init() {
	o := &DebugOptions{}

	c, err := cm.InitClientSet()
	if err != nil {
		log.Fatal(err)
	}

	// debugCmd represents the debug command
	debugCmd := &cobra.Command{
		Use:   `debug (CHAOSTYPE) [-c CHAOSNAME] [-n NAMESPACE]`,
		Short: `Print the debug information for certain chaos`,
		Long: `Print the debug information for certain chaos.
Currently only support networkchaos and stresschaos.

Examples:
  # Return debug information from all networkchaos in default namespace
  chaosctl debug networkchaos

  # Return debug information from certain networkchaos
  chaosctl debug networkchaos -c web-show-network-delay -n chaos-testing`,
		ValidArgsFunction: noCompletions,
	}

	networkCmd := &cobra.Command{
		Use:   `networkchaos (CHAOSNAME) [-n NAMESPACE]`,
		Short: `Print the debug information for certain network chaos`,
		Long:  `Print the debug information for certain network chaos`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Run("networkchaos", args, c); err != nil {
				log.Fatal(err)
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return listChaos("networkchaos", o.Namespace, toComplete, c.CtrlClient)
		},
	}

	stressCmd := &cobra.Command{
		Use:   `stresschaos (CHAOSNAME) [-n NAMESPACE]`,
		Short: `Print the debug information for certain stress chaos`,
		Long:  `Print the debug information for certain stress chaos`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Run("stresschaos", args, c); err != nil {
				log.Fatal(err)
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return listChaos("stresschaos", o.Namespace, toComplete, c.CtrlClient)
		},
	}

	ioCmd := &cobra.Command{
		Use:   `iochaos (CHAOSNAME) [-n NAMESPACE]`,
		Short: `Print the debug information for certain io chaos`,
		Long:  `Print the debug information for certain io chaos`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Run("iochaos", args, c); err != nil {
				log.Fatal(err)
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return listChaos("iochaos", o.Namespace, toComplete, c.CtrlClient)
		},
	}

	debugCmd.AddCommand(networkCmd)
	debugCmd.AddCommand(stressCmd)
	debugCmd.AddCommand(ioCmd)

	debugCmd.PersistentFlags().StringVarP(&o.Namespace, "namespace", "n", "default", "namespace to find chaos")
	err = debugCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return listNamespace(toComplete, c.K8sClient)
	})
	if err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(debugCmd)
}

func (o *DebugOptions) Run(chaosType string, args []string, c *cm.ClientSet) error {
	if len(args) > 1 {
		return fmt.Errorf("Only one chaos could be specified")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chaosNameArg := ""
	if len(args) == 1 {
		chaosNameArg = args[0]
	}

	chaosList, err := cm.GetChaosList(ctx, chaosType, chaosNameArg, o.Namespace, c.CtrlClient)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	for _, chaosName := range chaosList {
		cm.Print("[CHAOSNAME]: "+chaosName, 0, cm.ColorBlue)
		chaos, err := cm.GetChaos(ctx, chaosType, chaosName, o.Namespace, c.CtrlClient)
		if err != nil {
			return fmt.Errorf("failed to get chaos %s: %s", chaosName, err.Error())
		}

		switch chaosType {
		case "networkchaos":
			if err := networkchaos.Debug(ctx, chaos, c); err != nil {
				return err
			}
		case "stresschaos":
			if err := stresschaos.Debug(ctx, chaos, c); err != nil {
				return err
			}
		case "iochaos":
			if err := iochaos.Debug(ctx, chaos, c); err != nil {
				return err
			}
		default:
			return fmt.Errorf("chaos not supported")
		}
	}

	return nil
}

func listNamespace(toComplete string, c *kubernetes.Clientset) ([]string, cobra.ShellCompDirective) {
	namespaces, err := c.CoreV1().Namespaces().List(v1.ListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	var ret []string
	for _, ns := range namespaces.Items {
		if strings.HasPrefix(ns.Name, toComplete) {
			ret = append(ret, ns.Name)
		}
	}
	return ret, cobra.ShellCompDirectiveNoFileComp
}

func listChaos(chaosType string, namespace string, toComplete string, c client.Client) ([]string, cobra.ShellCompDirective) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	chaosList, err := cm.GetChaosList(ctx, chaosType, "", namespace, c)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	var ret []string
	for _, chaos := range chaosList {
		if strings.HasPrefix(chaos, toComplete) {
			ret = append(ret, chaos)
		}
	}
	return ret, cobra.ShellCompDirectiveNoFileComp
}
