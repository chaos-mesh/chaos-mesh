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

package debug

import (
	"context"
	"fmt"
	"strings"

	"github.com/hasura/go-graphql-client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

type ioDebugger struct {
	client *ctrlclient.CtrlClient
}

func IODebug(client *ctrlclient.CtrlClient) Debugger {
	return &ioDebugger{
		client: client,
	}
}

func (d *ioDebugger) Collect(ctx context.Context, namespace, chaosName string) ([]*common.ChaosResult, error) {
	var results []*common.ChaosResult

	var name *graphql.String
	if chaosName != "" {
		n := graphql.String(chaosName)
		name = &n
	}

	var query struct {
		Namespace []struct {
			IOChaos []struct {
				Name   string
				Podios []struct {
					Namespace string
					Name      string
					// TODO: fit types with v1alpha1.PodIOChaosSpec
					Spec struct {
						VolumeMountPath string
						Container       *string
						Actions         []struct {
							Type            v1alpha1.IOChaosType
							v1alpha1.Filter `json:",inline"`
							Faults          []v1alpha1.IoFault
							Latency         string
							Ino             *uint64               `json:"ino,omitempty"`
							Size            *uint64               `json:"size,omitempty"`
							Blocks          *uint64               `json:"blocks,omitempty"`
							Atime           *v1alpha1.Timespec    `json:"atime,omitempty"`
							Mtime           *v1alpha1.Timespec    `json:"mtime,omitempty"`
							Ctime           *v1alpha1.Timespec    `json:"ctime,omitempty"`
							Kind            *v1alpha1.FileType    `json:"kind,omitempty"`
							Perm            *uint                 `json:"perm,omitempty"`
							Nlink           *uint                 `json:"nlink,omitempty"`
							UID             *uint                 `json:"uid,omitempty"`
							GID             *uint                 `json:"gid,omitempty"`
							Rdev            *uint                 `json:"rdev,omitempty"`
							Filling         *v1alpha1.FillingType `json:"filling,omitempty"`
							MaxOccurrences  *int64                `json:"maxOccurrences,omitempty"`
							MaxLength       *int64                `json:"maxLength,omitempty"`
						}
					}
					Pod struct {
						Mounts    []string
						Processes []struct {
							Pid     string
							Command string
							Fds     []struct {
								Fd, Target string
							}
						}
					}
				}
			} `graphql:"iochaos(name: $name)"`
		} `graphql:"namespace(ns: $namespace)"`
	}

	variables := map[string]interface{}{
		"namespace": graphql.String(namespace),
		"name":      name,
	}

	err := d.client.QueryClient.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	if len(query.Namespace) == 0 {
		return results, nil
	}

	for _, ioChaos := range query.Namespace[0].IOChaos {
		result := &common.ChaosResult{
			Name: ioChaos.Name,
		}

		for _, podiochaos := range ioChaos.Podios {
			podResult := common.PodResult{
				Name: podiochaos.Name,
			}

			podResult.Items = append(podResult.Items, common.ItemResult{Name: "Mount Information", Value: strings.Join(podiochaos.Pod.Mounts, "\n")})
			for _, process := range podiochaos.Pod.Processes {
				var fds []string
				for _, fd := range process.Fds {
					fds = append(fds, fmt.Sprintf("%s -> %s", fd.Fd, fd.Target))
				}
				podResult.Items = append(podResult.Items, common.ItemResult{
					Name:  fmt.Sprintf("file descriptors of PID: %s, COMMAND: %s", process.Pid, process.Command),
					Value: strings.Join(fds, "\n"),
				})
			}
			output, err := common.MarshalChaos(podiochaos.Spec)
			if err != nil {
				return nil, err
			}
			podResult.Items = append(podResult.Items, common.ItemResult{Name: "podiochaos", Value: output})
			result.Pods = append(result.Pods, podResult)
		}

		results = append(results, result)
	}
	return results, nil
}

func (d *ioDebugger) List(ctx context.Context, namespace string) ([]string, error) {
	var query struct {
		Namespace []struct {
			IOChaos []struct {
				Name string
			} `graphql:"iochaos"`
		} `graphql:"namespace(ns: $namespace)"`
	}

	variables := map[string]interface{}{
		"namespace": graphql.String(namespace),
	}

	err := d.client.QueryClient.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	if len(query.Namespace) == 0 {
		return nil, nil
	}

	var names []string
	for _, ioChaos := range query.Namespace[0].IOChaos {
		names = append(names, string(ioChaos.Name))
	}
	return names, nil
}
