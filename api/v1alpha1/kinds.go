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

package v1alpha1

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:generate=false

// ChaosKindMap defines a map including all chaos kinds.
type chaosKindMap struct {
	sync.RWMutex
	kinds map[string]*ChaosKind
}

func (c *chaosKindMap) register(name string, kind *ChaosKind) {
	c.Lock()
	defer c.Unlock()
	c.kinds[name] = kind
}

func (c *chaosKindMap) clone() map[string]*ChaosKind {
	c.RLock()
	defer c.RUnlock()

	out := make(map[string]*ChaosKind)
	for key, kind := range c.kinds {
		out[key] = &ChaosKind{
			Chaos:     kind.Chaos,
			ChaosList: kind.ChaosList,
		}
	}

	return out
}

// AllKinds returns all chaos kinds.
func AllKinds() map[string]*ChaosKind {
	return all.clone()
}

// all is a ChaosKindMap instance.
var all = &chaosKindMap{
	kinds: make(map[string]*ChaosKind),
}

// +kubebuilder:object:generate=false

// ChaosKind includes one kind of chaos and its list type
type ChaosKind struct {
	Chaos runtime.Object
	ChaosList
}

// AllKinds returns all chaos kinds.
func AllScheduleItemKinds() map[string]*ChaosKind {
	return allScheduleItem.clone()
}

// all is a ChaosKindMap instance.
var allScheduleItem = &chaosKindMap{
	kinds: make(map[string]*ChaosKind),
}
