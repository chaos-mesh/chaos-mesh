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

package rollingrestart

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

type Resource struct {
	Namespace string                        `json:"namespace,omitempty"`
	Name      string                        `json:"name,omitempty"`
	Type      v1alpha1.SelectorResourceType `json:"type,omitempty"`
}

func (dep *Resource) Id() string {
	b, _ := json.Marshal(dep)
	return string(b)
}

func (impl *SelectImpl) Select(ctx context.Context, rrSelector *v1alpha1.RollingRestartSelector) ([]*Resource, error) {
	client, err := kubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	switch rrSelector.ResourceType {
	case v1alpha1.DaemonSetResourceType:
		_, err = client.AppsV1().DaemonSets(rrSelector.Namespace).Get(ctx, rrSelector.Name, v1.GetOptions{})
	case v1alpha1.DeploymentResourceType:
		_, err = client.AppsV1().Deployments(rrSelector.Namespace).Get(ctx, rrSelector.Name, v1.GetOptions{})
	case v1alpha1.StatefulSetResourceType:
		_, err = client.AppsV1().StatefulSets(rrSelector.Namespace).Get(ctx, rrSelector.Name, v1.GetOptions{})
	default:
		err = fmt.Errorf("invalid resource type %s", rrSelector.ResourceType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get %s for %q: %w", rrSelector.ResourceType, rrSelector.Namespace+"/"+rrSelector.Name, err)
	}

	dep := &Resource{
		Namespace: rrSelector.Namespace,
		Name:      rrSelector.Name,
		Type:      rrSelector.ResourceType,
	}

	return []*Resource{dep}, nil
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
