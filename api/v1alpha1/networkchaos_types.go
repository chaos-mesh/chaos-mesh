// Copyright 2019 PingCAP, Inc.
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
	"strconv"
	"time"

	chaosdaemon "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ChaosAction represents the chaos action about pods.
type NetworkChaosAction string

const (
	// DelayAction represents the chaos action of adding delay on pods.
	DelayAction NetworkChaosAction = "delay"

	// LossAction represents the chaos action of lossing packets on pods.
	LossAction NetworkChaosAction = "loss"

	// DuplicateAction represents the chaos action of duplicating packets on pods.
	DuplicateAction NetworkChaosAction = "duplicate"

	// CorruptAction represents the chaos action of corrupting packets on pods.
	CorruptAction NetworkChaosAction = "corrupt"

	// PartitionAction represents the chaos action of network partition of pods.
	PartitionAction NetworkChaosAction = "partition"
)

// PartitionDirection represents the block direction from source to target
type PartitionDirection string

const (
	// To represents block network packet from source to target
	To PartitionDirection = "to"

	// From represents block network packet to source from target
	From PartitionDirection = "from"

	// Both represents block both directions
	Both PartitionDirection = "both"
)

type PartitionTarget struct {
	// TargetSelector defines the partition target selector
	TargetSelector SelectorSpec `json:"selector"`

	// TargetMode defines the partition target selector mode
	TargetMode PodMode `json:"mode"`

	// TargetValue is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the max % of pods the server can do chaos action.
	// If `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the % of pods to do chaos action
	// +optional
	TargetValue string `json:"value"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (t *PartitionTarget) GetSelector() SelectorSpec {
	return t.TargetSelector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (t *PartitionTarget) GetMode() PodMode {
	return t.TargetMode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (t *PartitionTarget) GetValue() string {
	return t.TargetValue
}

// NetworkChaosSpec defines the desired state of NetworkChaos
type NetworkChaosSpec struct {
	// Action defines the specific network chaos action.
	// Supported action: delay
	// Default action: delay
	Action NetworkChaosAction `json:"action"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the max % of pods the server can do chaos action.
	// If `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the % of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Duration represents the duration of the chaos action
	Duration string `json:"duration"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about network.
	Scheduler SchedulerSpec `json:"scheduler"`

	// Delay represents the detail about delay action
	// +optional
	Delay *DelaySpec `json:"delay,omitempty"`

	// Loss represents the detail about loss action
	Loss *LossSpec `json:"loss,omitempty"`

	// DuplicateSpec represents the detail about loss action
	Duplicate *DuplicateSpec `json:"duplicate,omitempty"`

	// Corrupt represents the detail about loss action
	Corrupt *CorruptSpec `json:"corrupt,omitempty"`

	// Direction represents the partition direction
	// +optional
	Direction PartitionDirection `json:"direction"`

	// Target represents network partition target
	// +optional
	Target PartitionTarget `json:"target"`

	// Next time when this action will be applied again
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
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

// +kubebuilder:object:root=true

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

func (in *NetworkChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

func (in *NetworkChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

func (in *NetworkChaos) GetDuration() (time.Duration, error) {
	duration, err := time.ParseDuration(in.Spec.Duration)
	if err != nil {
		return time.Hour * 0, err
	}

	return duration, nil
}

func (in *NetworkChaos) GetNextStart() time.Time {
	if in.Spec.NextStart == nil {
		return time.Time{}
	}
	return in.Spec.NextStart.Time
}

func (in *NetworkChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Spec.NextStart = nil
		return
	}

	if in.Spec.NextStart == nil {
		in.Spec.NextStart = &metav1.Time{}
	}
	in.Spec.NextStart.Time = t
}

func (in *NetworkChaos) GetNextRecover() time.Time {
	if in.Spec.NextRecover == nil {
		return time.Time{}
	}
	return in.Spec.NextRecover.Time
}

func (in *NetworkChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Spec.NextRecover = nil
		return
	}

	if in.Spec.NextRecover == nil {
		in.Spec.NextRecover = &metav1.Time{}
	}
	in.Spec.NextRecover.Time = t
}

func (in *NetworkChaos) GetScheduler() SchedulerSpec {
	return in.Spec.Scheduler
}

// DelaySpec defines detail of a delay action
type DelaySpec struct {
	Latency     string `json:"latency"`
	Correlation string `json:"correlation"`
	Jitter      string `json:"jitter"`
}

func (delay *DelaySpec) ToNetem() (*chaosdaemon.Netem, error) {
	delayTime, err := time.ParseDuration(delay.Latency)
	if err != nil {
		return nil, err
	}
	jitter, err := time.ParseDuration(delay.Jitter)
	if err != nil {
		return nil, err
	}

	corr, err := strconv.ParseFloat(delay.Correlation, 32)
	if err != nil {
		return nil, err
	}

	return &chaosdaemon.Netem{
		Time:      uint32(delayTime.Nanoseconds() / 1e3),
		DelayCorr: float32(corr),
		Jitter:    uint32(jitter.Nanoseconds() / 1e3),
	}, nil
}

// LossSpec defines detail of a loss action
type LossSpec struct {
	Loss        string `json:"loss"`
	Correlation string `json:"correlation"`
}

func (loss *LossSpec) ToNetem() (*chaosdaemon.Netem, error) {
	lossPercentage, err := strconv.ParseFloat(loss.Loss, 32)
	if err != nil {
		return nil, err
	}

	corr, err := strconv.ParseFloat(loss.Correlation, 32)
	if err != nil {
		return nil, err
	}

	return &chaosdaemon.Netem{
		Loss:     float32(lossPercentage),
		LossCorr: float32(corr),
	}, nil
}

// DuplicateSpec defines detail of a duplicate action
type DuplicateSpec struct {
	Duplicate   string `json:"duplicate"`
	Correlation string `json:"correlation"`
}

func (duplicate *DuplicateSpec) ToNetem() (*chaosdaemon.Netem, error) {
	duplicatePercentage, err := strconv.ParseFloat(duplicate.Duplicate, 32)
	if err != nil {
		return nil, err
	}

	corr, err := strconv.ParseFloat(duplicate.Correlation, 32)
	if err != nil {
		return nil, err
	}

	return &chaosdaemon.Netem{
		Duplicate:     float32(duplicatePercentage),
		DuplicateCorr: float32(corr),
	}, nil
}

// CorruptSpec defines detail of a corrupt action
type CorruptSpec struct {
	Corrupt     string `json:"corrupt"`
	Correlation string `json:"correlation"`
}

func (corrupt *CorruptSpec) ToNetem() (*chaosdaemon.Netem, error) {
	corruptPercentage, err := strconv.ParseFloat(corrupt.Corrupt, 32)
	if err != nil {
		return nil, err
	}

	corr, err := strconv.ParseFloat(corrupt.Correlation, 32)
	if err != nil {
		return nil, err
	}

	return &chaosdaemon.Netem{
		Corrupt:     float32(corruptPercentage),
		CorruptCorr: float32(corr),
	}, nil
}

// +kubebuilder:object:root=true

// NetworkChaosList contains a list of NetworkChaos
type NetworkChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkChaos{}, &NetworkChaosList{})
}
