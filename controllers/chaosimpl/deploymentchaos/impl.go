// Copyright 2023 Chaos Mesh Authors.
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

package deploymentchaos

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Deployment struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Replicas  int32  `json:"replicas,omitempty"`
}

type Impl struct {
	client.Client
	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("deploymentchaos Apply", "namespace", obj.GetNamespace(), "name", obj.GetName())

	var dep Deployment
	if err := json.Unmarshal([]byte(records[index].Id), &dep); err != nil {
		return v1alpha1.NotInjected, err
	}

	client, err := impl.kubernetesClient()
	if err != nil {
		return v1alpha1.NotInjected, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	scale, err := client.AppsV1().Deployments(dep.Namespace).GetScale(ctx, dep.Name, metav1.GetOptions{})
	if err != nil {
		return v1alpha1.NotInjected, fmt.Errorf("failed to get deployment scale: %w", err)
	}

	scale.Spec.Replicas = 0

	_, err = client.AppsV1().Deployments(dep.Namespace).UpdateScale(ctx, dep.Name, scale, metav1.UpdateOptions{})
	if err != nil {
		return v1alpha1.NotInjected, fmt.Errorf("failed to scale deployment: %w", err)
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("deploymentchaos Recover", "namespace", obj.GetNamespace(), "name", obj.GetName())

	var dep Deployment
	if err := json.Unmarshal([]byte(records[index].Id), &dep); err != nil {
		return v1alpha1.Injected, err
	}

	client, err := impl.kubernetesClient()
	if err != nil {
		return v1alpha1.Injected, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	scale, err := client.AppsV1().Deployments(dep.Namespace).GetScale(ctx, dep.Name, metav1.GetOptions{})
	if err != nil {
		return v1alpha1.NotInjected, fmt.Errorf("failed to get deployment scale: %w", err)
	}

	scale.Spec.Replicas = dep.Replicas

	_, err = client.AppsV1().Deployments(dep.Namespace).UpdateScale(ctx, dep.Name, scale, metav1.UpdateOptions{})
	if err != nil {
		return v1alpha1.Injected, fmt.Errorf("failed to scale deployment: %w", err)
	}

	return v1alpha1.NotInjected, nil
}

func (impl *Impl) kubernetesClient() (*kubernetes.Clientset, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func NewImpl(c client.Client, log logr.Logger) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:   "deploymentchaos",
		Object: &v1alpha1.DeploymentChaos{},
		Impl: &Impl{
			Client: c,
			Log:    log.WithName("deploymentchaos"),
		},
		ObjectList: &v1alpha1.DeploymentChaosList{},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
