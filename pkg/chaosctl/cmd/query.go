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
	"strings"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "get [QUERY]",
	Short: "get the target resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		// TODO: input ns by args
		cancel, port, err := common.ForwardCtrlServer(ctx, nil)
		if err != nil {
			return err
		}
		defer cancel()

		client, err := common.NewCtrlClient(ctx, fmt.Sprintf("http://127.0.0.1:%d/query", port))
		if err != nil {
			return fmt.Errorf("fail to init ctrl client: %s", err)
		}

		if client.Schema == nil {
			return errors.New("fail to fetch schema")
		}

		queryType, err := client.Schema.MustGetType(string(client.Schema.QueryType.Name))
		if err != nil {
			return err
		}

		query, err := client.Schema.ParseQuery(strings.Split(strings.Trim(args[0], "/"), "/"), queryType)
		if err != nil {
			return err
		}

		fmt.Println(query)
		return nil
	},
}
