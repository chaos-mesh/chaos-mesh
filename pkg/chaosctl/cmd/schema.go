// Copyright 2021 Chaos Mesh Authors.
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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

// schemaCmd represents the schema command
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "get the graphql schema of ctrl server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrlclient.DisableRuntimeErrorHandler()
		ctx := context.Background()
		// TODO: input ns by args
		cancel, port, err := ctrlclient.ForwardCtrlServer(ctx, managerNamespace)
		if err != nil {
			return err
		}
		defer cancel()

		client, err := ctrlclient.NewCtrlClient(ctx, fmt.Sprintf("http://127.0.0.1:%d/query", port))
		if err != nil {
			return fmt.Errorf("fail to init ctrl client: %s", err)
		}

		if client.Schema == nil {
			return errors.New("fail to fetch schema")
		}

		data, err := yaml.Marshal(client.Schema)
		if err != nil {
			return err
		}

		fmt.Print(string(data))
		return nil
	},
}
