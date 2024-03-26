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

package podpvc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

var ErrNoPodSelected = errors.New("no pod is selected")

type SelectImpl struct {
	c client.Client
	r client.Reader

	generic.Option
}

type PodPVCTarget struct {
	Pod types.NamespacedName
	PVC types.NamespacedName
}

func (dep *PodPVCTarget) Id() string {
	b, _ := json.Marshal(dep)
	return string(b)
}

type Pod struct {
	v1.Pod
}

func (pod *Pod) Id() string {
	return (types.NamespacedName{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}).String()
}

func (impl *SelectImpl) Select(ctx context.Context, ps *v1alpha1.PodPVCSelector) ([]*PodPVCTarget, error) {
	if ps == nil {
		return []*PodPVCTarget{}, nil
	}

	pods, err := pod.SelectAndFilterPods(ctx, impl.c, impl.r, &ps.PodSelector, impl.ClusterScoped, impl.TargetNamespace, impl.EnableFilterNamespace)
	if err != nil {
		return nil, err
	}

	var result []*PodPVCTarget
	for _, pod := range pods {
		spec, err := SelectVolume(ctx, impl.c, impl.r, ps, &pod)
		if err != nil {
			return nil, err
		}
		result = append(result, spec)
	}

	return result, nil
}

type Params struct {
	fx.In

	Client client.Client
	Reader client.Reader `name:"no-cache"`
}

func New(params Params) *SelectImpl {
	return &SelectImpl{
		params.Client,
		params.Reader,
		generic.Option{
			ClusterScoped:         config.ControllerCfg.ClusterScoped,
			TargetNamespace:       config.ControllerCfg.TargetNamespace,
			EnableFilterNamespace: config.ControllerCfg.EnableFilterNamespace,
		},
	}
}

// SelectAndFilterPods returns the list of pods that filtered by selector and SelectorMode
// Deprecated: use pod.SelectImpl as instead
func SelectVolume(ctx context.Context, c client.Client, r client.Reader, selector *v1alpha1.PodPVCSelector, pod *v1.Pod) (*PodPVCTarget, error) {

	spec := &PodPVCTarget{Pod: types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}}
	for _, v := range pod.Spec.Volumes {
		if v.Name == selector.VolumeName {
			spec.PVC = types.NamespacedName{Namespace: pod.Namespace, Name: v.PersistentVolumeClaim.ClaimName}
			return spec, nil
		}
	}

	return nil, fmt.Errorf("volume not found: %s", selector.VolumeName)
}
