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

package podhttpchaosmanager

import (
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// PodHttpTransaction represents a modification on podhttpchaos
type PodHttpTransaction struct {
	Steps []Step
}

// Step represents a step of PodHttpTransaction
type Step interface {
	// Apply will apply an action on podnetworkchaos
	Apply(chaos *v1alpha1.PodHttpChaos) error
}

// Clear removes all resources with the same source
type Clear struct {
	Source string
}

// Apply runs this action
func (s *Clear) Apply(chaos *v1alpha1.PodHttpChaos) error {
	rules := []v1alpha1.PodHttpChaosRule{}
	for _, rule := range chaos.Spec.Rules {
		if rule.Source != s.Source {
			rules = append(rules, rule)
		}
	}
	chaos.Spec.Rules = rules
	return nil
}

// Append adds an item to corresponding list in podhttpchaos
type Append struct {
	Item interface{}
}

// Apply runs this action
func (a *Append) Apply(chaos *v1alpha1.PodHttpChaos) error {
	switch item := a.Item.(type) {
	case v1alpha1.PodHttpChaosRule:
		chaos.Spec.Rules = append(chaos.Spec.Rules, item)
	default:
		return fmt.Errorf("unknown type of item")
	}

	return nil
}

// Clear will clear all related items in podhttpchaos
func (t *PodHttpTransaction) Clear(source string) {
	t.Steps = append(t.Steps, &Clear{
		Source: source,
	})
}

// Append adds an item to corresponding list in podnetworkchaos
func (t *PodHttpTransaction) Append(item interface{}) error {
	switch item.(type) {
	case v1alpha1.PodHttpChaosRule:
		t.Steps = append(t.Steps, &Append{
			Item: item,
		})
		return nil
	default:
		return fmt.Errorf("unknown type of item")
	}
}

// Apply runs every step on the chaos
func (t *PodHttpTransaction) Apply(chaos *v1alpha1.PodHttpChaos) error {
	for _, s := range t.Steps {
		err := s.Apply(chaos)
		if err != nil {
			return err
		}
	}

	return nil
}
