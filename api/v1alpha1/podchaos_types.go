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

// +kubebuilder:object:root=true

// PodChaos is the control script`s spec.
type PodChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec PodChaosSpec `json:"spec"`

	// +optional
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
	Duration string `json:"duration"`

	// The duration in seconds before the object should be deleted. Value must be non-negative integer.
	// The value zero indicates delete immediately.
	// +optional
	GracePeriodSeconds int64 `json:"gracePeriodSeconds"`

	// Next time when this action will be applied again
	// +nullable
	NextStart metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +nullable
	NextRecover metav1.Time `json:"nextRecover,omitempty"`
}

// +kubebuilder:object:root=true

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

	// Experiment records the last experiment state.
	Experiment PodChaosExperimentStatus `json:"experiment"`
}

// ChaosPhase is the current status of chaos task.
type ChaosPhase string

const (
	ChaosPhaseNone     ChaosPhase = ""
	ChaosPhaseNormal              = "Normal"
	ChaosPhaseAbnormal            = "Abnormal"
)

// ExperimentPhase is the current status of chaos experiment.
type ExperimentPhase string

const (
	ExperimentPhaseRunning  ExperimentPhase = "Running"
	ExperimentPhaseFailed                   = "Failed"
	ExperimentPhaseFinished                 = "Finished"
)

// PodChaosExperimentStatus represents information about the status of the apis experiment.
type PodChaosExperimentStatus struct {
	// +optional
	Phase ExperimentPhase `json:"phase,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +nullable
	StartTime metav1.Time `json:"startTime,omitempty"`
	// +nullable
	EndTime metav1.Time `json:"endTime,omitempty"`
	// +optional
	Pods []PodStatus `json:"podChaos,omitempty"`
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
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Action    string `json:"action"`
	HostIP    string `json:"hostIP"`
	PodIP     string `json:"podIP"`

	// A brief CamelCase message indicating details about the chaos action.
	// e.g. "delete this pod" or "pause this pod duration 5m"
	// +optional
	Message string `json:"message"`
}
