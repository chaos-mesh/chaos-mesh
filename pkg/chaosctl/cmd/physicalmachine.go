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
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/physicalmachine"
)

func NewPhysicalMachineCommand() (*cobra.Command, error) {
	physicalMachineCmd := &cobra.Command{
		Use:               `physical-machine (ACTION)`,
		Aliases:           []string{"pm"},
		Short:             `Helper for generating TLS certs and creating resources for physical machines`,
		Long:              `Helper for generating TLS certs and creating resources for physical machine`,
		ValidArgsFunction: noCompletions,
	}

	initCmd, err := physicalmachine.NewPhysicalMachineInitCmd()
	if err != nil {
		return nil, err
	}

	generateCmd, err := physicalmachine.NewPhysicalMachineGenerateCmd()
	if err != nil {
		return nil, err
	}

	createCmd, err := physicalmachine.NewPhysicalMachineCreateCmd()
	if err != nil {
		return nil, err
	}

	physicalMachineCmd.AddCommand(initCmd)
	physicalMachineCmd.AddCommand(generateCmd)
	physicalMachineCmd.AddCommand(createCmd)

	return physicalMachineCmd, nil
}
