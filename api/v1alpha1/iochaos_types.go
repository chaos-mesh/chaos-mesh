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

// IoChaos is the Schema for the iochaos API
type IoChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IoChaosSpec   `json:"spec,omitempty"`
	Status IoChaosStatus `json:"status,omitempty"`
}

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
	Methods []IoMethod `json:"methods,omitempty" faker:"ioMethods"`

	// Percent defines the percentage of injection errors and provides a number from 0-100.
	// default: 100.
	// +optional
	Percent int `json:"percent,omitempty"`

	// VolumePath represents the mount path of injected volume
	VolumePath string `json:"volumePath"`

	// ContainerName indicates the target container to inject iochaos in
	// +optional
	ContainerName *string `json:"containerName,omitempty"`

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
