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

package physicalmachine

import (
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
)

type PhysicalMachineCreateOptions struct {
	logger     logr.Logger
	namespace  string
	labels     string
	remoteIP   string
	chaosdPort int
	secure     bool
}

func NewPhysicalMachineCreateCmd(logger logr.Logger) (*cobra.Command, error) {
	createOption := &PhysicalMachineCreateOptions{
		logger: logger,
	}

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
	createCmd.PersistentFlags().StringVarP(&createOption.namespace, "namespace", "n", "default", "namespace of the certain physical machine")
	createCmd.PersistentFlags().StringVarP(&createOption.labels, "labels", "l", "", "labels of the certain physical machine (e.g. -l key1=value1,key2=value2)")
	createCmd.PersistentFlags().StringVar(&createOption.remoteIP, "ip", "", "ip of the remote physical machine")
	createCmd.PersistentFlags().IntVar(&createOption.chaosdPort, "chaosd-port", 31768, "port of the remote chaosd server listen")
	createCmd.PersistentFlags().BoolVar(&createOption.secure, "secure", true, "if true, represent that the remote chaosd serve HTTPS")

	return createCmd, nil
}
