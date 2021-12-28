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

package httpchaos

import (
	"context"
	"fmt"
	"strings"

	"github.com/hasura/go-graphql-client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

// Debug get chaos debug information
func Debug(ctx context.Context, namespace, chaosName string, client *ctrlclient.CtrlClient) ([]*common.ChaosResult, error) {
	var results []*common.ChaosResult

	var name *graphql.String
	if chaosName != "" {
		n := graphql.String(chaosName)
		name = &n
	}

	var query struct {
		Namespace []struct {
			HTTPChaos []struct {
				Name    string
				Podhttp []struct {
					Namespace string
					Name      string
					Spec      *v1alpha1.PodHttpChaosSpec
					Pod       struct {
						Iptables  string
						Processes []struct {
							Pid     string
							Command string
							Fds     []struct {
								Fd, Target string
							}
						}
					}
				}
			} `graphql:"httpchaos(name: $name)"`
		} `graphql:"namespace(ns: $namespace)"`
	}

	variables := map[string]interface{}{
		"namespace": graphql.String(namespace),
		"name":      name,
	}

	err := client.Client.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	if len(query.Namespace) == 0 {
		return results, nil
	}

	for _, httpChaos := range query.Namespace[0].HTTPChaos {
		result := &common.ChaosResult{
			Name: string(httpChaos.Name),
		}

		for _, podhttpchaos := range httpChaos.Podhttp {
			podResult := common.PodResult{
				Name: string(podhttpchaos.Name),
			}

			podResult.Items = append(podResult.Items, common.ItemResult{Name: "iptables list", Value: string(podhttpchaos.Pod.Iptables)})
			for _, process := range podhttpchaos.Pod.Processes {
				var fds []string
				for _, fd := range process.Fds {
					fds = append(fds, fmt.Sprintf("%s -> %s", fd.Fd, fd.Target))
				}
				podResult.Items = append(podResult.Items, common.ItemResult{
					Name:  fmt.Sprintf("file descriptors of PID: %s, COMMAND: %s", process.Pid, process.Command),
					Value: strings.Join(fds, "\n"),
				})
			}
			output, err := common.MarshalChaos(podhttpchaos.Spec)
			if err != nil {
				return nil, err
			}
			podResult.Items = append(podResult.Items, common.ItemResult{Name: "podhttpchaos", Value: output})
			result.Pods = append(result.Pods, podResult)
		}

		results = append(results, result)
	}
	return results, nil
}
