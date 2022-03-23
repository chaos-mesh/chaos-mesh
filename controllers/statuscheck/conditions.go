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
	result := make([]v1alpha1.StatusCheckCondition, len(conditions), 0)
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

func generateConditions(statusCheck v1alpha1.StatusCheck) ([]v1alpha1.StatusCheckCondition, error) {
	conditions := toConditionMap(statusCheck.Status.Conditions)

	if err := setDurationExceedCondition(statusCheck, conditions); err != nil {
		return nil, err
	}
	setFailureThresholdExceedCondition(statusCheck, conditions)
	setSuccessThresholdExceedCondition(statusCheck, conditions)

	// this condition must be placed after the above three conditions
	setCompletedCondition(statusCheck, conditions)

	return toConditionList(conditions), nil
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

func setFailureThresholdExceedCondition(statusCheck v1alpha1.StatusCheck, conditions conditionMap) {
	if isThresholdExceed(statusCheck.Status.Records, v1alpha1.StatusCheckOutcomeFailure, statusCheck.Spec.FailureThreshold) {
		conditions.setCondition(v1alpha1.StatusCheckConditionFailureThresholdExceed, corev1.ConditionTrue, "")
	} else {
		conditions.setCondition(v1alpha1.StatusCheckConditionFailureThresholdExceed, corev1.ConditionFalse, "")
	}
}

func setSuccessThresholdExceedCondition(statusCheck v1alpha1.StatusCheck, conditions conditionMap) {
	if isThresholdExceed(statusCheck.Status.Records, v1alpha1.StatusCheckOutcomeSuccess, statusCheck.Spec.FailureThreshold) {
		conditions.setCondition(v1alpha1.StatusCheckConditionSuccessThresholdExceed, corev1.ConditionTrue, "")
	} else {
		conditions.setCondition(v1alpha1.StatusCheckConditionSuccessThresholdExceed, corev1.ConditionFalse, "")
	}
}

func setCompletedCondition(statusCheck v1alpha1.StatusCheck, conditions conditionMap) {
	condition, ok := conditions[v1alpha1.StatusCheckConditionDurationExceed]
	if ok && condition.Status == corev1.ConditionTrue {
		conditions.setCondition(v1alpha1.StatusCheckConditionCompleted, corev1.ConditionTrue, "")
		return
	}
	condition, ok = conditions[v1alpha1.StatusCheckConditionFailureThresholdExceed]
	if ok && condition.Status == corev1.ConditionTrue {
		conditions.setCondition(v1alpha1.StatusCheckConditionCompleted, corev1.ConditionTrue, "")
		return
	}
	if statusCheck.Spec.Mode == v1alpha1.StatusCheckSynchronous {
		condition, ok = conditions[v1alpha1.StatusCheckConditionSuccessThresholdExceed]
		if ok && condition.Status == corev1.ConditionTrue {
			conditions.setCondition(v1alpha1.StatusCheckConditionCompleted, corev1.ConditionTrue, "")
			return
		}
	}
	conditions.setCondition(v1alpha1.StatusCheckConditionCompleted, corev1.ConditionFalse, "")
}

func isThresholdExceed(records []v1alpha1.StatusCheckRecord, want v1alpha1.StatusCheckOutcome, threshold int) bool {
	count := 0
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].Outcome != want {
			return false
		}
		count++
		if count >= threshold {
			return true
		}
	}
	return false
}
