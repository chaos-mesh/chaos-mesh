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

type Recover interface {
	// Recover target pod forcedly
	Recover(ctx context.Context, pod *ctrlclient.PartialPod) error
}

type RecoverBuilder func(client *ctrlclient.CtrlClient) Recover

type PipelineRecover struct {
	chaosName string
	recovers  []Recover
}

type CleanProcessRecover struct {
	client  *ctrlclient.CtrlClient
	process string
}

func PipelineBuilder(chaosName string, builders ...RecoverBuilder) RecoverBuilder {
	return func(client *ctrlclient.CtrlClient) Recover {
		pipeline := &PipelineRecover{chaosName: chaosName}
		for _, builder := range builders {
			pipeline.recovers = append(pipeline.recovers, builder(client))
		}
		return pipeline
	}
}

func (r *PipelineRecover) Recover(ctx context.Context, pod *ctrlclient.PartialPod) error {
	printRecover(fmt.Sprintf("recovering %s from pod %s/%s", r.chaosName, pod.Namespace, pod.Name))
	for _, recover := range r.recovers {
		if err := recover.Recover(ctx, pod); err != nil {
			return errors.Wrapf(err, "recover pod %s/%s", pod.Namespace, pod.Name)
		}
	}
	return nil
}

func CleanProcessRecoverBuilder(process string) RecoverBuilder {
	return func(client *ctrlclient.CtrlClient) Recover {
		return &CleanProcessRecover{
			client:  client,
			process: process,
		}
	}
}

func (r *CleanProcessRecover) Recover(ctx context.Context, pod *ctrlclient.PartialPod) error {
	var pids []graphql.String
	for _, process := range pod.Processes {
		if process.Command == r.process {
			pids = append(pids, graphql.String(process.Pid))
		}
	}

	if len(pids) == 0 {
		printStep(fmt.Sprintf("all %s processes are cleaned up", r.process))
		return nil
	} else {
		printStep(fmt.Sprintf("cleaning %s processes: %v", r.process, pids))
	}

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
