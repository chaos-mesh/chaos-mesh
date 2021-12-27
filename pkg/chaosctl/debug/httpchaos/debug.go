package httpchaos

import (
	"context"
	"fmt"
	"strings"

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
			Httpchaos []struct {
				Name    graphql.String
				Podhttp []struct {
					Namespace graphql.String
					Name      graphql.String
					Pod       struct {
						Ipset     graphql.String
						Iptables  graphql.String
						Processes []struct {
							Pid     graphql.String
							Command graphql.String
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

	httpChaosList := &query.Namespace[0].Httpchaos

	for _, httpChaos := range *httpChaosList {
		result := &common.ChaosResult{
			Name: string(httpChaos.Name),
		}

		for _, podhttpchaos := range httpChaos.Podhttp {
			podResult := common.PodResult{
				Name: string(podhttpchaos.Name),
			}

			podResult.Items = append(podResult.Items, common.ItemResult{Name: "IP Set", Value: string(podhttpchaos.Pod.Ipset)})
			podResult.Items = append(podResult.Items, common.ItemResult{Name: "Iptables", Value: string(podhttpchaos.Pod.Iptables)})
			for _, process := range podhttpchaos.Pod.Processes {
				if strings.Contains(string(process.Command), "tproxy") {
					podResult.Items = append(podResult.Items, common.ItemResult{Name: fmt.Sprintf("Process %s", process.Pid), Value: string(process.Command)})
				}
			}
			result.Pods = append(result.Pods, podResult)
		}

		results = append(results, result)
	}
	return results, nil
}
