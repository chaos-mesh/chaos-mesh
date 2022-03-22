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

package server

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	ctrlconfig "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrl/server/model"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

const DefaultNamespace = "default"

func componentLabels(component model.Component) map[string]string {
	var componentLabel string
	switch component {
	case model.ComponentManager:
		componentLabel = "controller-manager"
	case model.ComponentDaemon:
		componentLabel = "chaos-daemon"
	case model.ComponentDashboard:
		componentLabel = "chaos-dashboard"
	case model.ComponentDNSServer:
		componentLabel = "chaos-dns-server"
	default:
		return nil
	}
	return map[string]string{
		"app.kubernetes.io/component": componentLabel,
	}
}

func parseNamespacedName(namespacedName string) types.NamespacedName {
	parts := strings.Split(namespacedName, "/")
	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}
}

// GetPods returns pod list and corresponding chaos daemon
func GetPods(ctx context.Context, status v1alpha1.ChaosStatus, selectorSpec v1alpha1.PodSelectorSpec, c client.Client) ([]v1.Pod, []v1.Pod, error) {
	pods, err := pod.SelectPods(ctx, c, c, selectorSpec, ctrlconfig.ControllerCfg.ClusterScoped, ctrlconfig.ControllerCfg.TargetNamespace, false)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to SelectPods")
	}
	if len(pods) == 0 {
		return nil, nil, nil
	}

	daemonMap, err := getDaemonMap(ctx, c)
	if err != nil {
		return nil, nil, errors.Wrap(err, "get daemon map")
	}

	var chaosDaemons []v1.Pod
	// get chaos daemon
	for _, chaosPod := range pods {
		daemon, exist := daemonMap[chaosPod.Spec.NodeName]
		if !exist {
			return nil, nil, errors.Errorf("no daemons found for pod %s", chaosPod.GetName())
		}
		chaosDaemons = append(chaosDaemons, daemon)
	}

	return pods, chaosDaemons, nil
}

// GetDaemonMap returns a map of node name to daemon pod
func getDaemonMap(ctx context.Context, c client.Client) (map[string]v1.Pod, error) {
	var list v1.PodList
	labels := componentLabels(model.ComponentDaemon)
	if err := c.List(ctx, &list, client.MatchingLabels(labels)); err != nil {
		return nil, errors.Wrapf(err, "list daemons by label %v", labels)
	}

	daemonMap := map[string]v1.Pod{}
	for _, d := range list.Items {
		if d.Spec.NodeName != "" {
			daemonMap[d.Spec.NodeName] = d
		}
	}
	return daemonMap, nil
}
