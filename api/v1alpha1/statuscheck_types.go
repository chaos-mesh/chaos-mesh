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
	// StatusCheckSynchronous means the status check will exit
	// immediately after success or failure.
	StatusCheckSynchronous StatusCheckMode = "Synchronous"
	// StatusCheckContinuous means the status check will continue to
	// execute until the duration is exceeded or the status check fails.
	StatusCheckContinuous StatusCheckMode = "Continuous"
)

type StatusCheckType string

const (
	TypeHTTP StatusCheckType = "HTTP"
)

type StatusCheckSpec struct {
	// Mode defines the execution mode of the status check.
	// Support type: Synchronous / Continuous
	// +optional
	// +kubebuilder:validation:Enum=Synchronous;Continuous
	Mode StatusCheckMode `json:"mode,omitempty"`

	// Type defines the specific status check type.
	// Support type: HTTP
	// +kubebuilder:default=HTTP
	// +kubebuilder:validation:Enum=HTTP
	Type StatusCheckType `json:"type"`

	// Duration defines the duration of the status check.
	// A duration string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	// TimeoutSeconds defines the number of seconds after which the status check times out.
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`

	// PeriodSeconds defines how often (in seconds) to perform the status check.
	// +optional
	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=1
	PeriodSeconds int `json:"periodSeconds,omitempty"`

	// FailureThreshold defines when a status check fails,
	// it will try FailureThreshold times before giving up.
	// +optional
	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=1
	FailureThreshold int `json:"failureThreshold,omitempty"`

	// RecordsHistoryLimit defines the number of record to retain.
	// +optional
	// +kubebuilder:default=100
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000
	RecordsHistoryLimit int `json:"recordsHistoryLimit,omitempty"`

	// +optional
	*EmbedStatusCheck `json:",inline,omitempty"`
}

type StatusCheckStatus struct {
	// StartTime represents time when the status check started to execute.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime represents time when the status check was completed.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Count represents the total number of the status check executed.
	// +optional
	Count int64 `json:"count,omitempty"`

	// Conditions represents the latest available observations of a StatusCheck's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []StatusCheckCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Records contains the history of the execution of StatusCheck.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Records []StatusCheckRecord `json:"records,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
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
	// StatusCheckConditionComplete means the status check has completed its execution.
	StatusCheckConditionComplete StatusCheckConditionType = "Complete"
	// StatusCheckConditionFailed means the status check has failed its execution.
	StatusCheckConditionFailed StatusCheckConditionType = "Failed"
)

type StatusCheckReason string

const (
	StatusCheckSuccess StatusCheckReason = "StatusCheckSuccess"
	// TODO add more reason when implementing StatusCheck
)

type StatusCheckCondition struct {
	Type               StatusCheckConditionType `json:"type"`
	Status             corev1.ConditionStatus   `json:"status"`
	Reason             StatusCheckReason        `json:"reason"`
	LastProbeTime      *metav1.Time             `json:"lastProbeTime"`
	LastTransitionTime *metav1.Time             `json:"lastTransitionTime"`
}

type EmbedStatusCheck struct {
	// +optional
	HTTPStatusCheck *HTTPStatusCheck `json:"http,omitempty"`
}

type HTTPHeaderPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HTTPCriteria struct {
	// StatusCode defines the expected http status code for the request.
	// A statusCode string could be a single code (e.g. 200),
	// or a range (e.g. 200-400).
	StatusCode string `json:"statusCode,omitempty" webhook:"StatusCode"`
	// TODO: support response body
}

type HTTPRequestMethod string

const (
	MethodGet  = "GET"
	MethodPost = "POST"
)

type HTTPStatusCheck struct {
	RequestUrl string `json:"url"`

	// +optional
	// +kubebuilder:validation:Enum=GET;POST
	// +kubebuilder:default=GET
	RequestMethod HTTPRequestMethod `json:"method,omitempty"`
	// +optional
	RequestHeaders []HTTPHeaderPair `json:"headers,omitempty"`
	// +optional
	RequestBody string `json:"body,omitempty"`
	// Criteria defines how to determine the result of the status check.
	Criteria HTTPCriteria `json:"criteria,omitempty"`
}

// +kubebuilder:object:root=true

// StatusCheckList contains a list of StatusCheck
type StatusCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StatusCheck `json:"items"`
}
