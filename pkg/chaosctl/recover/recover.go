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

package recover

import (
	"context"
	"fmt"
	"strings"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"

	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

// PartialPod is a subset of the Pod type.
// It contains necessary information for forced recovery.
type PartialPod struct {
	Namespace string
	Name      string
	Processes []struct {
		Pid, Command string
	}
	TcQdisc  []string
	Iptables []string
}

type Recoverer interface {
	// Recover target pod forcedly
	Recover(ctx context.Context, pod *PartialPod) error
}

type RecovererBuilder func(client *ctrlclient.CtrlClient) Recoverer

type cleanProcessRecoverer struct {
	client  *ctrlclient.CtrlClient
	process string
}

func newCleanProcessRecoverer(client *ctrlclient.CtrlClient, process string) Recoverer {
	return &cleanProcessRecoverer{
		client:  client,
		process: process,
	}
}

func (r *cleanProcessRecoverer) Recover(ctx context.Context, pod *PartialPod) error {
	var pids []graphql.String
	for _, process := range pod.Processes {
		if process.Command == r.process {
			pids = append(pids, graphql.String(process.Pid))
		}
	}

	if len(pids) == 0 {
		printStep(fmt.Sprintf("all %s processes are cleaned up", r.process))
		return nil
	}
	printStep(fmt.Sprintf("cleaning %s processes: %v", r.process, pids))

	var mutation struct {
		Pod struct {
			KillProcesses []struct {
				Pid, Command string
			} `graphql:"killProcesses(pids: $pids)"`
		} `graphql:"pod(ns: $ns, name: $name)"`
	}

	err := r.client.QueryClient.Mutate(ctx, &mutation, map[string]interface{}{
		"pids": pids,
		"ns":   graphql.String(pod.Namespace),
		"name": graphql.String(pod.Name),
	})

	if err != nil {
		return errors.Wrapf(err, "kill %s processes(%v)", r.process, pids)
	}

	if len(mutation.Pod.KillProcesses) != 0 {
		var cleanedPids []string
		for _, process := range mutation.Pod.KillProcesses {
			cleanedPids = append(cleanedPids, process.Pid)
		}
		printStep(fmt.Sprintf("%s processes(%s) are cleaned up", r.process, strings.Join(cleanedPids, ", ")))
	}

	return nil
}
