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

package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
)

type logsOptions struct {
	tail int64
	node string
}

type Component string

const (
	Manager   Component = "MANAGER"
	Daemon    Component = "DAEMON"
	Dashboard Component = "DASHBOARD"
	DnsServer Component = "DNSSERVER"
)

func NewLogsCmd() (*cobra.Command, error) {
	o := &logsOptions{}

	logsCmd := &cobra.Command{
		Use:   `logs [-t LINE]`,
		Short: `Print logs of controller-manager, chaos-daemon and chaos-dashboard`,
		Long: `Print logs of controller-manager, chaos-daemon and chaos-dashboard, to provide debug information.

Examples:
  # Default print all log of all chaosmesh components
  chaosctl logs

  # Print 100 log lines for chaosmesh components in node NODENAME
  chaosctl logs -t 100 -n NODENAME`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run(args)
		},
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: noCompletions,
	}

	logsCmd.Flags().Int64VarP(&o.tail, "tail", "t", -1, "number of lines of recent log")
	logsCmd.Flags().StringVarP(&o.node, "node", "n", "", "the node of target pods")
	err := logsCmd.RegisterFlagCompletionFunc("node", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		clientset, err := cm.InitClientSet()
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}
		return listNodes(toComplete, clientset.KubeCli)
	})
	if err != nil {
		return nil, err
	}
	return logsCmd, nil
}

// Run logs
func (o *logsOptions) Run(args []string) error {
	client, cancel, err := cm.CreateClient(context.TODO(), managerNamespace, managerSvc)
	if err != nil {
		return errors.Wrap(err, "failed to initialize clientset")
	}
	defer cancel()

	componentsNeeded := []Component{Manager, Daemon, Dashboard, DnsServer}
	for _, name := range componentsNeeded {
		var query struct {
			Namespace []struct {
				Component []struct {
					Name string
					Spec struct {
						NodeName string
					}
					Logs string
				} `graphql:"component(component: $component)"`
			} `graphql:"namespace(ns: $namespace)"`
		}

		variables := map[string]interface{}{
			"namespace": graphql.String(managerNamespace),
			"component": name,
		}

		err := client.QueryClient.Query(context.TODO(), &query, variables)
		if err != nil {
			return err
		}

		if len(query.Namespace) == 0 {
			return fmt.Errorf("no namespace %s found", managerNamespace)
		}

		for _, component := range query.Namespace[0].Component {
			if o.node != "" && component.Spec.NodeName != o.node {
				// ignore component on this node
				continue
			}

			logLines := strings.Split(component.Logs, "\n")
			if o.tail > 0 {
				if len(logLines) > int(o.tail) {
					logLines = logLines[len(logLines)-int(o.tail)-1:]
				}
			}

			cm.PrettyPrint(fmt.Sprintf("[%s]", component.Name), 0, cm.Cyan)
			cm.PrettyPrint(strings.Join(logLines, "\n"), 1, cm.NoColor)
		}
	}
	return nil
}

func listNodes(toComplete string, c *kubernetes.Clientset) ([]string, cobra.ShellCompDirective) {
	// FIXME: get context from parameter
	nodes, err := c.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}
	var ret []string
	for _, ns := range nodes.Items {
		if strings.HasPrefix(ns.Name, toComplete) {
			ret = append(ret, ns.Name)
		}
	}
	return ret, cobra.ShellCompDirectiveNoFileComp
}
