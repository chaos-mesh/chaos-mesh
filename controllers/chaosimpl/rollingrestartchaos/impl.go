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

package rollingrestartchaos

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	k8scr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client
	Log logr.Logger

	decoder *utils.ContainerRecordDecoder
}

type Resource struct {
	Namespace string                        `json:"namespace,omitempty"`
	Name      string                        `json:"name,omitempty"`
	Type      v1alpha1.SelectorResourceType `json:"type,omitempty"`
}

// This corresponds to the Apply phase of RollingRestartChaos. The execution of RollingRestartChaos will be triggered.
func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("rollingrestart Apply", "namespace", obj.GetNamespace(), "name", obj.GetName())

	var res Resource
	if err := json.Unmarshal([]byte(records[index].Id), &res); err != nil {
		return v1alpha1.NotInjected, err
	}

	// Implement logic
	client, err := impl.kubernetesClient()
	if err != nil {
		return v1alpha1.Injected, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Note that "kubectl rollout restart" actually sends a patch request setting 'kubectl.kubernetes.io/restartedAt: <now>'
	// See: https://stackoverflow.com/a/74624618
	data := []byte(fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405")))

	switch res.Type {
	case v1alpha1.DaemonSetResourceType:
		_, err = client.AppsV1().DaemonSets(res.Namespace).Patch(ctx, res.Name, k8stypes.StrategicMergePatchType, data, v1.PatchOptions{})
	case v1alpha1.DeploymentResourceType:
		_, err = client.AppsV1().Deployments(res.Namespace).Patch(ctx, res.Name, k8stypes.StrategicMergePatchType, data, v1.PatchOptions{})
	case v1alpha1.StatefulSetResourceType:
		_, err = client.AppsV1().StatefulSets(res.Namespace).Patch(ctx, res.Name, k8stypes.StrategicMergePatchType, data, v1.PatchOptions{})
	default:
		err = fmt.Errorf("invalid resource type: %s", res.Type)
	}

	if err != nil {
		return v1alpha1.NotInjected, fmt.Errorf("failed to do a rolling restart on %s/%s: %w", res.Type, res.Name, err)
	}

	return v1alpha1.Injected, nil
}

// This corresponds to the Recover phase of HelloWorldChaos. The reconciler will be triggered to recover the chaos action.
func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("rollingrestart Recover", "namespace", obj.GetNamespace(), "name", obj.GetName())
	return v1alpha1.NotInjected, nil
}

// NewImpl returns a new RollingRestartChaos implementation instance.
func NewImpl(c client.Client, log logr.Logger, decoder *utils.ContainerRecordDecoder) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:   "helloworldchaos",
		Object: &v1alpha1.RollingRestartChaos{},
		Impl: &Impl{
			Client:  c,
			Log:     log.WithName("rollingrestartchaos"),
			decoder: decoder,
		},
		ObjectList: &v1alpha1.RollingRestartChaosList{},
	}
}

func (impl *Impl) kubernetesClient() (*kubernetes.Clientset, error) {
	config, err := k8scr.GetConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
