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
	"k8s.io/apimachinery/pkg/types"
)

type SelectImpl struct{}

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
	return []*PhysicalMachine{physicalMachineSelector}, nil
}

func New() *SelectImpl {
	return &SelectImpl{}
}
