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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KindTimeChaos is the kind for time chaos
const KindTimeChaos = "TimeChaos"

func init() {
	all.register(KindTimeChaos, &ChaosKind{
		Chaos:     &TimeChaos{},
		ChaosList: &TimeChaosList{},
	})
}

// +kubebuilder:object:root=true

// TimeChaos is the Schema for the timechaos API
type TimeChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a time chaos experiment
	Spec TimeChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the time chaos experiment
	Status TimeChaosStatus `json:"status"`
}

// TimeChaosSpec defines the desired state of TimeChaos
type TimeChaosSpec struct {
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

	// TimeOffset defines the delta time of injected program. It's a possibly signed sequence of decimal numbers, such as
	// "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	TimeOffset string `json:"timeOffset"`

	// ClockIds defines all affected clock id
	// All available options are ["CLOCK_REALTIME","CLOCK_MONOTONIC","CLOCK_PROCESS_CPUTIME_ID","CLOCK_THREAD_CPUTIME_ID",
	// "CLOCK_MONOTONIC_RAW","CLOCK_REALTIME_COARSE","CLOCK_MONOTONIC_COARSE","CLOCK_BOOTTIME","CLOCK_REALTIME_ALARM",
	// "CLOCK_BOOTTIME_ALARM"]
	// Default value is ["CLOCK_REALTIME"]
	ClockIds []string `json:"clockIds,omitempty"`

	// ContainerName indicates the name of affected container.
	// If not set, all containers will be injected
	// +optional
	ContainerNames []string `json:"containerNames,omitempty"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`
}

// SetDefaultValue will set default value for empty fields
func (in *TimeChaos) SetDefaultValue() {
	in.Spec.DefaultClockIds()
}

// DefaultClockIds will set default value for empty ClockIds fields
func (in *TimeChaosSpec) DefaultClockIds() {
	if in.ClockIds == nil || len(in.ClockIds) == 0 {
		in.ClockIds = []string{"CLOCK_REALTIME"}
	}
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *TimeChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *TimeChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *TimeChaosSpec) GetValue() string {
	return in.Value
}

// TimeChaosStatus defines the observed state of TimeChaos
type TimeChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// GetDuration gets the duration of TimeChaos
func (in *TimeChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetNextStart gets NextStart field of TimeChaos
func (in *TimeChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

// SetNextStart sets NextStart field of TimeChaos
func (in *TimeChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

// GetNextRecover get NextRecover field of TimeChaos
func (in *TimeChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

// SetNextRecover sets NextRecover field of TimeChaos
func (in *TimeChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextRecover = nil
		return
	}

	if in.Status.Scheduler.NextRecover == nil {
		in.Status.Scheduler.NextRecover = &metav1.Time{}
	}
	in.Status.Scheduler.NextRecover.Time = t
}

// GetScheduler returns the scheduler of TimeChaos
func (in *TimeChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetStatus returns the status of TimeChaos
func (in *TimeChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *TimeChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// GetPause returns the annotation when the chaos needs to be paused
func (in *TimeChaos) GetPause() string {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] == "" {
		return ""
	}
	return in.Annotations[PauseAnnotationKey]
}

// SetPause set the pausetime of annotation. Use for empty pausetime for now.
func (in *TimeChaos) SetPause(s string) {
	in.Annotations[PauseAnnotationKey] = s
}

// GetChaos returns a chaos instance
func (in *TimeChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindTimeChaos,
		StartTime: in.CreationTimestamp.Time,
		Action:    "",
		Status:    string(in.GetStatus().Experiment.Phase),
		UID:       string(in.UID),
	}
	if in.Spec.Duration != nil {
		instance.Duration = *in.Spec.Duration
	}
	if in.DeletionTimestamp != nil {
		instance.EndTime = in.DeletionTimestamp.Time
	}
	return instance
}

// +kubebuilder:object:root=true

// TimeChaosList contains a list of TimeChaos
type TimeChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TimeChaos `json:"items"`
}

// ListChaos returns a list of time chaos
func (in *TimeChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

func init() {
	SchemeBuilder.Register(&TimeChaos{}, &TimeChaosList{})
}
