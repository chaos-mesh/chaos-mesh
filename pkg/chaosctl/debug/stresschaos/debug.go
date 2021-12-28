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

package stresschaos

import (
	"context"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/bytefmt"
	"github.com/hasura/go-graphql-client"

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
			StressChaos []struct {
				Name      string
				Podstress []struct {
					Pod struct {
						Namespace string
						Name      string
					}
					Cgroups struct {
						Raw string
						Cpu *struct {
							Quota  int
							Period int
						}
						Memory *struct {
							Limit uint64
						}
					}
					ProcessStress []struct {
						Process struct {
							Pid     string
							Command string
						}
						Cgroup string
					}
				}
			} `graphql:"stresschaos(name: $name)"`
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

	for _, stressChaos := range query.Namespace[0].StressChaos {
		result := &common.ChaosResult{
			Name: stressChaos.Name,
		}

		for _, podStressChaos := range stressChaos.Podstress {
			podResult := common.PodResult{
				Name: podStressChaos.Pod.Name,
			}

			podResult.Items = append(podResult.Items, common.ItemResult{Name: "cat /proc/cgroups", Value: podStressChaos.Cgroups.Raw})
			for _, process := range podStressChaos.ProcessStress {
				podResult.Items = append(podResult.Items, common.ItemResult{
					Name:  fmt.Sprintf("/proc/%s/cgroup of %s", process.Process.Pid, process.Process.Command),
					Value: process.Cgroup,
				})
			}
			if podStressChaos.Cgroups.Cpu != nil {
				podResult.Items = append(podResult.Items, common.ItemResult{Name: "cpu.cfs_quota_us", Value: strconv.Itoa(podStressChaos.Cgroups.Cpu.Quota)})
				periodItem := common.ItemResult{Name: "cpu.cfs_period_us", Value: strconv.Itoa(podStressChaos.Cgroups.Cpu.Period)}
				if podStressChaos.Cgroups.Cpu.Quota == -1 {
					periodItem.Status = common.ItemFailure
					periodItem.ErrInfo = "no cpu limit is set for now"
				} else {
					periodItem.Status = common.ItemSuccess
					periodItem.SucInfo = fmt.Sprintf("cpu limit is equals to %.2f", float64(podStressChaos.Cgroups.Cpu.Quota)/float64(podStressChaos.Cgroups.Cpu.Period))
				}
				podResult.Items = append(podResult.Items, periodItem)
			}

			if podStressChaos.Cgroups.Memory != nil {
				podResult.Items = append(podResult.Items, common.ItemResult{Name: "memory.limit_in_bytes", Value: bytefmt.ByteSize(podStressChaos.Cgroups.Memory.Limit) + "B"})
			}
			result.Pods = append(result.Pods, podResult)
		}
		results = append(results, result)
	}
	return results, nil
}
