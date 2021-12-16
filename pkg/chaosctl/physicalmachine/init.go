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
	"crypto"
	"crypto/x509"
	"encoding/base64"

	"github.com/pkg/errors"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PhysicalMachineInitOptions struct {
	logger             logr.Logger
	chaosMeshNamespace string
	remoteIP           string
	sshUser            string
	sshPort            int
	chaosdPort         int
	outputPath         string
	namespace          string
	labels             string
}

func NewPhysicalMachineInitCmd(logger logr.Logger) (*cobra.Command, error) {
	initOption := &PhysicalMachineInitOptions{
		logger: logger,
	}

	initCmd := &cobra.Command{
		Use:   `init (PHYSICALMACHINE_NAME) [-n NAMESPACE]`,
		Short: `Generate TLS certs for certain physical machine automatically, and create PhysicalMachine CustomResource in Kubernetes cluster`,
		Long: `Generate TLS certs for certain physical machine automatically, and create PhysicalMachine CustomResource in Kubernetes cluster

Examples:
  # Generate TLS certs for remote physical machine, and create PhysicalMachine CustomResource in certain namespace
  chaosctl pm init PHYSICALMACHINE_NAME -n NAMESPACE --ip REMOTEIP
  
  # Generate TLS certs for remote physical machine, and create PhysicalMachine CustomResource in certain namespace with specified labels
  chaosctl pm init PHYSICALMACHINE_NAME -n NAMESPACE --ip REMOTEIP -l key1=value1,key2=value2
  `,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	initCmd.PersistentFlags().StringVar(&initOption.chaosMeshNamespace, "chaos-mesh-namespace", "chaos-testing", "namespace where chaos mesh installed")
	initCmd.PersistentFlags().StringVar(&initOption.remoteIP, "ip", "", "ip of the remote physical machine")
	initCmd.PersistentFlags().StringVar(&initOption.sshUser, "ssh-user", "root", "username for ssh connection")
	initCmd.PersistentFlags().IntVar(&initOption.sshPort, "ssh-port", 22, "port of ssh connection")
	initCmd.PersistentFlags().IntVar(&initOption.chaosdPort, "chaosd-port", 31768, "port of the remote chaosd server listen")
	initCmd.PersistentFlags().StringVar(&initOption.outputPath, "path", "/etc/chaosd/ssl", "path to save generated certs")
	initCmd.PersistentFlags().StringVarP(&initOption.namespace, "namespace", "n", "default", "namespace of the certain physical machine")
	initCmd.PersistentFlags().StringVarP(&initOption.labels, "labels", "l", "", "labels of the certain physical machine (e.g. -l key1=value1,key2=value2)")
	return initCmd, nil
}

func (o *PhysicalMachineInitOptions) Run() error {
	// clientset, err := common.InitClientSet()
	// if err != nil {
	// 	return err
	// }

	return nil
}

func SSH() {

}

func GetChaosdCAFileFromCluster(ctx context.Context, namespace string, c client.Client) (caCert *x509.Certificate, caKey crypto.Signer, err error) {
	var secret v1.Secret
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      "chaos-mesh-chaosd-client-certs",
	}, &secret); err != nil {
		return nil, nil, errors.Wrapf(err, "could not found secret `chaos-mesh-chaosd-client-certs` in namespace %s", namespace)
	}

	var caCertBytes []byte
	if caCertRaw, ok := secret.Data["ca.crt"]; !ok {
		return nil, nil, errors.New("could not found ca cert file in `chaos-mesh-chaosd-client-certs` secret")
	} else {
		if _, err = base64.StdEncoding.Decode(caCertBytes, caCertRaw); err != nil {
			return nil, nil, errors.Wrap(err, "decode ca cert file failed")
		}
	}
	caCert, err = ParseCert(caCertBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parse certs pem failed")
	}

	var caKeyBytes []byte
	if caKeyRaw, ok := secret.Data["ca.key"]; !ok {
		return nil, nil, errors.New("could not found ca key file in `chaos-mesh-chaosd-client-certs` secret")
	} else {
		if _, err = base64.StdEncoding.Decode(caKeyBytes, caKeyRaw); err != nil {
			return nil, nil, errors.Wrap(err, "decode ca key file failed")
		}
	}
	caKey, err = ParsePrivateKey(caKeyBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parse ca key file failed")
	}

	return caCert, caKey, nil
}
