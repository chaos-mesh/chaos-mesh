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

	"github.com/pkg/errors"
	"go.uber.org/fx"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	stdLog "github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	genericannotation "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/annotation"
	genericfield "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/field"
	genericlabel "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/label"
	genericnamespace "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/namespace"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/registry"
)

var log = ctrl.Log.WithName("physical-machine-selector")

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
	Address string
}

func (pm *PhysicalMachine) Id() string {
	if len(pm.Address) > 0 {
		return pm.Address
	}
	return (types.NamespacedName{
		Name:      pm.Name,
		Namespace: pm.Namespace,
	}).String()
}

func (impl *SelectImpl) Select(ctx context.Context, physicalMachineSelector *v1alpha1.PhysicalMachineSelector) ([]*PhysicalMachine, error) {
	if physicalMachineSelector == nil {
		return []*PhysicalMachine{}, nil
	}

	physicalMachines, err := SelectAndFilterPhysicalMachines(ctx, impl.c, impl.r, physicalMachineSelector, impl.ClusterScoped, impl.TargetNamespace, impl.EnableFilterNamespace)
	if err != nil {
		return nil, err
	}

	filtered, err := filterPhysicalMachinesByMode(physicalMachines, physicalMachineSelector.Mode, physicalMachineSelector.Value)
	if err != nil {
		return nil, err
	}
	return filtered, nil
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

// SelectAndFilterPhysicalMachines returns the list of physical machines that filtered by selector and SelectorMode
func SelectAndFilterPhysicalMachines(ctx context.Context, c client.Client, r client.Reader, spec *v1alpha1.PhysicalMachineSelector, clusterScoped bool, targetNamespace string, enableFilterNamespace bool) ([]*PhysicalMachine, error) {
	if len(spec.Address) > 0 {
		var result []*PhysicalMachine
		for _, address := range spec.Address {
			result = append(result, &PhysicalMachine{
				Address: address,
			})
		}
		return result, nil
	}

	physicalMachines, err := SelectPhysicalMachines(ctx, c, r, spec.Selector, clusterScoped, targetNamespace, enableFilterNamespace)
	if err != nil {
		return nil, err
	}

	if len(physicalMachines) == 0 {
		err = errors.New("no physical machine is selected")
		return nil, err
	}

	var result []*PhysicalMachine
	for _, physicalMachine := range physicalMachines {
		result = append(result, &PhysicalMachine{
			PhysicalMachine: physicalMachine,
		})
	}
	return result, nil
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

	return listPhysicalMachines(ctx, c, r, selector, selectorChain, enableFilterNamespace)
}

func listPhysicalMachines(ctx context.Context, c client.Client, r client.Reader, spec v1alpha1.PhysicalMachineSelectorSpec,
	selectorChain generic.SelectorChain, enableFilterNamespace bool) ([]v1alpha1.PhysicalMachine, error) {
	var physicalMachines []v1alpha1.PhysicalMachine
	namespaceCheck := make(map[string]bool)

	if err := selectorChain.ListObjects(c, r,
		func(listFunc generic.ListFunc, opts client.ListOptions) error {
			var pmList v1alpha1.PhysicalMachineList
			if len(spec.Namespaces) > 0 {
				logger, err := stdLog.NewDefaultZapLogger()
				if err != nil {
					return errors.Wrap(err, "failed to create logger")
				}
				for _, namespace := range spec.Namespaces {
					if enableFilterNamespace {
						allow, ok := namespaceCheck[namespace]
						if !ok {
							allow = genericnamespace.CheckNamespace(ctx, c, namespace, logger)
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

	filterList := make([]v1alpha1.PhysicalMachine, 0, len(physicalMachines))
	for _, physicalMachine := range physicalMachines {
		physicalMachine := physicalMachine
		if selectorChain.Match(&physicalMachine) {
			filterList = append(filterList, physicalMachine)
		}
	}
	return filterList, nil
}

func newSelectorRegistry() registry.Registry {
	return map[string]registry.SelectorFactory{
		genericlabel.Name:      genericlabel.New,
		genericnamespace.Name:  genericnamespace.New,
		genericfield.Name:      genericfield.New,
		genericannotation.Name: genericannotation.New,
	}
}

func selectSpecifiedPhysicalMachines(ctx context.Context, c client.Client, spec v1alpha1.PhysicalMachineSelectorSpec,
	clusterScoped bool, targetNamespace string, enableFilterNamespace bool) ([]v1alpha1.PhysicalMachine, error) {
	var physicalMachines []v1alpha1.PhysicalMachine
	namespaceCheck := make(map[string]bool)
	logger, err := stdLog.NewDefaultZapLogger()
	if err != nil {
		return physicalMachines, errors.Wrap(err, "failed to create logger")
	}
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
				allow = genericnamespace.CheckNamespace(ctx, c, ns, logger)
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

// filterPhysicalMachinesByMode filters physical machines by mode from physical machine list
func filterPhysicalMachinesByMode(physicalMachines []*PhysicalMachine, mode v1alpha1.SelectorMode, value string) ([]*PhysicalMachine, error) {
	indexes, err := generic.FilterObjectsByMode(mode, value, len(physicalMachines))
	if err != nil {
		return nil, err
	}

	var filtered []*PhysicalMachine
	for _, index := range indexes {
		index := index
		filtered = append(filtered, physicalMachines[index])
	}
	return filtered, nil
}
