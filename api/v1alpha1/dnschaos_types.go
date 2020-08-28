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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KindDNSChaos is the kind for network chaos
const KindDNSChaos = "DNSChaos"

func init() {
	all.register(KindDNSChaos, &ChaosKind{
		Chaos:     &DNSChaos{},
		ChaosList: &DNSChaosList{},
	})
}

// ChaosAction represents the chaos action about pods.
type DNSChaosAction string

// DNSChaosSpec defines the desired state of DNSChaos
type DNSChaosSpec struct {
	// Action defines the specific network chaos action.
	// Supported action: partition, netem, delay, loss, duplicate, corrupt
	// Default action: delay
	// +kubebuilder:validation:Enum=netem;delay;loss;duplicate;corrupt;partition;bandwidth
	Action DNSChaosAction `json:"action"`

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

	// ExternalTargets represents network targets outside k8s
	// +optional
	ExternalTargets []string `json:"externalTargets,omitempty"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *DNSChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *DNSChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *DNSChaosSpec) GetValue() string {
	return in.Value
}

// DNSChaosStatus defines the observed state of DNSChaos
type DNSChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// DNSChaos is the Schema for the networkchaos API
type DNSChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec DNSChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status DNSChaosStatus `json:"status"`
}

func (in *DNSChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *DNSChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *DNSChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetDuration would return the duration for chaos
func (in *DNSChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *DNSChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *DNSChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *DNSChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *DNSChaos) SetNextRecover(t time.Time) {
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
func (in *DNSChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos returns a chaos instance
func (in *DNSChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindDNSChaos,
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

// DNSChaosList contains a list of DNSChaos
type DNSChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSChaos `json:"items"`
}

// ListChaos returns a list of network chaos
func (in *DNSChaosList) ListChaos() []*ChaosInstance {
	if len(in.Items) == 0 {
		return nil
	}
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

func init() {
	SchemeBuilder.Register(&DNSChaos{}, &DNSChaosList{})
}
