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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// PauseAnnotationKey defines the annotation used to pause a chaos
	PauseAnnotationKey = "experiment.chaos-mesh.org/pause"
)

type ChaosStatus struct {
	// Conditions represents the current global condition of the chaos
	// +optional
	Conditions []ChaosCondition `json:"conditions,omitempty"`

	// Experiment records the last experiment state.
	Experiment ExperimentStatus `json:"experiment"`
}

type ChaosConditionType string

const (
	ConditionSelected     ChaosConditionType = "Selected"
	ConditionAllInjected  ChaosConditionType = "AllInjected"
	ConditionAllRecovered ChaosConditionType = "AllRecovered"
	ConditionPaused       ChaosConditionType = "Paused"
)

type ChaosCondition struct {
	Type   ChaosConditionType     `json:"type"`
	Status corev1.ConditionStatus `json:"status"`
	// +optional
	Reason string `json:"reason"`
}

type DesiredPhase string

const (
	// The target of `RunningPhase` is to make all selected targets (container or pod) into "Injected" phase
	RunningPhase DesiredPhase = "Run"
	// The target of `StoppedPhase` is to make all selected targets (container or pod) into "NotInjected" phase
	StoppedPhase DesiredPhase = "Stop"
)

type ExperimentStatus struct {
	// +kubebuilder:validation:Enum=Run;Stop
	DesiredPhase `json:"desiredPhase,omitempty"`
	// +optional
	// Records are used to track the running status
	Records []*Record `json:"containerRecords,omitempty"`
}

type Record struct {
	Id          string `json:"id"`
	SelectorKey string `json:"selectorKey"`
	Phase       Phase  `json:"phase"`
}

type Phase string

const (
	// NotInjected means the target is not injected yet. The controller could call "Inject" on the target
	NotInjected Phase = "Not Injected"
	// Injected means the target is injected. It's safe to recover it.
	Injected Phase = "Injected"
)

var log = ctrl.Log.WithName("api")

// +kubebuilder:object:generate=false

// InnerObject is basic Object for the Reconciler
type InnerObject interface {
	IsDeleted() bool
	IsPaused() bool
	GetChaos() *ChaosInstance
	DurationExceeded(time.Time) (bool, time.Duration, error)
	IsOneShot() bool
	StatefulObject
}

// +kubebuilder:object:generate=false

// StatefulObject defines a basic Object that can get the status
type StatefulObject interface {
	runtime.Object
	GetStatus() *ChaosStatus
	GetObjectMeta() *metav1.ObjectMeta
}

// +kubebuilder:object:generate=false
// MetaObject defines a very basic Object that can get meta
type MetaObject interface {
	runtime.Object
	GetObjectMeta() *metav1.ObjectMeta
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
	Status    ChaosStatus
	UID       string
}

// +kubebuilder:object:generate=false

// ChaosList defines a common interface for chaos lists
type ChaosList interface {
	runtime.Object
	ListChaos() []*ChaosInstance
}
