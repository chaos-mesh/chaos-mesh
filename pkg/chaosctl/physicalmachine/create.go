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
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/label"
)

type PhysicalMachineCreateOptions struct {
	namespace  string
	labels     string
	remoteIP   string
	chaosdPort int
	secure     bool
}

func NewPhysicalMachineCreateCmd() (*cobra.Command, error) {
	createOption := &PhysicalMachineCreateOptions{}

	createCmd := &cobra.Command{
		Use:           `create (PHYSICALMACHINE_NAME) [-n NAMESPACE]`,
		Short:         `Create PhysicalMachine CustomResource in Kubernetes cluster`,
		Long:          `Create PhysicalMachine CustomResource in Kubernetes cluster`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := createOption.Validate(); err != nil {
				return err
			}
			return createOption.Run(args)
		},
	}
	createCmd.PersistentFlags().StringVarP(&createOption.namespace, "namespace", "n", "default", "namespace of the certain physical machine")
	createCmd.PersistentFlags().StringVarP(&createOption.labels, "labels", "l", "", "labels of the certain physical machine (e.g. -l key1=value1,key2=value2)")
	createCmd.PersistentFlags().StringVar(&createOption.remoteIP, "ip", "", "ip of the remote physical machine")
	createCmd.PersistentFlags().IntVar(&createOption.chaosdPort, "chaosd-port", 31768, "port of the remote chaosd server listen")
	createCmd.PersistentFlags().BoolVar(&createOption.secure, "secure", true, "if true, represent that the remote chaosd serve HTTPS")

	return createCmd, nil
}

func (o *PhysicalMachineCreateOptions) Validate() error {
	if len(o.remoteIP) == 0 {
		return errors.New("--ip must be specified")
	}
	return nil
}

func (o *PhysicalMachineCreateOptions) Run(args []string) error {
	if len(args) < 1 {
		return errors.New("physical machine name is required")
	}
	physicalMachineName := args[0]

	labels, err := label.ParseLabel(o.labels)
	if err != nil {
		return err
	}
	address := formatAddress(o.remoteIP, o.chaosdPort, o.secure)

	clientset, err := common.InitClientSet()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return CreatePhysicalMachine(ctx, clientset.CtrlCli, o.namespace, physicalMachineName, address, labels)
}

func CreatePhysicalMachine(ctx context.Context, c client.Client,
	namespace, name, address string, labels map[string]string) error {
	pm := v1alpha1.PhysicalMachine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: v1alpha1.PhysicalMachineSpec{
			Address: address,
		},
	}

	return c.Create(ctx, &pm)
}

func formatAddress(ip string, port int, secure bool) string {
	protocol := "http"
	if secure {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s:%d", protocol, ip, port)
}
