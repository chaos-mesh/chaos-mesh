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

package podiochaosmanager

import (
	"fmt"

	v1alpha1 "github.com/chaos-mesh/api"
)

// PodIoTransaction represents a modification on podnetwork
type PodIoTransaction struct {
	Steps []Step
}

// Step represents a step of PodIoTransaction
type Step interface {
	// Apply will apply an action on podnetworkchaos
	Apply(chaos *v1alpha1.PodIoChaos) error
}

// Clear removes all resources with the same source
type Clear struct {
	Source string
}

// Apply runs this action
func (s *Clear) Apply(chaos *v1alpha1.PodIoChaos) error {
	actions := []v1alpha1.IoChaosAction{}
	for _, action := range chaos.Spec.Actions {
		if action.Source != s.Source {
			actions = append(actions, action)
		}
	}
	chaos.Spec.Actions = actions

	return nil
}

// Append adds an item to corresponding list in podnetworkchaos
type Append struct {
	Item interface{}
}

// Apply runs this action
func (a *Append) Apply(chaos *v1alpha1.PodIoChaos) error {
	switch item := a.Item.(type) {
	case v1alpha1.IoChaosAction:
		chaos.Spec.Actions = append(chaos.Spec.Actions, item)
	default:
		return fmt.Errorf("unknown type of item")
	}

	return nil
}

// SetContainer sets the container field of podiochaos
type SetContainer struct {
	Container string
}

// Apply runs this action
func (s *SetContainer) Apply(chaos *v1alpha1.PodIoChaos) error {
	chaos.Spec.Container = &s.Container

	return nil
}

// SetVolumePath sets the volumePath field of podiochaos
type SetVolumePath struct {
	Path string
}

// Apply runs this action
func (s *SetVolumePath) Apply(chaos *v1alpha1.PodIoChaos) error {
	chaos.Spec.VolumeMountPath = s.Path

	return nil
}

// Clear will clear all related items in podnetworkchaos
func (t *PodIoTransaction) Clear(source string) {
	t.Steps = append(t.Steps, &Clear{
		Source: source,
	})
}

// Append adds an item to corresponding list in podnetworkchaos
func (t *PodIoTransaction) Append(item interface{}) error {
	switch item.(type) {
	case v1alpha1.IoChaosAction:
		t.Steps = append(t.Steps, &Append{
			Item: item,
		})
		return nil
	default:
		return fmt.Errorf("unknown type of item")
	}
}

// SetVolumePath sets the volumePath field of podiochaos
func (t *PodIoTransaction) SetVolumePath(path string) error {
	t.Steps = append(t.Steps, &SetVolumePath{
		Path: path,
	})

	return nil
}

func (t *PodIoTransaction) SetContainer(container string) error {
	t.Steps = append(t.Steps, &SetContainer{
		Container: container,
	})

	return nil
}

// Apply runs every step on the chaos
func (t *PodIoTransaction) Apply(chaos *v1alpha1.PodIoChaos) error {
	for _, s := range t.Steps {
		err := s.Apply(chaos)
		if err != nil {
			return err
		}
	}

	return nil
}
