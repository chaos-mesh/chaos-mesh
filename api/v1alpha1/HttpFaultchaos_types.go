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
	"strconv"
)
const KindHttpFaultChaos = "HttpFaultChaos"

func init() {
	all.register(KindHttpFaultChaos, &ChaosKind{
		Chaos:     &HttpFaultChaos{},
		ChaosList: &HttpFaultChaosList{},
	})
}

// HttpFaultAction represents the chaos action about I/O action.
type HttpFaultAction string

const (
	HttpFaultDelayAction HttpFaultAction = "delay"
)


type HttpFaultSpec struct {
	// Action defines the specific pod chaos action.
	// Supported action: delay
	// Default action: delay
	// +kubebuilder:validation:Enum=delay
	Action HttpFaultAction `json:"action"`
	// Duration represents the duration of the chaos action.
	// It is required when the action is `PodFailureAction`.
	// A duration string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Duration *string `json:"duration,omitempty"`
	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`
	// Mode defines the mode to run chaos action.
	// Supported mode: one
	// +kubebuilder:validation:Enum=one
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the percent of pods the server can do chaos action.
	// IF `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the max percent of pods to do chaos action
	// +optional
	Value string `json:"value"`
}

func (in *HttpFaultSpec) GetMode() PodMode {
	return in.Mode
}

func (in *HttpFaultSpec) GetValue() string {
	return in.Value
}

func (in *HttpFaultSpec) GetSelector() SelectorSpec {
	return in.Selector
}

type HttpFaultChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// HttpFaultChaos is the Schema for the HttpFaultchaos API
type HttpFaultChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   HttpFaultSpec   `json:"spec,omitempty"`
	Status IoChaosStatus `json:"status,omitempty"`
}

func (in *HttpFaultChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *HttpFaultChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *HttpFaultChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}
// GetDuration would return the duration for chaos
func (in *HttpFaultChaos) GetDuration() (int, error) {
	if in.Spec.Duration == nil {
		return 0, nil
	}
	duration, err := strconv.Atoi(*in.Spec.Duration)
	if err != nil {
		return 0, err
	}
	return duration, nil
}

func (in *HttpFaultChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindHttpFaultChaos,
		StartTime: in.CreationTimestamp.Time,
		Action:    string(in.Spec.Action),
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

// HttpFaultChaosList contains a list of HttpFaultChaos
type HttpFaultChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HttpFaultChaos `json:"items"`
}

// ListChaos returns a list of io chaos
func (in *HttpFaultChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

func init() {
	SchemeBuilder.Register(&HttpFaultChaos{}, &HttpFaultChaosList{})
}
