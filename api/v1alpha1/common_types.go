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
	"math"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ProtectorSpec protects the target cluster from being corrupted by the chaos.
// For now, some chaos could not be recovered such as pod-kill. If we continue
// to inject it into our application cluster, it will finally pull down our
// entire application cluster. That may be not our intention. We should protect
// our application cluster from that disaster via a protector to make the
// application cluster always in available.
type ProtectorSpec struct {

	// MaxInjectPercent specifies the ratio of maximum pods involved in the
	// chaos to be injected later.  A value of less than or equal to 0, or
	// greater than 100, we see it as illegal. It takes no effects for the
	// protection.
	MaxInjectPercent int `json:"maxInjectPercent,omitempty"`
}

// GetMaxInjectablePods calculates the maximum number pods to inject the chaos.
func (in *ProtectorSpec) GetMaxInjectablePods(total int, failed int) int {
	if in.MaxInjectPercent <= 0 || in.MaxInjectPercent > 100 {
		in.MaxInjectPercent = 100
	}
	return int(math.Floor(float64(total*in.MaxInjectPercent/100.0 - failed)))
}

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

	// Map of string keys and values that can be used to select objects.
	// A selector based on annotations.
	// +optional
	AnnotationSelectors map[string]string `json:"annotationSelectors,omitempty"`

	// PodPhaseSelectors is a set of condition of a pod at the current time.
	// supported value: Pending / Running / Succeeded / Failed / Unknown
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
	// Phase is the chaos status.
	Phase  ChaosPhase `json:"phase"`
	Reason string     `json:"reason,omitempty"`

	// Experiment records the last experiment state.
	Experiment ExperimentStatus `json:"experiment"`

	MaxInjectablePods int `json:"maxInjectNumPods,omitempty"`
}

// ExperimentPhase is the current status of chaos experiment.
type ExperimentPhase string

const (
	ExperimentPhaseRunning  ExperimentPhase = "Running"
	ExperimentPhaseFailed   ExperimentPhase = "Failed"
	ExperimentPhaseFinished ExperimentPhase = "Finished"
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
	Pods []PodStatus `json:"podChaos,omitempty"`
}

const (
	invalidConfigurationMsg = "invalid configuration"
)

var log = ctrl.Log.WithName("validate-webhook")
