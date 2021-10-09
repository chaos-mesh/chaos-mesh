// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package physicalmachine

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = ctrl.Log.WithName("physicalmachineselector")

type Option struct {
	ClusterScoped         bool
	TargetNamespace       string
	EnableFilterNamespace bool
}

type SelectImpl struct {
	c client.Client
	r client.Reader

	Option
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

func New() *SelectImpl {
	return &SelectImpl{}
}

func SelectPhysicalMachines(ctx context.Context, c client.Client, r client.Reader,
	selector v1alpha1.PhysicalMachineSelectorSpec,
	clusterScoped bool, targetNamespace string, enableFilterNamespace bool) ([]v1alpha1.PhysicalMachine, error) {
	var physicalMachines []v1alpha1.PhysicalMachine

	if len(selector.PhysicalMachines) > 0 {
		for ns, names := range selector.PhysicalMachines {
			if !clusterScoped {
				if targetNamespace != ns {
					log.Info("skip namespace because ns is out of scope within namespace scoped mode", "namespace", ns)
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

	// TODO refactor
	return physicalMachines, nil
}
