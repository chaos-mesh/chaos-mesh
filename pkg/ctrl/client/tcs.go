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

func (c *CtrlClient) CleanTcs(ctx context.Context, namespace, name string, devices []string) ([]string, error) {
	var graphqlDevices []graphql.String
	for _, dev := range devices {
		graphqlDevices = append(graphqlDevices, graphql.String(dev))
	}
	var mutation struct {
		Pod struct {
			CleanTcs []string `graphql:"cleanTcs(devices: $devices)"`
		} `graphql:"pod(ns: $ns, name: $name)"`
	}

	err := c.QueryClient.Mutate(ctx, &mutation, map[string]interface{}{
		"devices": graphqlDevices,
		"ns":      graphql.String(namespace),
		"name":    graphql.String(name),
	})

	if err != nil {
		return nil, errors.Wrapf(err, "cleaned tc rules for device %v", devices)
	}

	return mutation.Pod.CleanTcs, nil
}
