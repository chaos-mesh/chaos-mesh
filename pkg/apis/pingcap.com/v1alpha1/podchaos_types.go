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

// PodChaosAction represents the chaos action about pods.
type PodChaosAction string

const (
	// PodKillAction represents the chaos action of killing pods.
	PodKillAction PodChaosAction = "pod-kill"
	// PodFailureAction represents the chaos action of injecting errors to pods.
	// This action will cause the pod to not be created for a while.
	PodFailureAction PodChaosAction = "pod-failure"
)

// PodChaosMode represents the mode to run pod chaos action.
type PodChaosMode string

const (
	// OnePodMode represents that the system will do the chaos action on one pod selected randomly.
	OnePodMode PodChaosMode = "one"
	// AllPodMode represents that the system will do the chaos action on all pods
	// regardless of status (not ready or not running pods includes).
	// Use this label carefully.
	AllPodMode PodChaosMode = "all"
	// FixedPodMode represents that the system will do the chaos action on a specific number of running pods.
	FixedPodMode PodChaosMode = "fixed"
	// FixedPercentPodMode to specify a fixed % that can be inject chaos action.
	FixedPercentPodMode PodChaosMode = "fixed-percent"
	// RandomMaxPercentPodMode to specify a maximum % that can be inject chaos action.
	RandomMaxPercentPodMode PodChaosMode = "random-max-percent"
)

const (
	DefaultStatusRetentionTime = 1 * time.Hour
	DefaultCleanStatusInterval = 5 * time.Minute
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodChaos is the control script`s spec.
type PodChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec PodChaosSpec `json:"spec"`

	// Most recently observed status of the chaos experiment about pods
	Status PodChaosStatus `json:"status"`
}

// PodChaosSpec defines the attributes that a user creates on a chaos experiment about pods.
type PodChaosSpec struct {
	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Scheduler defines some schedule rules to
	// control the running time of the chaos experiment about pods.
	Scheduler SchedulerSpec `json:"scheduler"`

	// Action defines the specific pod chaos action.
	// Supported action: pod-kill / pod-failure
	// Default action: pod-kill
	Action PodChaosAction `json:"action"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	Mode PodChaosMode `json:"mode"`

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
	Duration string `json:"duration"`

	// The duration in seconds before the object should be deleted. Value must be non-negative integer.
	// The value zero indicates delete immediately.
	// +optional
	GracePeriodSeconds int64 `json:"gracePeriodSeconds"`

	// StatusRetentionTime defines the retention time of experiment status.
	// A statusRetentionTime string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	StatusRetentionTime string `json:"statusSavedTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodChaosList is PodChaos list.
type PodChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PodChaos `json:"items"`
}

// PodChaosStatus represents the current status of the chaos experiment about pods.
type PodChaosStatus struct {
	// Phase is the chaos status.
	Phase  ChaosPhase `json:"phase"`
	Reason string     `json:"reason,omitempty"`

	// Experiments records the experiment records of the most recent period of time.
	// Keeping one hour of status record by default.
	Experiments []PodChaosExperimentStatus `json:"experiments"`
}

func (ps *PodChaosStatus) SetExperimentRecord(record PodChaosExperimentStatus) {
	for index, exp := range ps.Experiments {
		if exp.Time == record.Time {
			ps.Experiments[index] = record
		}
	}

	ps.Experiments = append(ps.Experiments, record)
}

// CleanExpiredStatusRecords cleans the expired status records.
func (ps *PodChaosStatus) CleanExpiredStatusRecords(retentionTime time.Duration) {
	var experiments []PodChaosExperimentStatus

	nowTime := time.Now()
	for _, exp := range ps.Experiments {
		if exp.Time.Add(retentionTime).After(nowTime) {
			experiments = append(experiments, exp)
		}
	}

	ps.Experiments = experiments
}

// ChaosPhase is the current status of chaos task.
type ChaosPhase string

const (
	ChaosPhaseNone     ChaosPhase = ""
	ChaosPhaseNormal              = "Normal"
	ChaosPahseAbnormal            = "Abnormal"
)

// ExperimentPhase is the current status of chaos experiment.
type ExperimentPhase string

const (
	ExperimentPhaseNone     ExperimentPhase = ""
	ExperimentPhaseRunning                  = "Running"
	ExperimentPhaseFailed                   = "Failed"
	ExperimentPhaseFinished                 = "Finished"
)

// PodChaosExperimentStatus represents information about the status of the podchaos experiment.
type PodChaosExperimentStatus struct {
	Phase  ExperimentPhase `json:"phase"`
	Reason string          `json:"reason"`
	Time   metav1.Time     `json:"time"`
	Pods   []PodStatus     `json:"podChaos"`
}

func (pe *PodChaosExperimentStatus) SetPods(pod PodStatus) {
	for index, p := range pe.Pods {
		if p.Namespace == pod.Namespace && p.Name == pod.Namespace {
			pe.Pods[index] = pod
		}
	}

	pe.Pods = append(pe.Pods, pod)
}

// PodStatus represents information about the status of a pod in chaos experiment.
type PodStatus struct {
	Namespace string      `json:"namespace"`
	Name      string      `json:"name"`
	Action    string      `json:"action"`
	HostIP    string      `json:"hostIP"`
	PodIP     string      `json:"podIP"`
	StartTime metav1.Time `json:"startTime"`
	EndTime   metav1.Time `json:"endTime"`

	// A brief CamelCase message indicating details about the chaos action.
	// e.g. "delete this pod" or "pause this pod duration 5m"
	// +optional
	Message string `json:"message"`
}
