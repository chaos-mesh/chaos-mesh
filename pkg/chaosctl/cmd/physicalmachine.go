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
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
)

type PhysicalMachineInitOptions struct {
	logger             logr.Logger
	chaosMeshNamespace string
	remoteIP           string
	sshPort            int
	chaosdPort         int
	outputPath         string
	namespace          string
	labels             string
}

type PhysicalMachineGenerateOptions struct {
	logger     logr.Logger
	outputPath string
	caCertFile string
	caKeyFile  string
}

type PhysicalMachineCreateOptions struct {
	logger     logr.Logger
	namespace  string
	labels     string
	remoteIP   string
	chaosdPort int
}

func NewPhysicalMachineCommand(logger logr.Logger) (*cobra.Command, error) {
	initOption := &PhysicalMachineInitOptions{
		logger: logger,
	}
	generateOption := &PhysicalMachineGenerateOptions{
		logger: logger,
	}
	createOption := &PhysicalMachineCreateOptions{
		logger: logger,
	}

	physicalMachineCmd := &cobra.Command{
		Use:     `physical-machine (ACTION)`,
		Aliases: []string{"pm"},
		Short:   `Print the debug information for certain chaos`,
		Long: `Print the debug information for certain chaos.
Currently support networkchaos, stresschaos and iochaos.

Examples:
  # Return debug information from all networkchaos in default namespace
  chaosctl debug networkchaos

  # Return debug information from certain networkchaos
  chaosctl debug networkchaos CHAOSNAME -n NAMESPACE`,
		ValidArgsFunction: noCompletions,
	}

	initCmd := &cobra.Command{
		Use:           `init (PHYSICALMACHINE_NAME) [-n NAMESPACE]`,
		Short:         `Generate TLS certs for certain physical machine automatically, and create PhysicalMachine CustomResource in Kubernetes cluster`,
		Long:          `Generate TLS certs for certain physical machine automatically, and create PhysicalMachine CustomResource in Kubernetes cluster`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	initCmd.PersistentFlags().StringVar(&initOption.chaosMeshNamespace, "chaos-mesh-namespace", "chaos-testing", "namespace where chaos mesh installed")
	initCmd.PersistentFlags().StringVar(&initOption.remoteIP, "ip", "", "")
	initCmd.PersistentFlags().IntVar(&initOption.sshPort, "ssh-port", 22, "")
	initCmd.PersistentFlags().IntVar(&initOption.chaosdPort, "chaosd-port", 31768, "")
	initCmd.PersistentFlags().StringVar(&initOption.outputPath, "path", "/etc/chaosd/ssl", "")
	initCmd.PersistentFlags().StringVarP(&initOption.namespace, "namespace", "n", "default", "namespace where chaos mesh installed")
	initCmd.PersistentFlags().StringVarP(&initOption.labels, "labels", "l", "", "Selector (label query) to filter on.(e.g. -l key1=value1,key2=value2)")

	generateCmd := &cobra.Command{
		Use:           `generate`,
		Short:         `Generate TLS certs for certain physical machine`,
		Long:          `Generate TLS certs for certain physical machine (please execute this command on the certain physical machine)`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	generateCmd.PersistentFlags().StringVar(&generateOption.outputPath, "path", "/etc/chaosd/ssl", "")
	generateCmd.PersistentFlags().StringVar(&generateOption.caCertFile, "cacert", "", "")
	generateCmd.PersistentFlags().StringVar(&generateOption.caKeyFile, "cakey", "", "")

	createCmd := &cobra.Command{
		Use:           `create (PHYSICALMACHINE_NAME) [-n NAMESPACE]`,
		Short:         `Create PhysicalMachine CustomResource in Kubernetes cluster`,
		Long:          `Create PhysicalMachine CustomResource in Kubernetes cluster`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	createCmd.PersistentFlags().StringVarP(&createOption.namespace, "namespace", "n", "default", "namespace where chaos mesh installed")
	createCmd.PersistentFlags().StringVarP(&createOption.labels, "labels", "l", "", "Selector (label query) to filter on.(e.g. -l key1=value1,key2=value2)")
	createCmd.PersistentFlags().StringVar(&createOption.remoteIP, "ip", "", "")
	createCmd.PersistentFlags().IntVar(&createOption.chaosdPort, "chaosd-port", 31768, "")

	physicalMachineCmd.AddCommand(initCmd)
	physicalMachineCmd.AddCommand(generateCmd)
	physicalMachineCmd.AddCommand(createCmd)

	return physicalMachineCmd, nil
}
