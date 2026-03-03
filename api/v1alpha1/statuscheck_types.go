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
	"net/http"
	"time"

	"github.com/pkg/errors"
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
	Status StatusCheckStatus `json:"status,omitempty"`
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

	// Duration defines the duration of the whole status check if the
	// number of failed execution does not exceed the failure threshold.
	// Duration is available to both `Synchronous` and `Continuous` mode.
	// A duration string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix,
	// such as "300ms", "-1.5h" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	// TimeoutSeconds defines the number of seconds after which
	// an execution of status check times out.
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`

	// IntervalSeconds defines how often (in seconds) to perform
	// an execution of status check.
	// +optional
	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=1
	IntervalSeconds int `json:"intervalSeconds,omitempty"`

	// FailureThreshold defines the minimum consecutive failure
	// for the status check to be considered failed.
	// +optional
	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=1
	FailureThreshold int `json:"failureThreshold,omitempty"`

	// SuccessThreshold defines the minimum consecutive successes
	// for the status check to be considered successful.
	// SuccessThreshold only works for `Synchronous` mode.
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	SuccessThreshold int `json:"successThreshold,omitempty"`

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
	Records []StatusCheckRecord `json:"records,omitempty"`
}

type StatusCheckOutcome string

const (
	StatusCheckOutcomeSuccess StatusCheckOutcome = "Success"
	StatusCheckOutcomeFailure StatusCheckOutcome = "Failure"
)

type StatusCheckRecord struct {
	StartTime *metav1.Time       `json:"startTime"`
	Outcome   StatusCheckOutcome `json:"outcome"`
}

type StatusCheckConditionType string

const (
	// StatusCheckConditionCompleted means the status check is completed.
	// It will be `True`, in the following scenarios:
	// 1. the duration is exceeded
	// 2. the failure threshold is exceeded
	// 3. the success threshold is exceeded (only the `Synchronous` mode)
	StatusCheckConditionCompleted StatusCheckConditionType = "Completed"
	// StatusCheckConditionDurationExceed means the duration is exceeded.
	StatusCheckConditionDurationExceed StatusCheckConditionType = "DurationExceed"
	// StatusCheckConditionFailureThresholdExceed means the failure threshold is exceeded.
	StatusCheckConditionFailureThresholdExceed StatusCheckConditionType = "FailureThresholdExceed"
	// StatusCheckConditionSuccessThresholdExceed means the success threshold is exceeded.
	StatusCheckConditionSuccessThresholdExceed StatusCheckConditionType = "SuccessThresholdExceed"
)

type StatusCheckReason string

const (
	StatusCheckDurationExceed         StatusCheckReason = "StatusCheckDurationExceed"
	StatusCheckFailureThresholdExceed StatusCheckReason = "StatusCheckFailureThresholdExceed"
	StatusCheckSuccessThresholdExceed StatusCheckReason = "StatusCheckSuccessThresholdExceed"
	StatusCheckExecutionFailed        StatusCheckReason = "StatusCheckExecutionFailed"
	StatusCheckExecutionSucceed       StatusCheckReason = "StatusCheckExecutionSucceed"
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

type HTTPCriteria struct {
	// StatusCode defines the expected http status code for the request.
	// A statusCode string could be a single code (e.g. 200), or
	// an inclusive range (e.g. 200-400, both `200` and `400` are included).
	StatusCode string `json:"statusCode" webhook:"StatusCode"`
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
	RequestHeaders http.Header `json:"headers,omitempty"`
	// +optional
	RequestBody string `json:"body,omitempty"`
	// Criteria defines how to determine the result of the status check.
	Criteria HTTPCriteria `json:"criteria"`
}

// StatusCheckList contains a list of StatusCheck
// +kubebuilder:object:root=true
type StatusCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StatusCheck `json:"items"`
}

func (in *StatusCheckSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, errors.Wrapf(err, "parse duration %s", *in.Duration)
	}
	return &duration, nil
}

func (in *StatusCheck) DurationExceed(now time.Time) (bool, time.Duration, error) {
	if in.Status.StartTime == nil {
		return false, 0, nil
	}
	duration, err := in.Spec.GetDuration()
	if err != nil {
		return false, 0, errors.Wrap(err, "get duration")
	}

	if duration != nil {
		stopTime := in.Status.StartTime.Add(*duration)
		if stopTime.Before(now) {
			return true, 0, nil
		}

		return false, stopTime.Sub(now), nil
	}

	return false, 0, nil
}

// IsCompleted checks if the status check is completed, according to the StatusCheckConditionCompleted condition.
func (in *StatusCheck) IsCompleted() bool {
	for _, condition := range in.Status.Conditions {
		if condition.Type == StatusCheckConditionCompleted &&
			condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
