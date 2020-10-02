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

package podnetworkchaosmanager

import (
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// PodNetworkTransaction represents a modification on podnetwork
type PodNetworkTransaction struct {
	Steps []Step
}

// Step represents a step of PodNetworkTransaction
type Step interface {
	// Apply will apply an action on podnetworkchaos
	Apply(chaos *v1alpha1.PodNetworkChaos) error
}

// Clear removes all resources with the same source
type Clear struct {
	Source string
}

// Apply runs this action
func (s *Clear) Apply(chaos *v1alpha1.PodNetworkChaos) error {
	ipsets := []v1alpha1.RawIPSet{}
	for _, ipset := range chaos.Spec.IPSets {
		if ipset.Source != s.Source {
			ipsets = append(ipsets, ipset)
		}
	}
	chaos.Spec.IPSets = ipsets

	chains := []v1alpha1.RawIptables{}
	for _, chain := range chaos.Spec.Iptables {
		if chain.Source != s.Source {
			chains = append(chains, chain)
		}
	}
	chaos.Spec.Iptables = chains

	qdiscs := []v1alpha1.RawTrafficControl{}
	for _, qdisc := range chaos.Spec.TrafficControls {
		if qdisc.Source != s.Source {
			qdiscs = append(qdiscs, qdisc)
		}
	}
	chaos.Spec.TrafficControls = qdiscs

	return nil
}

// Append adds an item to corresponding list in podnetworkchaos
type Append struct {
	Item interface{}
}

// Apply runs this action
func (a *Append) Apply(chaos *v1alpha1.PodNetworkChaos) error {
	switch item := a.Item.(type) {
	case v1alpha1.RawIPSet:
		chaos.Spec.IPSets = append(chaos.Spec.IPSets, item)
	case v1alpha1.RawIptables:
		chaos.Spec.Iptables = append(chaos.Spec.Iptables, item)
	case v1alpha1.RawTrafficControl:
		chaos.Spec.TrafficControls = append(chaos.Spec.TrafficControls, item)
	default:
		return fmt.Errorf("unknown type of item")
	}

	return nil
}

// Clear will clear all related items in podnetworkchaos
func (t *PodNetworkTransaction) Clear(source string) {
	t.Steps = append(t.Steps, &Clear{
		Source: source,
	})
}

// Append adds an item to corresponding list in podnetworkchaos
func (t *PodNetworkTransaction) Append(item interface{}) error {
	switch item.(type) {
	case v1alpha1.RawIPSet, v1alpha1.RawIptables, v1alpha1.RawTrafficControl:
		t.Steps = append(t.Steps, &Append{
			Item: item,
		})
		return nil
	default:
		return fmt.Errorf("unknown type of item")
	}
}

// Apply runs every step on the chaos
func (t *PodNetworkTransaction) Apply(chaos *v1alpha1.PodNetworkChaos) error {
	for _, s := range t.Steps {
		err := s.Apply(chaos)
		if err != nil {
			return err
		}
	}

	return nil
}
