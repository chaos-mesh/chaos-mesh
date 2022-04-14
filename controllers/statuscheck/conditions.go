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

package statuscheck

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type conditionMap map[v1alpha1.StatusCheckConditionType]v1alpha1.StatusCheckCondition

func toConditionMap(conditions []v1alpha1.StatusCheckCondition) conditionMap {
	result := make(map[v1alpha1.StatusCheckConditionType]v1alpha1.StatusCheckCondition)
	for _, condition := range conditions {
		condition := condition
		result[condition.Type] = condition
	}
	return result
}

func toConditionList(conditions conditionMap) []v1alpha1.StatusCheckCondition {
	result := make([]v1alpha1.StatusCheckCondition, 0, len(conditions))
	for _, condition := range conditions {
		condition := condition
		result = append(result, condition)
	}
	return result
}

func (in conditionMap) setCondition(
	t v1alpha1.StatusCheckConditionType,
	status corev1.ConditionStatus,
	reason v1alpha1.StatusCheckReason) {
	condition, ok := in[t]
	now := &metav1.Time{Time: time.Now()}
	if !ok {
		in[t] = v1alpha1.StatusCheckCondition{
			Type:               t,
			Status:             status,
			Reason:             reason,
			LastProbeTime:      now,
			LastTransitionTime: now,
		}
	} else {
		condition.LastProbeTime = now
		condition.Reason = reason
		if status != condition.Status {
			condition.Status = status
			condition.LastTransitionTime = now
		}
		in[t] = condition
	}
}

func (in conditionMap) isCompleted() bool {
	cond, ok := in[v1alpha1.StatusCheckConditionCompleted]
	return ok && cond.Status == corev1.ConditionTrue
}

func (in conditionMap) isDurationExceed() bool {
	cond, ok := in[v1alpha1.StatusCheckConditionDurationExceed]
	return ok && cond.Status == corev1.ConditionTrue
}

func (in conditionMap) isFailureThresholdExceed() bool {
	cond, ok := in[v1alpha1.StatusCheckConditionFailureThresholdExceed]
	return ok && cond.Status == corev1.ConditionTrue
}

func (in conditionMap) isSuccessThresholdExceed() bool {
	cond, ok := in[v1alpha1.StatusCheckConditionSuccessThresholdExceed]
	return ok && cond.Status == corev1.ConditionTrue
}

func setDurationExceedCondition(statusCheck v1alpha1.StatusCheck, conditions conditionMap) error {
	now := time.Now()
	ok, _, err := statusCheck.DurationExceed(now)
	if err != nil {
		return err
	}
	if !ok {
		conditions.setCondition(v1alpha1.StatusCheckConditionDurationExceed, corev1.ConditionFalse, "")
	} else {
		conditions.setCondition(v1alpha1.StatusCheckConditionDurationExceed, corev1.ConditionTrue, "")
	}
	return nil
}

// setFailureThresholdExceedCondition check if the failure threshold is exceeded, and then set the condition into conditionMap
// Notice: the method `execute` of `worker` struct in `controllers/statuscheck/worker.go` checks the failure threshold,
// so if you want to modify the logic here, don't forget to modify that function as well.
func setFailureThresholdExceedCondition(statusCheck v1alpha1.StatusCheck, conditions conditionMap) {
	if isThresholdExceed(statusCheck.Status.Records, v1alpha1.StatusCheckOutcomeFailure, statusCheck.Spec.FailureThreshold) {
		conditions.setCondition(v1alpha1.StatusCheckConditionFailureThresholdExceed, corev1.ConditionTrue, "")
	} else {
		conditions.setCondition(v1alpha1.StatusCheckConditionFailureThresholdExceed, corev1.ConditionFalse, "")
	}
}

// setSuccessThresholdExceedCondition check if the success threshold is exceeded, and then set the condition into conditionMap
// Notice: the method `execute` of `worker` struct in `controllers/statuscheck/worker.go` checks the success threshold,
// so if you want to modify the logic here, don't forget to modify that function as well.
func setSuccessThresholdExceedCondition(statusCheck v1alpha1.StatusCheck, conditions conditionMap) {
	if isThresholdExceed(statusCheck.Status.Records, v1alpha1.StatusCheckOutcomeSuccess, statusCheck.Spec.SuccessThreshold) {
		conditions.setCondition(v1alpha1.StatusCheckConditionSuccessThresholdExceed, corev1.ConditionTrue, "")
	} else {
		conditions.setCondition(v1alpha1.StatusCheckConditionSuccessThresholdExceed, corev1.ConditionFalse, "")
	}
}

func setCompletedCondition(statusCheck v1alpha1.StatusCheck, conditions conditionMap) {
	if conditions.isDurationExceed() {
		conditions.setCondition(v1alpha1.StatusCheckConditionCompleted, corev1.ConditionTrue, v1alpha1.StatusCheckDurationExceed)
		return
	}
	if conditions.isFailureThresholdExceed() {
		conditions.setCondition(v1alpha1.StatusCheckConditionCompleted, corev1.ConditionTrue, v1alpha1.StatusCheckFailureThresholdExceed)
		return
	}
	if statusCheck.Spec.Mode == v1alpha1.StatusCheckSynchronous {
		if conditions.isSuccessThresholdExceed() {
			conditions.setCondition(v1alpha1.StatusCheckConditionCompleted, corev1.ConditionTrue, v1alpha1.StatusCheckSuccessThresholdExceed)
			return
		}
	}
	conditions.setCondition(v1alpha1.StatusCheckConditionCompleted, corev1.ConditionFalse, "")
}

func isThresholdExceed(records []v1alpha1.StatusCheckRecord, outcome v1alpha1.StatusCheckOutcome, threshold int) bool {
	count := 0
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].Outcome != outcome {
			return false
		}
		count++
		if count >= threshold {
			return true
		}
	}
	return false
}
