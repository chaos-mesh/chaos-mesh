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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DNSChaosAction represents the chaos action about DNS.
type DNSChaosAction string

const (
	// ErrorAction represents get error when send DNS request.
	ErrorAction DNSChaosAction = "error"

	// RandomAction represents get random IP when send DNS request.
	RandomAction DNSChaosAction = "random"
)

// DNSChaosScope is the scope which the DNS chaos works.
type DNSChaosScope string

const (
	// OuterScope represents DNS chaos only works on the inner host in Kubernetes cluster
	OuterScope DNSChaosScope = "outer"

	// InnerScope represents DNS chaos only works on the outer host of Kubernetes cluster
	InnerScope DNSChaosScope = "inner"

	// AllScope represents DNS chaos works on host
	AllScope DNSChaosScope = "all"

	// TODO: maybe we can support set the RegExp for the host
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +chaos-mesh:base

// DNSChaos is the Schema for the networkchaos API
type DNSChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec DNSChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status DNSChaosStatus `json:"status"`
}

// DNSChaosSpec defines the desired state of DNSChaos
type DNSChaosSpec struct {
	// Action defines the specific DNS chaos action.
	// Supported action: error, random
	// Default action: error
	// +kubebuilder:validation:Enum=error;random
	Action DNSChaosAction `json:"action"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	// +kubebuilder:validation:Enum=one;all;fixed;fixed-percent;random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the percent of pods the server can do chaos action.
	// If `RandomMaxPercentPodMod`, provide a number from 0-100 to specify the max percent of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about network.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Action defines the scope which the DNS chaos works.
	// Supported action: outer, inner, all
	// Default action: outer
	// +kubebuilder:validation:Enum=outer;inner;all
	Scope DNSChaosScope `json:"scope"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *DNSChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *DNSChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *DNSChaosSpec) GetValue() string {
	return in.Value
}

// DNSChaosStatus defines the observed state of DNSChaos
type DNSChaosStatus struct {
	ChaosStatus `json:",inline"`
}
