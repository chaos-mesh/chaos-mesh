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
	"errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KindHTTPChaos is the kind for http chaos
const KindHTTPChaos = "HTTPChaos"

func init() {
	all.register(KindHTTPChaos, &ChaosKind{
		Chaos:     &HTTPChaos{},
		ChaosList: &HTTPChaosList{},
	})
}

// HTTPChaosAction represents the chaos action about HTTP.
type HTTPChaosAction string

const (
	HTTPDelayAction HTTPChaosAction = "delay"
	HTTPAbortAction                 = "abort"
	HTTPMixedAction                 = "mixed"
)

type Matcher struct {
	Name           string  `json:"name"`
	ExactMatch     *string `json:"exact_match,omitempty"`
	RegexMatch     *string `json:"regex_match,omitempty"`
	SafeRegexMatch *string `json:"safe_regex_match,omitempty"`
	RangeMatch     *string `json:"range_match,omitempty"`
	PresentMatch   *string `json:"present_match,omitempty"`
	PrefixMatch    *string `json:"prefix_match,omitempty"`
	SuffixMatch    *string `json:"suffix_match,omitempty"`
	InvertMatch    *string `json:"invert_match,omitempty"`
}

type HTTPChaosSpec struct {
	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Scheduler defines some schedule rules to
	// control the running time of the chaos experiment about pods.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Action defines the specific pod chaos action.
	// Supported action: delay | abort | mixed
	// Default action: delay
	// +kubebuilder:validation:Enum=delay;abort;mixed
	Action HTTPChaosAction `json:"action"`

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

	// Duration represents the duration of the chaos action.
	// It is required when the action is `PodFailureAction`.
	// A duration string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Percent defines the percentage of injection errors and provides a number from 0-100.
	// default: 100.
	// +optional
	Percent string `json:"percent,omitempty"`

	// Specifies how the header match will be performed to route the request.
	Headers []Matcher `json:"headers,omitempty"`
}

func (in *HTTPChaosSpec) GetHeaders() []Matcher {
	return in.Headers
}

func (in *HTTPChaosSpec) GetMode() PodMode {
	return in.Mode
}

func (in *HTTPChaosSpec) GetValue() string {
	return in.Value
}

func (in *HTTPChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

type HTTPChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// HTTPChaos is the Schema for the HTTPchaos API
type HTTPChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HTTPChaosSpec   `json:"spec,omitempty"`
	Status HTTPChaosStatus `json:"status,omitempty"`
}

func (in *HTTPChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *HTTPChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *HTTPChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// IsRenewed returns whether this resource selected source/target item has been changed
func (in *HTTPChaos) IsRenewed() bool {
	// check chaos status experiment select items with its staging values.
	// if there is a differ, return true.
	experiment := in.Status.Experiment
	if !IsSameTwoPodStatuses(experiment.SourcePodRecords, experiment.StagingSourcePodRecords) ||
		!IsSameTwoPodStatuses(experiment.TargetPodRecords, experiment.StagingTargetPodRecords) ||
		!IsSameTwoExternalCIDRs(experiment.ExternalCIDRs, experiment.StagingExternalCIDRs) {
		return true
	}
	return false
}

// PromoteSelectItems promotes the staging select items to production
func (in *HTTPChaos) PromoteSelectItems() error {
	return errors.New("not implemented yet")
}

// GetDuration would return the duration for chaos
func (in *HTTPChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *HTTPChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *HTTPChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *HTTPChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *HTTPChaos) SetNextRecover(t time.Time) {
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
func (in *HTTPChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

func (in *HTTPChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindHTTPChaos,
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

// HTTPChaosList contains a list of HTTPChaos
type HTTPChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPChaos `json:"items"`
}

// ListChaos returns a list of io chaos
func (in *HTTPChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

func init() {
	SchemeBuilder.Register(&HTTPChaos{}, &HTTPChaosList{})
}
