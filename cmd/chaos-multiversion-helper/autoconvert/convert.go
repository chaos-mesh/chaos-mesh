// Copyright 2022 Chaos Mesh Authors.
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

package autoconvert

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewConvertCmd(log logr.Logger) *cobra.Command {
	var version, hub string

	var cmd = &cobra.Command{
		Use:   "autoconvert --version <version> --hub <hub-version>",
		Short: "Generates code to automatically convert between two versions",
		Long: `Autoconvert will do the following things:
		1. remove the Hub declaration in <version>, if it is.
		2. create the Hub tag for <hub-version>, if it is not.
		3. generate ConvertTo and ConvertFrom function for the <version>, and assume it has 
			a type which is deeply the same with the <hub-version>.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(log, version, hub)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "the version to generate the convert")
	cmd.Flags().StringVar(&hub, "hub", "", "the hub version")

	err := cmd.MarkFlagRequired("version")
	if err != nil {
		log.Error(errors.WithStack(err), "fail to mark 'version' as required")
		panic("unreachable")
	}
	err = cmd.MarkFlagRequired("hub")
	if err != nil {
		log.Error(errors.WithStack(err), "fail to mark 'hub' as required")
		panic("unreachable")
	}

	return cmd
}

func removeHub(version string) error {
	generatedHubFile := fmt.Sprintf("api/%s/zz_generated.hub.chaosmesh.go", version)
	err := os.Remove(generatedHubFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return errors.Wrapf(err, "remove file %s", generatedHubFile)
	}

	return nil
}

func run(log logr.Logger, version, hub string) error {
	err := removeHub(version)
	if err != nil {
		return err
	}

	err = setHub(hub)
	if err != nil {
		return err
	}

	err = generateConvert(log, version, hub)
	if err != nil {
		return err
	}
	log.Info("generate convert successfully")

	return nil
}
