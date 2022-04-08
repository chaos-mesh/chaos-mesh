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
	"crypto"
	"crypto/x509"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type PhysicalMachineGenerateOptions struct {
	outputPath string
	caCertFile string
	caKeyFile  string
}

func NewPhysicalMachineGenerateCmd() (*cobra.Command, error) {
	generateOption := &PhysicalMachineGenerateOptions{}

	generateCmd := &cobra.Command{
		Use:           `generate`,
		Short:         `Generate TLS certs for certain physical machine`,
		Long:          `Generate TLS certs for certain physical machine (please execute this command on the certain physical machine)`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := generateOption.Validate(); err != nil {
				return err
			}
			return generateOption.Run()
		},
	}
	generateCmd.PersistentFlags().StringVar(&generateOption.outputPath, "path", "/etc/chaosd/pki", "path to save generated certs")
	generateCmd.PersistentFlags().StringVar(&generateOption.caCertFile, "cacert", "", "file path to cacert file")
	generateCmd.PersistentFlags().StringVar(&generateOption.caKeyFile, "cakey", "", "file path to cakey file")
	return generateCmd, nil
}

func (o *PhysicalMachineGenerateOptions) Validate() error {
	if len(o.caCertFile) == 0 {
		return errors.New("--cacert must be specified")
	}
	if len(o.caKeyFile) == 0 {
		return errors.New("--cakey must be specified")
	}
	return nil
}

func (o *PhysicalMachineGenerateOptions) Run() error {
	caCert, caKey, err := GetChaosdCAFileFromFile(o.caCertFile, o.caKeyFile)
	if err != nil {
		return err
	}

	serverCert, serverKey, err := NewCertAndKey(caCert, caKey)
	if err != nil {
		return err
	}

	return WriteCertAndKey(o.outputPath, ChaosdPkiName, serverCert, serverKey)
}

func GetChaosdCAFileFromFile(caCertFile, caKeyFile string) (*x509.Certificate, crypto.Signer, error) {
	certData, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot read cert file")
	}

	keyData, err := os.ReadFile(caKeyFile)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot read private key file")
	}
	return ParseCertAndKey(certData, keyData)
}
