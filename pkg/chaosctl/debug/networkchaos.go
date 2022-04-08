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
	"strings"

	"github.com/hasura/go-graphql-client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

type networkDebugger struct {
	client *ctrlclient.CtrlClient
}

func NetworkDebug(client *ctrlclient.CtrlClient) Debugger {
	return &networkDebugger{
		client: client,
	}
}

func (d *networkDebugger) Collect(ctx context.Context, namespace, chaosName string) ([]*common.ChaosResult, error) {
	var results []*common.ChaosResult

	var name *graphql.String
	if chaosName != "" {
		n := graphql.String(chaosName)
		name = &n
	}

	var query struct {
		Namespace []struct {
			NetworkChaos []struct {
				Name       string
				Podnetwork []struct {
					Spec      *v1alpha1.PodNetworkChaosSpec
					Namespace string
					Name      string
					Pod       struct {
						Ipset    string
						TcQdisc  []string
						Iptables []string
					}
				}
			} `graphql:"networkchaos(name: $name)"`
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

	for _, networkChaos := range query.Namespace[0].NetworkChaos {
		result := &common.ChaosResult{
			Name: networkChaos.Name,
		}

		for _, podNetworkChaos := range networkChaos.Podnetwork {
			podResult := common.PodResult{
				Name: podNetworkChaos.Name,
			}

			podResult.Items = append(podResult.Items, common.ItemResult{Name: "ipset list", Value: podNetworkChaos.Pod.Ipset})
			podResult.Items = append(podResult.Items, common.ItemResult{Name: "tc qdisc list", Value: strings.Join(podNetworkChaos.Pod.TcQdisc, "\n")})
			podResult.Items = append(podResult.Items, common.ItemResult{Name: "iptables list", Value: strings.Join(podNetworkChaos.Pod.Iptables, "\n")})
			output, err := common.MarshalChaos(podNetworkChaos.Spec)
			if err != nil {
				return nil, err
			}
			podResult.Items = append(podResult.Items, common.ItemResult{Name: "podnetworkchaos", Value: output})
			result.Pods = append(result.Pods, podResult)
		}

		results = append(results, result)
	}
	return results, nil
}

func (d *networkDebugger) List(ctx context.Context, namespace string) ([]string, error) {
	var query struct {
		Namespace []struct {
			NetworkChaos []struct {
				Name string
			} `graphql:"networkchaos"`
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
	for _, networkChaos := range query.Namespace[0].NetworkChaos {
		names = append(names, string(networkChaos.Name))
	}
	return names, nil
}
