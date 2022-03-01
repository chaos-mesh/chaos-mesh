// Copyright Chaos Mesh Authors.
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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +chaos-mesh:base
type StatusCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a status check
	Spec StatusCheckSpec `json:"spec"`

	// +optional
	// Most recently observed status of status check
	Status StatusCheckStatus `json:"status"`
}

type StatusCheckMode string

const (
	StatusCheckSynchronous StatusCheckMode = "Synchronous"
	StatusCheckContinuous  StatusCheckMode = "Continuous"
)

type StatusCheckType string

const (
	TypeHttp StatusCheckType = "HTTP"
)

type StatusCheckSpec struct {
	// +kubebuilder:validation:Enum=Synchronous;Continuous
	Mode StatusCheckMode `json:"mode"`

	// +kubebuilder:default=HTTP
	// +kubebuilder:validation:Enum=HTTP
	Type StatusCheckType `json:"type"`

	// +optional
	Deadline *string `json:"deadline"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	TimeoutSeconds int `json:"timeoutSeconds"`

	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=1
	PeriodSeconds int `json:"periodSeconds"`

	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=1
	FailureThreshold int `json:"failureThreshold"`

	// +kubebuilder:default=100
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000
	HistoryLimit int `json:"historyLimit"`

	// +optional
	// +kubebuilder:default=true
	AbortIfFailed bool `json:"abortIfFailed"`
}

type StatusCheckStatus struct {
	// Conditions represents the latest available observations of a StatusCheck's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []StatusCheckCondition `json:"conditions" patchStrategy:"merge" patchMergeKey:"type"`

	// Records represents the latest available observations of a StatusCheck's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Records []StatusCheckRecord `json:"records" patchStrategy:"merge" patchMergeKey:"type"`
}

type StatusCheckOutcome string

const (
	StatusCheckOutcomeSuccess StatusCheckOutcome = "Success"
	StatusCheckOutcomeFailure StatusCheckOutcome = "Failure"
)

type StatusCheckRecord struct {
	ProbeTime *metav1.Time       `json:"probeTime"`
	Outcome   StatusCheckOutcome `json:"outcome"`
}

type StatusCheckConditionType string

const (
	StatusCheckConditionAccomplished   StatusCheckConditionType = "Accomplished"
	StatusCheckConditionDeadlineExceed StatusCheckConditionType = "DeadlineExceed"
	StatusCheckConditionProbeSucceed   StatusCheckConditionType = "ProbeSucceed"
	StatusCheckConditionAborted        StatusCheckConditionType = "Aborted"
)

type StatusCheckCondition struct {
	Type   StatusCheckConditionType `json:"type"`
	Status corev1.ConditionStatus   `json:"status"`
	// +optional
	Reason string `json:"reason"`
}

// Reasons
const (
	StatusCheckSuccess string = "StatusCheckSuccess"
)

type HTTPHeaderPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HTTPCriteria struct {
	ResponseCode string `json:"responseCode"`
	// TODO: support response body
}

type HTTPStatusCheck struct {
	RequestUrl     string           `json:"requestUrl"`
	RequestMethod  string           `json:"requestMethod"`
	RequestHeaders []HTTPHeaderPair `json:"requestHeaders"`
	Criteria       HTTPCriteria     `json:"criteria"`
}

// +kubebuilder:object:root=true

// StatusCheckList contains a list of StatusCheck
type StatusCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StatusCheck `json:"items"`
}