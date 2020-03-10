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

// StressChaos is the Schema for the timechaos API
type StressChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a time chaos experiment
	Spec StressChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the time chaos experiment
	Status StressChaosStatus `json:"status"`
}

// StressChaosSpec defines the desired state of StressChaos
type StressChaosSpec struct {
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

	// Stressors defines the delta time of injected program
	Stressors Stressors `json:"stressors"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	Scheduler *SchedulerSpec `json:"scheduler"`

	// Next time when this action will be applied again
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
}


// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *StressChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *StressChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *StressChaosSpec) GetValue() string {
	return in.Value
}

// Stressors defines the delta time of injected program
// As `clock_gettime` return a struct contains two field: `tv_sec` and `tv_nsec`.
// `Sec` is the offset of seconds, corresponding to `tv_sec` field.
// `NSec` is the offset of nanoseconds, corresponding to `tv_nsec` field.
type Stressors struct {
	Sec  int64 `json:"sec"`
	NSec int64 `json:"nsec"`
}

// StressChaosStatus defines the observed state of StressChaos
type StressChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// GetDuration gets the duration of StressChaos
func (in *StressChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetNextStart gets NextStart field of StressChaos
func (in *StressChaos) GetNextStart() time.Time {
	if in.Spec.NextStart == nil {
		return time.Time{}
	}
	return in.Spec.NextStart.Time
}

// SetNextStart sets NextStart field of StressChaos
func (in *StressChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Spec.NextStart = nil
		return
	}

	if in.Spec.NextStart == nil {
		in.Spec.NextStart = &metav1.Time{}
	}
	in.Spec.NextStart.Time = t
}

// GetNextRecover get NextRecover field of StressChaos
func (in *StressChaos) GetNextRecover() time.Time {
	if in.Spec.NextRecover == nil {
		return time.Time{}
	}
	return in.Spec.NextRecover.Time
}

// SetNextRecover sets NextRecover field of StressChaos
func (in *StressChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Spec.NextRecover = nil
		return
	}

	if in.Spec.NextRecover == nil {
		in.Spec.NextRecover = &metav1.Time{}
	}
	in.Spec.NextRecover.Time = t
}

// GetScheduler returns the scheduler of StressChaos
func (in *StressChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetStatus returns the status of StressChaos
func (in *StressChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *StressChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// Validate describe the timechaos validation logic
func (in *StressChaos) Validate() (bool, string, error) {
	if in.Spec.Duration != nil && in.Spec.Scheduler != nil {
		return true, "", nil
	} else if in.Spec.Duration == nil && in.Spec.Scheduler == nil {
		return true, "", nil
	}
	return false, invalidConfigurationMsg, nil
}

// +kubebuilder:object:root=true

// StressChaosList contains a list of StressChaos
type StressChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StressChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StressChaos{}, &StressChaosList{})
}
