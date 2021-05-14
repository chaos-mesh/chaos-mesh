// Copyright 2021 Chaos Mesh Authors.
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

package controllers

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func SetCondition(status *v1alpha1.WorkflowNodeStatus, condition v1alpha1.WorkflowNodeCondition) {
	currentCond := GetCondition(*status, condition.Type)
	if currentCond != nil && currentCond.Status == condition.Status && currentCond.Reason == condition.Reason {
		return
	}
	newConditions := filterOutCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, condition)
}

func GetCondition(status v1alpha1.WorkflowNodeStatus, conditionType v1alpha1.WorkflowNodeConditionType) *v1alpha1.WorkflowNodeCondition {
	for _, item := range status.Conditions {
		if item.Type == conditionType {
			return &item
		}
	}
	return nil
}

func ConditionEqualsTo(status v1alpha1.WorkflowNodeStatus, conditionType v1alpha1.WorkflowNodeConditionType, expected corev1.ConditionStatus) bool {
	condition := GetCondition(status, conditionType)
	if condition == nil {
		return false
	}
	return condition.Status == expected
}

func filterOutCondition(conditions []v1alpha1.WorkflowNodeCondition, except v1alpha1.WorkflowNodeConditionType) []v1alpha1.WorkflowNodeCondition {
	var newConditions []v1alpha1.WorkflowNodeCondition
	for _, c := range conditions {
		if c.Type == except {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}

func WorkflowNodeFinished(status v1alpha1.WorkflowNodeStatus) bool {
	return ConditionEqualsTo(status, v1alpha1.ConditionAccomplished, corev1.ConditionTrue) ||
		ConditionEqualsTo(status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionTrue)
}
