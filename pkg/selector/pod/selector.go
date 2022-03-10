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

package pod

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	genericannotation "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/annotation"
	genericfield "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/field"
	genericlabel "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/label"
	genericnamespace "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/namespace"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/registry"
)

var log = ctrl.Log.WithName("pod-selector")

var ErrNoPodSelected = errors.New("no pod is selected")

type SelectImpl struct {
	c client.Client
	r client.Reader

	generic.Option
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

func (impl *SelectImpl) Select(ctx context.Context, ps *v1alpha1.PodSelector) ([]*Pod, error) {
	if ps == nil {
		return []*Pod{}, nil
	}
	legacyPodSelector := NewLegacyPodSelector(impl.c, impl.r, impl.ClusterScoped, impl.TargetNamespace, impl.EnableFilterNamespace)
	pods, err := legacyPodSelector.SelectAndFilterPods(ctx, ps)
	if err != nil {
		return nil, err
	}

	var result []*Pod
	for _, pod := range pods {
		result = append(result, &Pod{
			pod,
		})
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
		// TODO: remove this reference to global values
		generic.Option{
			ClusterScoped:         config.ControllerCfg.ClusterScoped,
			TargetNamespace:       config.ControllerCfg.TargetNamespace,
			EnableFilterNamespace: config.ControllerCfg.EnableFilterNamespace,
		},
	}
}

//revive:enable:flag-parameter

// GetService get k8s service by service name
func GetService(ctx context.Context, c client.Client, namespace, controllerNamespace string, serviceName string) (*v1.Service, error) {
	// use the environment value if namespace is empty
	if len(namespace) == 0 {
		namespace = controllerNamespace
	}

	service := &v1.Service{}
	err := c.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      serviceName,
	}, service)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func newSelectorRegistry(ctx context.Context, c client.Client, spec v1alpha1.PodSelectorSpec) registry.Registry {
	return map[string]registry.SelectorFactory{
		genericlabel.Name:      genericlabel.New,
		genericnamespace.Name:  genericnamespace.New,
		genericfield.Name:      genericfield.New,
		genericannotation.Name: genericannotation.New,
		nodeSelectorName: func(selector v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
			return newNodeSelector(ctx, c, spec)
		},
		phaseSelectorName: func(selector v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
			return newPhaseSelector(spec)
		},
	}
}
