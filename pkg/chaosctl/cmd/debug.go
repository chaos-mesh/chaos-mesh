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
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/iochaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/networkchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/debug/stresschaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/grpc"
)

type DebugOptions struct {
	logger     logr.Logger
	namespace  string
	CaCertFile string
	CertFile   string
	KeyFile    string
	Insecure   bool
}

const (
	networkChaos = "networkchaos"
	stressChaos  = "stresschaos"
	ioChaos      = "iochaos"
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
			clientset, err := common.InitClientSet()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return listChaos(networkChaos, o.namespace, toComplete, clientset.CtrlCli)
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
			clientset, err := common.InitClientSet()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return listChaos(stressChaos, o.namespace, toComplete, clientset.CtrlCli)
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
			clientset, err := common.InitClientSet()
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return listChaos(ioChaos, o.namespace, toComplete, clientset.CtrlCli)
		},
	}

	debugCmd.AddCommand(networkCmd)
	debugCmd.AddCommand(stressCmd)
	debugCmd.AddCommand(ioCmd)

	debugCmd.PersistentFlags().StringVarP(&o.namespace, "namespace", "n", "default", "namespace to find chaos")
	debugCmd.PersistentFlags().StringVar(&o.CaCertFile, "cacert", "", "file path to cacert file")
	debugCmd.PersistentFlags().StringVar(&o.CertFile, "cert", "", "file path to cert file")
	debugCmd.PersistentFlags().StringVar(&o.KeyFile, "key", "", "file path to key file")
	debugCmd.PersistentFlags().BoolVarP(&o.Insecure, "insecure", "i", false, "Insecure mode will use unauthorized grpc")
	err := debugCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		clientset, err := common.InitClientSet()
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}
		return listNamespace(toComplete, clientset.KubeCli)
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
	common.TLSFiles = grpc.TLSFile{CaCert: o.CaCertFile, Cert: o.CertFile, Key: o.KeyFile}
	common.Insecure = o.Insecure

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
	_, chaosList, err := common.GetChaosList(ctx, chaosType, "", namespace, c)
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
