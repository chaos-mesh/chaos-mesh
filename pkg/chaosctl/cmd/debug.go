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
	validArgs := []string{"networkchaos", "stresschaos", "iochaos"}

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
  # Return debug information from all networkchaos
  chaosctl debug networkchaos

  # Return debug information from certain networkchaos
  chaosctl debug networkchaos -c web-show-network-delay -n chaos-testing`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Run(args, c); err != nil {
				log.Fatal(err)
			}
		},
		ValidArgs: validArgs,
	}

	debugCmd.Flags().StringVarP(&o.ChaosName, "chaosname", "c", "", "certain chaos name")
	debugCmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "default", "namespace to find chaos")

	if err := flagCompletion(debugCmd, c.K8sClient); err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(debugCmd)
}

func (o *DebugOptions) Run(args []string, c *cm.ClientSet) error {
	if len(args) == 0 {
		return fmt.Errorf("The type of chaos need to be specified")
	}
	if len(args) > 1 {
		return fmt.Errorf("Only one type of chaos need to be specified")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chaosType := strings.ToLower(args[0])
	chaosList, err := cm.GetChaosList(ctx, chaosType, o.ChaosName, o.Namespace, c.CtrlClient)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	for _, chaosName := range chaosList {
		fmt.Println(string(cm.ColorCyan), "[CHAOSNAME]:", chaosName, string(cm.ColorReset))

		switch chaosType {
		case "networkchaos":
			if err := networkchaos.Debug(ctx, chaosName, o.Namespace, c); err != nil {
				return err
			}
		case "stresschaos":
			if err := stresschaos.Debug(ctx, chaosName, o.Namespace, c); err != nil {
				return err
			}
		case "iochaos":
			if err := iochaos.Debug(ctx, chaosName, o.Namespace, c); err != nil {
				return err
			}
		default:
			return fmt.Errorf("chaos not supported")
		}
	}

	return nil
}

func flagCompletion(cmd *cobra.Command, c *kubernetes.Clientset) error {
	err := cmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return listNameSpace(toComplete, c)
	})
	if err != nil {
		return err
	}
	return nil
}

func listNameSpace(toComplete string, c *kubernetes.Clientset) ([]string, cobra.ShellCompDirective) {
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
