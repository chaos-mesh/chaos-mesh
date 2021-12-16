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

type PhysicalMachineGenerateOptions struct {
	logger     logr.Logger
	outputPath string
	caCertFile string
	caKeyFile  string
}

func NewPhysicalMachineGenerateCmd(logger logr.Logger) (*cobra.Command, error) {
	generateOption := &PhysicalMachineGenerateOptions{
		logger: logger,
	}

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
	generateCmd.PersistentFlags().StringVar(&generateOption.outputPath, "path", "/etc/chaosd/ssl", "path to save generated certs")
	generateCmd.PersistentFlags().StringVar(&generateOption.caCertFile, "cacert", "", "file path to cacert file")
	generateCmd.PersistentFlags().StringVar(&generateOption.caKeyFile, "cakey", "", "file path to cakey file")
	return generateCmd, nil
}
