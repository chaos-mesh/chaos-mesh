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
	chaosmeshv1alph1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// FIXME: reconsider about actor-playground things

type Actor interface {
	PlayOn(pg Playground) error
}

// TODO: multi type of chaos injection
type Playground interface {
	CreateNetworkChaos(networkChaos chaosmeshv1alph1.NetworkChaos) error
	DeleteNetworkChaos(namespace, name string) error
	CreatePodChaos(podChaos chaosmeshv1alph1.PodChaos) error
	DeletePodChaos(namespace, name string) error
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
	namespace string
	name      string
}

func NewDeleteNetworkChaosActor(namespace string, name string) *DeleteNetworkChaosActor {
	return &DeleteNetworkChaosActor{namespace: namespace, name: name}
}

func (it *DeleteNetworkChaosActor) PlayOn(pg Playground) error {
	return pg.DeleteNetworkChaos(it.namespace, it.name)
}

type CreatePodChaosActor struct {
	chaosToCreate chaosmeshv1alph1.PodChaos
}

func NewCreatePodChaosActor(chaosToCreate chaosmeshv1alph1.PodChaos) *CreatePodChaosActor {
	return &CreatePodChaosActor{chaosToCreate: chaosToCreate}
}

func (it *CreatePodChaosActor) PlayOn(pg Playground) error {
	return pg.CreatePodChaos(it.chaosToCreate)
}

type DeletePodChaosActor struct {
	namespace string
	name      string
}

func NewDeletePodChaosActor(namespace string, name string) *DeletePodChaosActor {
	return &DeletePodChaosActor{namespace: namespace, name: name}
}

func (it *DeletePodChaosActor) PlayOn(pg Playground) error {
	return pg.DeletePodChaos(it.namespace, it.name)
}
