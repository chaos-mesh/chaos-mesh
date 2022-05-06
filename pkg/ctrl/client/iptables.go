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

package client

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

func (c *CtrlClient) CleanIptables(ctx context.Context, namespace, name string, chains []string) ([]string, error) {
	var graphqlChains []graphql.String
	for _, chain := range chains {
		graphqlChains = append(graphqlChains, graphql.String(chain))
	}
	var mutation struct {
		Pod struct {
			CleanIptables []string `graphql:"cleanIptables(chains: $chains)"`
		} `graphql:"pod(ns: $ns, name: $name)"`
	}

	err := c.QueryClient.Mutate(ctx, &mutation, map[string]interface{}{
		"chains": graphqlChains,
		"ns":     graphql.String(namespace),
		"name":   graphql.String(name),
	})

	if err != nil {
		return nil, errors.Wrapf(err, "cleaned iptables rules for chains %v", chains)
	}

	return mutation.Pod.CleanIptables, nil
}
