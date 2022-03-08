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

type httpRecover struct {
	client *ctrlclient.CtrlClient
}

func HTTPRecover(client *ctrlclient.CtrlClient) Recover {
	return &httpRecover{
		client: client,
	}
}

func (r *httpRecover) Recover(ctx context.Context, namespace, podName string) error {
	var query struct {
		Namespace []struct {
			Pod []struct {
				Namespace string
				Name      string
				Processes []struct {
					Pid, Command string
				}
			} `graphql:"pod(name: $name)"`
		} `graphql:"namespace(ns: $namespace)"`
	}

	err := r.client.QueryClient.Query(ctx, &query, map[string]interface{}{
		"namespace": graphql.String(namespace),
		"name":      graphql.String(podName),
	})

	if err != nil {
		errors.Wrapf(err, "query pod %s in namespace %s", podName, namespace)
	}

	if len(query.Namespace) == 0 || len(query.Namespace[0].Pod) == 0 {
		return errors.Errorf("pod %s in namespace %s not found", podName, namespace)
	}

	pod := query.Namespace[0].Pod[0]
	printRecover(fmt.Sprintf("recovering HTTPChaos from pod %s/%s", pod.Namespace, pod.Name))

	var pids []graphql.String
	for _, process := range pod.Processes {
		if process.Command == "tproxy" {
			pids = append(pids, graphql.String(process.Pid))
		}
	}

	if len(pids) == 0 {
		printStep("all tproxy processes are cleaned up")
	} else {
		printStep(fmt.Sprintf("cleaning tproxy processes: %v", pids))
	}

	var mutation struct {
		Pod struct {
			KillProcesses []struct {
				Pid, Command string
			} `graphql:"killProcesses(pids: $pids)"`
		} `graphql:"pod(ns: $ns, name: $name)"`
	}

	err = r.client.QueryClient.Mutate(ctx, &mutation, map[string]interface{}{
		"pids": pids,
		"ns":   graphql.String(pod.Namespace),
		"name": graphql.String(pod.Name),
	})

	if err != nil {
		return errors.Wrapf(err, "kill tproxy processes(%v)", pids)
	}

	if len(mutation.Pod.KillProcesses) != 0 {
		var cleanedPids []string
		for _, process := range mutation.Pod.KillProcesses {
			cleanedPids = append(cleanedPids, process.Pid)
		}
		printStep(fmt.Sprintf("tproxy processes(%s) are cleaned up", strings.Join(cleanedPids, ", ")))
	}

	return nil
}
