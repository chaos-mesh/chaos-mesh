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

func (c *CtrlClient) KillProcesses(ctx context.Context, namespace, name string, pids []string) ([]string, error) {
	var graphqlPids []graphql.String
	for _, p := range pids {
		graphqlPids = append(graphqlPids, graphql.String(p))
	}
	var mutation struct {
		Pod struct {
			KillProcesses []struct {
				Pid, Command string
			} `graphql:"killProcesses(pids: $pids)"`
		} `graphql:"pod(ns: $ns, name: $name)"`
	}

	err := c.QueryClient.Mutate(ctx, &mutation, map[string]interface{}{
		"pids": graphqlPids,
		"ns":   graphql.String(namespace),
		"name": graphql.String(name),
	})

	if err != nil {
		return nil, errors.Wrapf(err, "kill processes(%v)", pids)
	}

	var killedPids []string
	for _, p := range mutation.Pod.KillProcesses {
		killedPids = append(killedPids, p.Pid)
	}
	return killedPids, nil
}
