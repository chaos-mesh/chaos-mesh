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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IOChaosAction represents the chaos action about I/O action.
type IOChaosAction string

const (
	IODelayAction IOChaosAction = "delay"
	IOErrnoAction               = "errno"
	IOMixedAction               = "mixed"
)

const (
	// TODO: add config file
	WebhookNamespaceLabelKey    = "admission-webhook"
	WebhookNamespaceLabelValue  = "enabled"
	WebhookPodAnnotationKey     = "admission-webhook.pingcap.com/request"
	WebhookInitPodAnnotationKey = "admission-webhook.pingcap.com/init-request"
)

// IOLayer represents the layer of I/O system.
type IOLayer string

const (
	FileSystemLayer = "fs"
	BlockLayer      = "block"
	DeviceLayer     = "device"
)

const (
	DefaultChaosfsAddr = ":65534"
)

// IoChaosSpec defines the desired state of IoChaos
type IoChaosSpec struct {
	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Scheduler defines some schedule rules to
	// control the running time of the chaos experiment about pods.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Action defines the specific pod chaos action.
	// Supported action: delay / errno / mixed
	// Default action: delay
	Action IOChaosAction `json:"action"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the max % of pods the server can do chaos action.
	// IF `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the % of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Duration represents the duration of the chaos action.
	// It is required when the action is `PodFailureAction`.
	// A duration string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Layer represents the layer of the I/O action.
	// Supported value: fs.
	// Default layer: fs
	Layer IOLayer `json:"layer"`

	// Delay defines the value of I/O chaos action delay.
	// A delay string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	//
	// If `Delay` is empty, the operator will generate a value for it randomly.
	// +optional
	Delay string `json:"delay,omitempty"`

	// Errno defines the error code that returned by I/O action.
	// refer to: https://www-numi.fnal.gov/offline_software/srt_public_context/WebDocs/Errors/unix_system_errors.html
	//
	// If `Errno` is empty, the operator will generate a error code for it randomly.
	// +optional
	Errno string `json:"errno,omitempty"`

	// Percent defines the percentage of injection errors and provides a number from 0-100.
	// default: 100.
	// +optional
	Percent string `json:"percent,omitempty"`

	// Path defines the path of files for injecting I/O chaos action.
	// +optional
	Path string `json:"path,omitempty"`

	// Methods defines the I/O methods for injecting I/O chaos action.
	// default: all I/O methods.
	// +optional
	Methods []string `json:"methods,omitempty"`

	// Addr defines the address for sidecar container.
	// +optional
	Addr string `json:"addr,omitempty"`

	// ConfigName defines the config name which used to inject pod.
	// +required
	ConfigName string `json:"configName"`

	// Next time when this action will be applied again.
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered.
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
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

// IsPaused returns whether this resource is paused
func (in *IoChaos) IsPaused() bool {
	return false
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
	if in.Spec.NextStart == nil {
		return time.Time{}
	}
	return in.Spec.NextStart.Time
}

func (in *IoChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Spec.NextStart = nil
		return
	}

	if in.Spec.NextStart == nil {
		in.Spec.NextStart = &metav1.Time{}
	}
	in.Spec.NextStart.Time = t
}

func (in *IoChaos) GetNextRecover() time.Time {
	if in.Spec.NextRecover == nil {
		return time.Time{}
	}
	return in.Spec.NextRecover.Time
}

func (in *IoChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Spec.NextRecover = nil
		return
	}

	if in.Spec.NextRecover == nil {
		in.Spec.NextRecover = &metav1.Time{}
	}
	in.Spec.NextRecover.Time = t
}

// GetScheduler would return the scheduler for chaos
func (in *IoChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// +kubebuilder:object:root=true

// IoChaosList contains a list of IoChaos
type IoChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IoChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IoChaos{}, &IoChaosList{})
}
