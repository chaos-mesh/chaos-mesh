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

package physicalmachine

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	generic_annotation "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/annotation"
	generic_field "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/field"
	generic_label "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/label"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/registry"
	"go.uber.org/fx"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	generic_namespace "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/namespace"
)

var log = ctrl.Log.WithName("physicalmachineselector")

type SelectImpl struct {
	c client.Client
	r client.Reader

	generic.Option
}

type Params struct {
	fx.In

	Client client.Client
	Reader client.Reader `name:"no-cache"`
}

type PhysicalMachine struct {
	v1alpha1.PhysicalMachine
}

func (pm *PhysicalMachine) Id() string {
	return (types.NamespacedName{
		Name:      pm.Name,
		Namespace: pm.Namespace,
	}).String()
}

func (impl *SelectImpl) Select(ctx context.Context, physicalMachineSelector *v1alpha1.PhysicalMachineSelector) ([]*PhysicalMachine, error) {
	if physicalMachineSelector == nil {
		return []*PhysicalMachine{}, nil
	}

	physicalMachines, err := SelectPhysicalMachines(ctx, impl.c, impl.r, physicalMachineSelector.Selector, impl.ClusterScoped, impl.TargetNamespace, impl.EnableFilterNamespace)
	if err != nil {
		return nil, err
	}

	var result []*PhysicalMachine
	for _, physicalMachine := range physicalMachines {
		result = append(result, &PhysicalMachine{physicalMachine})
	}
	return result, nil
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

func SelectPhysicalMachines(ctx context.Context, c client.Client, r client.Reader,
	selector v1alpha1.PhysicalMachineSelectorSpec,
	clusterScoped bool, targetNamespace string, enableFilterNamespace bool) ([]v1alpha1.PhysicalMachine, error) {
	if len(selector.PhysicalMachines) > 0 {
		return selectSpecifiedPhysicalMachines(ctx, c, selector, clusterScoped, targetNamespace, enableFilterNamespace)
	}

	selectorRegistry := newSelectorRegistry()
	selectorChain, err := registry.Parse(selectorRegistry, selector.GenericSelectorSpec, generic.Option{
		ClusterScoped:         clusterScoped,
		TargetNamespace:       targetNamespace,
		EnableFilterNamespace: enableFilterNamespace,
	})
	if err != nil {
		return nil, err
	}

	var physicalMachines []v1alpha1.PhysicalMachine
	physicalMachines, err = listPhysicalMachines(ctx, c, r, selector, selectorChain, enableFilterNamespace)
	if err != nil {
		return nil, err
	}

	filterList := make([]v1alpha1.PhysicalMachine, 0, len(physicalMachines))
	for _, physicalMachine := range physicalMachines {
		if selectorChain.Match(&physicalMachine) {
			filterList = append(filterList, physicalMachine)
		}
	}
	return filterList, nil
}

func listPhysicalMachines(ctx context.Context, c client.Client, r client.Reader, spec v1alpha1.PhysicalMachineSelectorSpec,
	selectorChain generic.SelectorChain, enableFilterNamespace bool) ([]v1alpha1.PhysicalMachine, error) {
	var physicalMachines []v1alpha1.PhysicalMachine
	namespaceCheck := make(map[string]bool)

	if err := selectorChain.ListObjects(c, r,
		func(listFunc generic.ListFunc, opts client.ListOptions) error {
			var pmList v1alpha1.PhysicalMachineList
			if len(spec.Namespaces) > 0 {
				for _, namespace := range spec.Namespaces {
					if enableFilterNamespace {
						allow, ok := namespaceCheck[namespace]
						if !ok {
							allow = generic_namespace.CheckNamespace(ctx, c, namespace)
							namespaceCheck[namespace] = allow
						}
						if !allow {
							continue
						}
					}

					opts.Namespace = namespace
					if err := listFunc(ctx, &pmList, &opts); err != nil {
						return err
					}
					physicalMachines = append(physicalMachines, pmList.Items...)
				}
			} else {
				// in fact, this will never happen
				if err := listFunc(ctx, &pmList, &opts); err != nil {
					return err
				}
				physicalMachines = append(physicalMachines, pmList.Items...)
			}
			return nil
		}); err != nil {
		return nil, err
	}

	return physicalMachines, nil
}

func newSelectorRegistry() registry.Registry {
	return map[string]registry.SelectorFactory{
		generic_label.Name:      generic_label.New,
		generic_namespace.Name:  generic_namespace.New,
		generic_field.Name:      generic_field.New,
		generic_annotation.Name: generic_annotation.New,
	}
}

func selectSpecifiedPhysicalMachines(ctx context.Context, c client.Client, spec v1alpha1.PhysicalMachineSelectorSpec,
	clusterScoped bool, targetNamespace string, enableFilterNamespace bool) ([]v1alpha1.PhysicalMachine, error) {
	var physicalMachines []v1alpha1.PhysicalMachine
	namespaceCheck := make(map[string]bool)

	for ns, names := range spec.PhysicalMachines {
		if !clusterScoped {
			if targetNamespace != ns {
				log.Info("skip namespace because ns is out of scope within namespace scoped mode", "namespace", ns)
				continue
			}
		}
		if enableFilterNamespace {
			allow, ok := namespaceCheck[ns]
			if !ok {
				allow = generic_namespace.CheckNamespace(ctx, c, ns)
				namespaceCheck[ns] = allow
			}
			if !allow {
				continue
			}
		}
		for _, name := range names {
			var physicalMachine v1alpha1.PhysicalMachine
			err := c.Get(ctx, types.NamespacedName{
				Namespace: ns,
				Name:      name,
			}, &physicalMachine)
			if err == nil {
				physicalMachines = append(physicalMachines, physicalMachine)
				continue
			}

			if apierrors.IsNotFound(err) {
				log.Error(err, "PhysicalMachine is not found", "namespace", ns, "physical machine name", name)
				continue
			}

			return nil, err
		}
	}
	return physicalMachines, nil
}
