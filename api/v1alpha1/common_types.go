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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// PauseAnnotationKey defines the annotation used to pause a chaos
	PauseAnnotationKey = "experiment.chaos-mesh.org/pause"
)

// LabelSelectorRequirements is list of LabelSelectorRequirement
type LabelSelectorRequirements []metav1.LabelSelectorRequirement

// SelectorSpec defines the some selectors to select objects.
// If the all selectors are empty, all objects will be used in chaos experiment.
type SelectorSpec struct {
	// Namespaces is a set of namespace to which objects belong.
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`

	// Nodes is a set of node name and objects must belong to these nodes.
	// +optional
	Nodes []string `json:"nodes,omitempty"`

	// Pods is a map of string keys and a set values that used to select pods.
	// The key defines the namespace which pods belong,
	// and the each values is a set of pod names.
	// +optional
	Pods map[string][]string `json:"pods,omitempty"`

	// Map of string keys and values that can be used to select nodes.
	// Selector which must match a node's labels,
	// and objects must belong to these selected nodes.
	// +optional
	NodeSelectors map[string]string `json:"nodeSelectors,omitempty"`

	// Map of string keys and values that can be used to select objects.
	// A selector based on fields.
	// +optional
	FieldSelectors map[string]string `json:"fieldSelectors,omitempty"`

	// Map of string keys and values that can be used to select objects.
	// A selector based on labels.
	// +optional
	LabelSelectors map[string]string `json:"labelSelectors,omitempty"`

	// a slice of label selector expressions that can be used to select objects.
	// A list of selectors based on set-based label expressions.
	// +optional
	ExpressionSelectors LabelSelectorRequirements `json:"expressionSelectors,omitempty"`

	// Map of string keys and values that can be used to select objects.
	// A selector based on annotations.
	// +optional
	AnnotationSelectors map[string]string `json:"annotationSelectors,omitempty"`

	// PodPhaseSelectors is a set of condition of a pod at the current time.
	// supported value: Pending / Running / Succeeded / Failed / Unknown
	// +optional
	PodPhaseSelectors []string `json:"podPhaseSelectors,omitempty"`
}

// SchedulerSpec defines information about schedule of the chaos experiment.
type SchedulerSpec struct {
	// Cron defines a cron job rule.
	//
	// Some rule examples:
	// "0 30 * * * *" means to "Every hour on the half hour"
	// "@hourly"      means to "Every hour"
	// "@every 1h30m" means to "Every hour thirty"
	//
	// More rule info: https://godoc.org/github.com/robfig/cron
	Cron string `json:"cron"`
}

// PodMode represents the mode to run pod chaos action.
type PodMode string

const (
	// OnePodMode represents that the system will do the chaos action on one pod selected randomly.
	OnePodMode PodMode = "one"
	// AllPodMode represents that the system will do the chaos action on all pods
	// regardless of status (not ready or not running pods includes).
	// Use this label carefully.
	AllPodMode PodMode = "all"
	// FixedPodMode represents that the system will do the chaos action on a specific number of running pods.
	FixedPodMode PodMode = "fixed"
	// FixedPercentPodMode to specify a fixed % that can be inject chaos action.
	FixedPercentPodMode PodMode = "fixed-percent"
	// RandomMaxPercentPodMode to specify a maximum % that can be inject chaos action.
	RandomMaxPercentPodMode PodMode = "random-max-percent"
)

// ChaosPhase is the current status of chaos task.
type ChaosPhase string

const (
	ChaosPhaseNone     ChaosPhase = ""
	ChaosPhaseNormal   ChaosPhase = "Normal"
	ChaosPhaseAbnormal ChaosPhase = "Abnormal"
)

type ChaosStatus struct {
	FailedMessage string `json:"failedMessage,omitempty"`

	Scheduler ScheduleStatus `json:"scheduler,omitempty"`

	// Experiment records the last experiment state.
	Experiment ExperimentStatus `json:"experiment"`
}

func (in *ChaosStatus) GetNextStart() time.Time {
	if in.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Scheduler.NextStart.Time
}

func (in *ChaosStatus) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Scheduler.NextStart = nil
		return
	}

	if in.Scheduler.NextStart == nil {
		in.Scheduler.NextStart = &metav1.Time{}
	}
	in.Scheduler.NextStart.Time = t
}

func (in *ChaosStatus) GetNextRecover() time.Time {
	if in.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Scheduler.NextRecover.Time
}

func (in *ChaosStatus) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Scheduler.NextRecover = nil
		return
	}

	if in.Scheduler.NextRecover == nil {
		in.Scheduler.NextRecover = &metav1.Time{}
	}
	in.Scheduler.NextRecover.Time = t
}

// ScheduleStatus is the current status of chaos scheduler.
type ScheduleStatus struct {
	// Next time when this action will be applied again
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
}

// ExperimentPhase is the current status of chaos experiment.
type ExperimentPhase string

const (
	ExperimentPhaseUninitialized ExperimentPhase = ""
	ExperimentPhaseRunning       ExperimentPhase = "Running"
	ExperimentPhaseWaiting       ExperimentPhase = "Waiting"
	ExperimentPhasePaused        ExperimentPhase = "Paused"
	ExperimentPhaseFailed        ExperimentPhase = "Failed"
	ExperimentPhaseFinished      ExperimentPhase = "Finished"
)

type ExperimentStatus struct {
	// +optional
	Phase ExperimentPhase `json:"phase,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	EndTime *metav1.Time `json:"endTime,omitempty"`
	// +optional
	Duration string `json:"duration,omitempty"`
	// +optional
	PodRecords []PodStatus `json:"podRecords,omitempty"`
}

var log = ctrl.Log.WithName("validate-webhook")

// +kubebuilder:object:generate=false

// InnerSchedulerObject is the Object for the twophase reconcile
type InnerSchedulerObject interface {
	InnerObject
	GetDuration() (*time.Duration, error)

	GetNextStart() time.Time
	SetNextStart(time.Time)

	GetNextRecover() time.Time
	SetNextRecover(time.Time)

	GetScheduler() *SchedulerSpec
}

// +kubebuilder:object:generate=false

// InnerObject is basic Object for the Reconciler
type InnerObject interface {
	IsDeleted() bool
	IsPaused() bool
	GetChaos() *ChaosInstance
	StatefulObject
}

// +kubebuilder:object:generate=false

// StatefulObject defines a basic Object that can get the status
type StatefulObject interface {
	runtime.Object
	GetStatus() *ChaosStatus
}

// +kubebuilder:object:generate=false

// ChaosInstance defines some common attribute for a chaos
type ChaosInstance struct {
	Name      string
	Namespace string
	Kind      string
	StartTime time.Time
	EndTime   time.Time
	Action    string
	Duration  string
	Status    string
	UID       string
}

// +kubebuilder:object:generate=false

// ChaosList defines a common interface for chaos lists
type ChaosList interface {
	runtime.Object
	ListChaos() []*ChaosInstance
}
