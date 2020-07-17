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
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// DefaultNamespace set the namespace of chaos object as the default namespace selector if namespaces not set
func (in *SelectorSpec) DefaultNamespace(namespace string) {
	if len(in.Namespaces) == 0 {
		in.Namespaces = []string{namespace}
	}
}

// +kubebuilder:object:generate=false

// ChaosValidator describes the interface should be implemented in chaos
type ChaosValidator interface {
	webhook.Validator
	// Validate validates chaos object
	Validate() error
	// ValidateScheduler validates the scheduler and duration
	ValidateScheduler(spec *field.Path) field.ErrorList
	// ValidatePodMode validates the value with podmode
	ValidatePodMode(spec *field.Path) field.ErrorList
}
