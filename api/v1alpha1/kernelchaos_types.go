// Copyright 2020 PingCAP, Inc.
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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// KernelChaos is the Schema for the kernelchaos API
type KernelChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a kernel chaos experiment
	Spec KernelChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the kernel chaos experiment
	Status KernelChaosStatus `json:"status"`
}

// KernelChaosSpec defines the desired state of KernelChaos
type KernelChaosSpec struct {
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

	// FailKernRequest defines the request of kernel injection
	FailKernRequest FailKernRequest `json:"failKernRequest"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Next time when this action will be applied again
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *KernelChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *KernelChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *KernelChaosSpec) GetValue() string {
	return in.Value
}

// FailKernRequest defines the injection conditions
type FailKernRequest struct {
	FailType    int32    `json:"failtype"`
	Headers     []string `json:"headers,omitempty"`
	Callchain   []Frame  `json:"callchain,omitempty"`
	Probability uint32   `json:"probability,omitempty"`
	Times       uint32   `json:"times,omitempty"`
}

// Frame defines the function signature and predicate in function's body
type Frame struct {
	Funcname   string `json:"funcname,omitempty"`
	Parameters string `json:"parameters,omitempty"`
	Predicate  string `json:"predicate,omitempty"`
}

// KernelChaosStatus defines the observed state of KernelChaos
type KernelChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// GetDuration gets the duration of KernelChaos
func (in *KernelChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetNextStart gets NextStart field of KernelChaos
func (in *KernelChaos) GetNextStart() time.Time {
	if in.Spec.NextStart == nil {
		return time.Time{}
	}
	return in.Spec.NextStart.Time
}

// SetNextStart sets NextStart field of KernelChaos
func (in *KernelChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Spec.NextStart = nil
		return
	}

	if in.Spec.NextStart == nil {
		in.Spec.NextStart = &metav1.Time{}
	}
	in.Spec.NextStart.Time = t
}

// GetNextRecover get NextRecover field of KernelChaos
func (in *KernelChaos) GetNextRecover() time.Time {
	if in.Spec.NextRecover == nil {
		return time.Time{}
	}
	return in.Spec.NextRecover.Time
}

// SetNextRecover sets NextRecover field of KernelChaos
func (in *KernelChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Spec.NextRecover = nil
		return
	}

	if in.Spec.NextRecover == nil {
		in.Spec.NextRecover = &metav1.Time{}
	}
	in.Spec.NextRecover.Time = t
}

// GetScheduler returns the scheduler of KernelChaos
func (in *KernelChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetStatus returns the status of KernelChaos
func (in *KernelChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *KernelChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// +kubebuilder:object:root=true

// KernelChaosList contains a list of KernelChaos
type KernelChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KernelChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KernelChaos{}, &KernelChaosList{})
}
