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

package recorder

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type StatusCheckCompleted struct {
	Msg v1alpha1.StatusCheckReason
}

func (it StatusCheckCompleted) Type() string {
	return corev1.EventTypeNormal
}

func (it StatusCheckCompleted) Reason() string {
	return v1alpha1.StatusCheckCompleted
}

func (it StatusCheckCompleted) Message() string {
	return fmt.Sprintf("status check completed: %s", string(it.Msg))
}

type StatusCheckExecutionFailed struct {
	ExecutorType string
	Msg          string
}

func (it StatusCheckExecutionFailed) Type() string {
	return corev1.EventTypeWarning
}

func (it StatusCheckExecutionFailed) Reason() string {
	return string(v1alpha1.StatusCheckExecutionFailed)
}

func (it StatusCheckExecutionFailed) Message() string {
	return fmt.Sprintf("%s execution of status check failed: %s", it.ExecutorType, it.Msg)
}

type StatusCheckExecutionSucceed struct {
	ExecutorType string
}

func (it StatusCheckExecutionSucceed) Type() string {
	return corev1.EventTypeNormal
}

func (it StatusCheckExecutionSucceed) Reason() string {
	return string(v1alpha1.StatusCheckExecutionSucceed)
}

func (it StatusCheckExecutionSucceed) Message() string {
	return fmt.Sprintf("%s execution of status check succeed", it.ExecutorType)
}

type StatusCheckDurationExceed struct {
}

func (it StatusCheckDurationExceed) Type() string {
	return corev1.EventTypeWarning
}

func (it StatusCheckDurationExceed) Reason() string {
	return string(v1alpha1.StatusCheckDurationExceed)
}

func (it StatusCheckDurationExceed) Message() string {
	return "duration exceed"
}

type StatusCheckFailureThresholdExceed struct {
}

func (it StatusCheckFailureThresholdExceed) Type() string {
	return corev1.EventTypeWarning
}

func (it StatusCheckFailureThresholdExceed) Reason() string {
	return string(v1alpha1.StatusCheckFailureThresholdExceed)
}

func (it StatusCheckFailureThresholdExceed) Message() string {
	return "failure threshold exceed"
}

type StatusCheckSuccessThresholdExceed struct {
}

func (it StatusCheckSuccessThresholdExceed) Type() string {
	return corev1.EventTypeNormal
}

func (it StatusCheckSuccessThresholdExceed) Reason() string {
	return string(v1alpha1.StatusCheckSuccessThresholdExceed)
}

func (it StatusCheckSuccessThresholdExceed) Message() string {
	return "success threshold exceed"
}

func init() {
	register(
		StatusCheckCompleted{},
		StatusCheckExecutionFailed{},
		StatusCheckDurationExceed{},
		StatusCheckFailureThresholdExceed{},
		StatusCheckSuccessThresholdExceed{},
	)
}
