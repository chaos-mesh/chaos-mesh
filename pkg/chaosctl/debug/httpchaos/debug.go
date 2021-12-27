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
			HTTPChaos []struct {
				Name    string
				Podhttp []struct {
					Namespace string
					Name      string
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
			result.Pods = append(result.Pods, podResult)
		}

		results = append(results, result)
	}
	return results, nil
}
