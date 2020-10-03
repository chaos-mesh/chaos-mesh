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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KindIOChaos is the kind for io chaos
const KindIOChaos = "IoChaos"

func init() {
	all.register(KindIOChaos, &ChaosKind{
		Chaos:     &IoChaos{},
		ChaosList: &IoChaosList{},
	})
}

// IOLayer represents the layer of I/O system.
type IOLayer string

// IoChaosSpec defines the desired state of IoChaos
type IoChaosSpec struct {
	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	// +kubebuilder:validation:Enum=one;all;fixed;fixed-percent;random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the percent of pods the server can do chaos action.
	// IF `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the max percent of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Action defines the specific pod chaos action.
	// Supported action: latency / fault / attrOverride
	// +kubebuilder:validation:Enum=latency;fault;attrOverride
	Action IoChaosType `json:"action"`

	// Delay defines the value of I/O chaos action delay.
	// A delay string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Delay string `json:"delay,omitempty"`

	// Errno defines the error code that returned by I/O action.
	// refer to: https://www-numi.fnal.gov/offline_software/srt_public_context/WebDocs/Errors/unix_system_errors.html
	// +optional
	Errno uint32 `json:"errno,omitempty"`

	// Attr defines the overrided attribution
	// +optional
	Attr *AttrOverrideSpec `json:"attr,omitempty"`

	// Path defines the path of files for injecting I/O chaos action.
	// +optional
	Path string `json:"path,omitempty"`

	// Methods defines the I/O methods for injecting I/O chaos action.
	// default: all I/O methods.
	// +optional
	Methods []IoMethod `json:"methods,omitempty"`

	// Percent defines the percentage of injection errors and provides a number from 0-100.
	// default: 100.
	// +optional
	Percent int `json:"percent,omitempty"`

	// VolumePath represents the mount path of injected volume
	VolumePath string `json:"volumePath"`

	// Scheduler defines some schedule rules to
	// control the running time of the chaos experiment about pods.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Duration represents the duration of the chaos action.
	// It is required when the action is `PodFailureAction`.
	// A duration string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Duration *string `json:"duration,omitempty"`
}

func (in *IoChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

func (in *IoChaosSpec) GetMode() PodMode {
	return in.Mode
}

func (in *IoChaosSpec) GetValue() string {
	return in.Value
}

// IoChaosStatus defines the observed state of IoChaos
type IoChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// IoChaos is the Schema for the iochaos API
type IoChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IoChaosSpec   `json:"spec,omitempty"`
	Status IoChaosStatus `json:"status,omitempty"`
}

func (in *IoChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *IoChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// GetPause returns the annotation when the chaos needs to be paused
func (in *IoChaos) GetPause() string {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] == "" {
		return ""
	}
	return in.Annotations[PauseAnnotationKey]
}

// GetDuration would return the duration for chaos
func (in *IoChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *IoChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *IoChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *IoChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *IoChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextRecover = nil
		return
	}

	if in.Status.Scheduler.NextRecover == nil {
		in.Status.Scheduler.NextRecover = &metav1.Time{}
	}
	in.Status.Scheduler.NextRecover.Time = t
}

// GetScheduler would return the scheduler for chaos
func (in *IoChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos returns a chaos instance
func (in *IoChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindIOChaos,
		StartTime: in.CreationTimestamp.Time,
		Action:    string(in.Spec.Action),
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

// IoChaosList contains a list of IoChaos
type IoChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IoChaos `json:"items"`
}

// ListChaos returns a list of io chaos
func (in *IoChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

func init() {
	SchemeBuilder.Register(&IoChaos{}, &IoChaosList{})
}
