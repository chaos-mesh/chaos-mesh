// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v1alpha1

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	// PauseAnnotationKey defines the annotation used to pause a chaos
	PauseAnnotationKey = "experiment.chaos-mesh.org/pause"
	LabelManagedBy     = "managed-by"
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
	Reason string `json:"reason,omitempty"`
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
	// InjectedCount is a counter to record the sum of successful injections
	InjectedCount int `json:"injectedCount"`
	// RecoveredCount is a counter to record the sum of successful recoveries
	RecoveredCount int `json:"recoveredCount"`
	// Events are the essential details about the injections and recoveries
	Events []RecordEvent `json:"events,omitempty"`
}

type Phase string

const (
	// NotInjected means the target is not injected yet. The controller could call "Inject" on the target
	NotInjected Phase = "Not Injected"
	// Injected means the target is injected. It's safe to recover it.
	Injected Phase = "Injected"
)

type RecordEvent struct {
	// Type means the stage of this event
	Type RecordEventType `json:"type"`
	// Operation represents the operation we are doing, when we crate this event
	Operation RecordEventOperation `json:"operation"`
	// Message is the detail message, e.g. the reason why we failed to inject the chaos
	Message string `json:"message,omitempty"`
	// Timestamp is time when we create this event
	Timestamp *metav1.Time `json:"timestamp"`
}

type RecordEventType string

const (
	// TypeSucceeded means the stage of this event is successful
	TypeSucceeded RecordEventType = "Succeeded"
	// TypeFailed means the stage of this event is failed
	TypeFailed RecordEventType = "Failed"
)

type RecordEventOperation string

const (
	// Apply means this event is recorded, when we inject the chaos
	// typically, when we call impl.Apply()
	Apply RecordEventOperation = "Apply"
	// Recover means this event is recorded, when we recover the chaos
	// typically, when we call impl.Recover()
	Recover RecordEventOperation = "Recover"
)

// NewRecordEvent is a constructor of RecordEvent in status
func NewRecordEvent(eventType RecordEventType, eventStage RecordEventOperation,
	msg string, time metav1.Time) *RecordEvent {
	return &RecordEvent{
		Type:      eventType,
		Operation: eventStage,
		Message:   msg,
		Timestamp: &time,
	}
}

var log = ctrl.Log.WithName("api")

// +kubebuilder:object:generate=false

// InnerObject is basic Object for the Reconciler
type InnerObject interface {
	StatefulObject
	IsDeleted() bool
	IsPaused() bool
	DurationExceeded(time.Time) (bool, time.Duration, error)
	IsOneShot() bool
}

// +kubebuilder:object:generate=false

// StatefulObject defines a basic Object that can get the status
type StatefulObject interface {
	GenericChaos
	GetStatus() *ChaosStatus
}

// +kubebuilder:object:generate=false
type InnerObjectWithCustomStatus interface {
	InnerObject

	GetCustomStatus() interface{}
}

// +kubebuilder:object:generate=false
type InnerObjectWithSelector interface {
	InnerObject

	GetSelectorSpecs() map[string]interface{}
}

// +kubebuilder:object:generate=false

// WebhookObject is basic Object which implement `webhook.Validator` and `webhook.Defaulter`
type WebhookObject interface {
	webhook.Validator
	webhook.Defaulter
}

// +kubebuilder:object:generate=false
type RemoteObject interface {
	StatefulObject
	GetRemoteCluster() string
}
