// Copyright 2019 Chaos Mesh Authors.
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

// +kubebuilder:object:root=true
// +chaos-mesh:base

// NetworkChaos is the Schema for the networkchaos API
type NetworkChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec NetworkChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status NetworkChaosStatus `json:"status"`
}

// NetworkChaosAction represents the chaos action about network.
type NetworkChaosAction string

const (
	// NetemAction is a combination of several chaos actions i.e. delay, loss, duplicate, corrupt.
	// When using this action multiple specs are merged into one Netem RPC and sends to chaos daemon.
	NetemAction NetworkChaosAction = "netem"

	// DelayAction represents the chaos action of adding delay on pods.
	DelayAction NetworkChaosAction = "delay"

	// LossAction represents the chaos action of losing packets on pods.
	LossAction NetworkChaosAction = "loss"

	// DuplicateAction represents the chaos action of duplicating packets on pods.
	DuplicateAction NetworkChaosAction = "duplicate"

	// CorruptAction represents the chaos action of corrupting packets on pods.
	CorruptAction NetworkChaosAction = "corrupt"

	// PartitionAction represents the chaos action of network partition of pods.
	PartitionAction NetworkChaosAction = "partition"

	// BandwidthAction represents the chaos action of network bandwidth of pods.
	BandwidthAction NetworkChaosAction = "bandwidth"
)

// Direction represents traffic direction from source to target,
// it could be netem, delay, loss, duplicate, corrupt or partition,
// check comments below for detail direction flow.
type Direction string

const (
	// To represents network packet from source to target
	To Direction = "to"

	// From represents network packet to source from target
	From Direction = "from"

	// Both represents both directions
	Both Direction = "both"
)

// Target represents network partition and netem action target.
type Target struct {
	// TargetSelector defines the target selector
	TargetSelector SelectorSpec `json:"selector" mapstructure:"selector"`

	// TargetMode defines the target selector mode
	// +kubebuilder:validation:Enum=one;all;fixed;fixed-percent;random-max-percent;""
	TargetMode PodMode `json:"mode" mapstructure:"mode"`

	// TargetValue is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the percent of pods the server can do chaos action.
	// If `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the max percent of pods to do chaos action
	// +optional
	TargetValue string `json:"value" mapstructure:"value"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *Target) GetSelector() SelectorSpec {
	return in.TargetSelector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *Target) GetMode() PodMode {
	return in.TargetMode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *Target) GetValue() string {
	return in.TargetValue
}

// NetworkChaosSpec defines the desired state of NetworkChaos
type NetworkChaosSpec struct {
	// Action defines the specific network chaos action.
	// Supported action: partition, netem, delay, loss, duplicate, corrupt
	// Default action: delay
	// +kubebuilder:validation:Enum=netem;delay;loss;duplicate;corrupt;partition;bandwidth
	Action NetworkChaosAction `json:"action"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	// +kubebuilder:validation:Enum=one;all;fixed;fixed-percent;random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the percent of pods the server can do chaos action.
	// If `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the max percent of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about network.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// TcParameter represents the traffic control definition
	TcParameter `json:",inline"`

	// Direction represents the direction, this applies on netem and network partition action
	// +optional
	// +kubebuilder:validation:Enum=to;from;both;""
	Direction Direction `json:"direction,omitempty"`

	// Target represents network target, this applies on netem and network partition action
	// +optional
	Target *Target `json:"target,omitempty"`

	// ExternalTargets represents network targets outside k8s
	// +optional
	ExternalTargets []string `json:"externalTargets,omitempty"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *NetworkChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *NetworkChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *NetworkChaosSpec) GetValue() string {
	return in.Value
}

// NetworkChaosStatus defines the observed state of NetworkChaos
type NetworkChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// DelaySpec defines detail of a delay action
type DelaySpec struct {
	Latency     string       `json:"latency"`
	Correlation string       `json:"correlation,omitempty"`
	Jitter      string       `json:"jitter,omitempty"`
	Reorder     *ReorderSpec `json:"reorder,omitempty"`
}

// LossSpec defines detail of a loss action
type LossSpec struct {
	Loss        string `json:"loss"`
	Correlation string `json:"correlation"`
}

// DuplicateSpec defines detail of a duplicate action
type DuplicateSpec struct {
	Duplicate   string `json:"duplicate"`
	Correlation string `json:"correlation"`
}

// CorruptSpec defines detail of a corrupt action
type CorruptSpec struct {
	Corrupt     string `json:"corrupt"`
	Correlation string `json:"correlation"`
}

// BandwidthSpec defines detail of bandwidth limit.
type BandwidthSpec struct {
	// Rate is the speed knob. Allows bps, kbps, mbps, gbps, tbps unit. bps means bytes per second.
	Rate string `json:"rate"`
	// Limit is the number of bytes that can be queued waiting for tokens to become available.
	// +kubebuilder:validation:Minimum=1
	Limit uint32 `json:"limit"`
	// Buffer is the maximum amount of bytes that tokens can be available for instantaneously.
	// +kubebuilder:validation:Minimum=1
	Buffer uint32 `json:"buffer"`
	// Peakrate is the maximum depletion rate of the bucket.
	// The peakrate does not need to be set, it is only necessary
	// if perfect millisecond timescale shaping is required.
	// +optional
	// +kubebuilder:validation:Minimum=0
	Peakrate *uint64 `json:"peakrate,omitempty"`
	// Minburst specifies the size of the peakrate bucket. For perfect
	// accuracy, should be set to the MTU of the interface.  If a
	// peakrate is needed, but some burstiness is acceptable, this
	// size can be raised. A 3000 byte minburst allows around 3mbit/s
	// of peakrate, given 1000 byte packets.
	// +optional
	// +kubebuilder:validation:Minimum=0
	Minburst *uint32 `json:"minburst,omitempty"`
}

// ReorderSpec defines details of packet reorder.
type ReorderSpec struct {
	Reorder     string `json:"reorder"`
	Correlation string `json:"correlation"`
	Gap         int    `json:"gap"`
}
