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

package resourcescale

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type SelectImpl struct{}

type ResourceSpec struct {
	Selector *v1alpha1.ResourceScaleSelector `json:"selector,omitempty"`
	Replicas int32                           `json:"replicas,omitempty"`
}

func (dep *ResourceSpec) Id() string {
	b, _ := json.Marshal(dep)
	return string(b)
}

func (impl *SelectImpl) Select(ctx context.Context, selector *v1alpha1.ResourceScaleSelector) ([]*ResourceSpec, error) {
	client, err := kubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	specs := &ResourceSpec{
		Replicas: 0,
		Selector: selector,
	}

	// Get current number of replicas
	switch selector.ResourceType {
	case v1alpha1.ResourceTypeDaemonSet:
		ds, err := client.AppsV1().DaemonSets(selector.Namespace).Get(ctx, selector.Name, v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to select daemonset %q: %w", selector.Namespace+"/"+selector.Name, err)
		}

		// Daemonsets can only scale to 0 or 1 pod per node
		if ds.Status.DesiredNumberScheduled > 0 {
			specs.Replicas = 1
		}
	case v1alpha1.ResourceTypeDeployment:
		dep, err := client.AppsV1().Deployments(selector.Namespace).GetScale(ctx, selector.Name, v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to select deployment %q: %w", selector.Namespace+"/"+selector.Name, err)
		}
		specs.Replicas = dep.Status.Replicas
	case v1alpha1.ResourceTypeReplicaSet:
		rs, err := client.AppsV1().ReplicaSets(selector.Namespace).GetScale(ctx, selector.Name, v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to select replicaset %q: %w", selector.Namespace+"/"+selector.Name, err)
		}
		specs.Replicas = rs.Status.Replicas
	case v1alpha1.ResourceTypeStatefulSet:
		sts, err := client.AppsV1().StatefulSets(selector.Namespace).GetScale(ctx, selector.Name, v1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to select statefulset %q: %w", selector.Namespace+"/"+selector.Name, err)
		}
		specs.Replicas = sts.Status.Replicas
	default:
		return nil, fmt.Errorf("failed to get select to get resource %q: invalid resource type %s", selector.Namespace+"/"+selector.Name, selector.ResourceType)
	}

	return []*ResourceSpec{specs}, nil
}

func New() *SelectImpl {
	return &SelectImpl{}
}

func kubernetesClient() (*kubernetes.Clientset, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
