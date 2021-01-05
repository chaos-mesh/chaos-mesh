// Copyright 2020 Chaos Mesh Authors.
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

package actor

import (
	"k8s.io/apimachinery/pkg/types"

	chaosmeshv1alph1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// FIXME: reconsider about actor-playground things

type Actor interface {
	PlayOn(pg Playground) error
}

type Playground interface {
	CreateNetworkChaos(networkChaos chaosmeshv1alph1.NetworkChaos) error
	DeleteNetworkChaos(name types.NamespacedName) error
}

type CreateNetworkChaosActor struct {
	chaosToCreate chaosmeshv1alph1.NetworkChaos
}

func NewCreateNetworkChaosActor(chaosToCreate chaosmeshv1alph1.NetworkChaos) *CreateNetworkChaosActor {
	return &CreateNetworkChaosActor{chaosToCreate: chaosToCreate}
}

func (it *CreateNetworkChaosActor) PlayOn(pg Playground) error {
	return pg.CreateNetworkChaos(it.chaosToCreate)
}

type DeleteNetworkChaosActor struct {
	name types.NamespacedName
}

func NewDeleteNetworkChaosActor(name types.NamespacedName) *DeleteNetworkChaosActor {
	return &DeleteNetworkChaosActor{name: name}
}

func (it *DeleteNetworkChaosActor) PlayOn(pg Playground) error {
	return pg.DeleteNetworkChaos(it.name)
}
