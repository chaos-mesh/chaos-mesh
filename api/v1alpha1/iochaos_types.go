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

// IOChaos is the Schema for the iochaos API
type IOChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IOChaosSpec   `json:"spec,omitempty"`
	Status IOChaosStatus `json:"status,omitempty"`
}

// IOChaosSpec defines the desired state of IOChaos
type IOChaosSpec struct {
	ContainerSelector `json:",inline"`

	// Action defines the specific pod chaos action.
	// Supported action: latency / fault / attrOverride / mistake
	// +kubebuilder:validation:Enum=latency;fault;attrOverride;mistake
	Action IOChaosType `json:"action"`

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

	// Mistake defines what types of incorrectness are injected to IO operations
	// +optional
	Mistake *MistakeSpec `json:"mistake,omitempty"`

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

	// Duration represents the duration of the chaos action.
	// It is required when the action is `PodFailureAction`.
	// A duration string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Duration *string `json:"duration,omitempty"`
}

// IOChaosStatus defines the observed state of IOChaos
type IOChaosStatus struct {
	ChaosStatus `json:",inline"`

	// Instances always specifies podiochaos generation or empty
	// +optional
	Instances map[string]int64 `json:"instances,omitempty"`
}

func (obj *IOChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.ContainerSelector,
	}
}

func (obj *IOChaos) GetCustomStatus() interface{} {
	return &obj.Status.Instances
}
