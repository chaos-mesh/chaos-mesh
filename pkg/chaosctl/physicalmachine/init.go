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
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/keyutil"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/label"
)

type PhysicalMachineInitOptions struct {
	chaosMeshNamespace string
	remoteIP           string
	sshUser            string
	sshPort            int
	sshPrivateKeyFile  string
	chaosdPort         int
	outputPath         string
	namespace          string
	labels             string
}

func NewPhysicalMachineInitCmd() (*cobra.Command, error) {
	initOption := &PhysicalMachineInitOptions{}

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
			if err := initOption.Validate(); err != nil {
				return err
			}
			return initOption.Run(args)
		},
	}
	initCmd.PersistentFlags().StringVar(&initOption.chaosMeshNamespace, "chaos-mesh-namespace", "chaos-mesh", "namespace where chaos mesh installed")
	initCmd.PersistentFlags().StringVar(&initOption.remoteIP, "ip", "", "ip of the remote physical machine")
	initCmd.PersistentFlags().StringVar(&initOption.sshUser, "ssh-user", "root", "username for ssh connection")
	initCmd.PersistentFlags().StringVar(&initOption.sshPrivateKeyFile, "ssh-key", filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "private key filepath for ssh connection")
	initCmd.PersistentFlags().IntVar(&initOption.sshPort, "ssh-port", 22, "port of ssh connection")
	initCmd.PersistentFlags().IntVar(&initOption.chaosdPort, "chaosd-port", 31768, "port of the remote chaosd server listen")
	initCmd.PersistentFlags().StringVar(&initOption.outputPath, "path", "/etc/chaosd/pki", "path to save generated certs")
	initCmd.PersistentFlags().StringVarP(&initOption.namespace, "namespace", "n", "default", "namespace of the certain physical machine")
	initCmd.PersistentFlags().StringVarP(&initOption.labels, "labels", "l", "", "labels of the certain physical machine (e.g. -l key1=value1,key2=value2)")
	return initCmd, nil
}

func (o *PhysicalMachineInitOptions) Validate() error {
	if len(o.remoteIP) == 0 {
		return errors.New("--ip must be specified")
	}
	return nil
}

func (o *PhysicalMachineInitOptions) Run(args []string) error {
	if len(args) < 1 {
		return errors.New("physical machine name is required")
	}
	physicalMachineName := args[0]

	labels, err := label.ParseLabel(o.labels)
	if err != nil {
		return err
	}
	address := formatAddress(o.remoteIP, o.chaosdPort, true)

	clientset, err := common.InitClientSet()
	if err != nil {
		return err
	}

	ctx := context.Background()
	caCert, caKey, err := GetChaosdCAFileFromCluster(ctx, o.chaosMeshNamespace, clientset.CtrlCli)
	if err != nil {
		return err
	}

	// generate chaosd cert and private key
	serverCert, serverKey, err := NewCertAndKey(caCert, caKey)
	if err != nil {
		return err
	}

	sshTunnel, err := NewSshTunnel(o.remoteIP, strconv.Itoa(o.sshPort), o.sshUser, o.sshPrivateKeyFile)
	if err != nil {
		return err
	}
	if err := sshTunnel.Open(); err != nil {
		return err
	}
	defer sshTunnel.Close()

	if err := writeCertAndKeyToRemote(sshTunnel, o.outputPath, ChaosdPkiName, serverCert, serverKey); err != nil {
		return err
	}
	if err := writeCertToRemote(sshTunnel, o.outputPath, "ca", caCert); err != nil {
		return err
	}

	return CreatePhysicalMachine(ctx, clientset.CtrlCli, o.namespace, physicalMachineName, address, labels)
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
	var ok bool
	if caCertBytes, ok = secret.Data["ca.crt"]; !ok {
		return nil, nil, errors.New("could not found ca cert file in `chaos-mesh-chaosd-client-certs` secret")
	}

	var caKeyBytes []byte
	if caKeyBytes, ok = secret.Data["ca.key"]; !ok {
		return nil, nil, errors.New("could not found ca key file in `chaos-mesh-chaosd-client-certs` secret")
	}

	return ParseCertAndKey(caCertBytes, caKeyBytes)
}

func writeCertAndKeyToRemote(sshTunnel *SshTunnel, pkiPath, pkiName string, cert *x509.Certificate, key crypto.Signer) error {
	keyBytes, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return err
	}
	if err := writeCertToRemote(sshTunnel, pkiPath, pkiName, cert); err != nil {
		return err
	}
	return sshTunnel.SFTP(pathForKey(pkiPath, pkiName), keyBytes)
}

func writeCertToRemote(sshTunnel *SshTunnel, pkiPath, pkiName string, cert *x509.Certificate) error {
	return sshTunnel.SFTP(pathForCert(pkiPath, pkiName), EncodeCertPEM(cert))
}
