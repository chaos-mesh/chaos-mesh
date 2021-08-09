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
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

func SetWorkflowCondition(status *v1alpha1.WorkflowStatus, condition v1alpha1.WorkflowCondition) {
	currentCond := GetWorkflowCondition(*status, condition.Type)
	if currentCond != nil && currentCond.Status == condition.Status && currentCond.Reason == condition.Reason {
		return
	}
	newConditions := filterOutWorkflowCondition(status.Conditions, condition.Type)
	status.Conditions = append(newConditions, condition)
}

func GetWorkflowCondition(status v1alpha1.WorkflowStatus, conditionType v1alpha1.WorkflowConditionType) *v1alpha1.WorkflowCondition {
	for _, item := range status.Conditions {
		if item.Type == conditionType {
			return &item
		}
	}
	return nil
}

func WorkflowConditionEqualsTo(status v1alpha1.WorkflowStatus, conditionType v1alpha1.WorkflowConditionType, expected corev1.ConditionStatus) bool {
	condition := GetWorkflowCondition(status, conditionType)
	if condition == nil {
		return false
	}
	return condition.Status == expected
}

func filterOutWorkflowCondition(conditions []v1alpha1.WorkflowCondition, except v1alpha1.WorkflowConditionType) []v1alpha1.WorkflowCondition {
	var newConditions []v1alpha1.WorkflowCondition
	for _, c := range conditions {
		if c.Type == except {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}

type SortByCreationTimestamp []v1alpha1.WorkflowNode

func (it SortByCreationTimestamp) Len() int {
	return len(it)
}

func (it SortByCreationTimestamp) Less(i, j int) bool {
	return it[j].GetCreationTimestamp().After(it[i].GetCreationTimestamp().Time)
}

func (it SortByCreationTimestamp) Swap(i, j int) {
	it[i], it[j] = it[j], it[i]
}

type ChildNodesFetcher struct {
	kubeClient client.Client
	logger     logr.Logger
}

func NewChildNodesFetcher(kubeClient client.Client, logger logr.Logger) *ChildNodesFetcher {
	return &ChildNodesFetcher{kubeClient: kubeClient, logger: logger}
}

// fetchChildNodes will return children workflow nodes controlled by given node
// Should only be used with Parallel and Serial Node
func (it *ChildNodesFetcher) fetchChildNodes(ctx context.Context, node v1alpha1.WorkflowNode) (activeChildNodes []v1alpha1.WorkflowNode, finishedChildNodes []v1alpha1.WorkflowNode, err error) {
	childNodes := v1alpha1.WorkflowNodeList{}
	controlledByThisNode, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			v1alpha1.LabelControlledBy: node.Name,
		},
	})

	if err != nil {
		it.logger.Error(err, "failed to build label selector with filtering children workflow node controlled by current node",
			"current node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return nil, nil, err
	}

	err = it.kubeClient.List(ctx, &childNodes, &client.ListOptions{
		Namespace:     node.Namespace,
		LabelSelector: controlledByThisNode,
	})

	if err != nil {
		it.logger.Error(err, "failed to list children workflow node controlled by current node",
			"current node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return nil, nil, err
	}

	sortedChildNodes := SortByCreationTimestamp(childNodes.Items)
	sort.Sort(sortedChildNodes)

	it.logger.V(4).Info("list children node", "current node",
		"current node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		len(sortedChildNodes), "children", sortedChildNodes)

	var activeChildren []v1alpha1.WorkflowNode
	var finishedChildren []v1alpha1.WorkflowNode

	for _, item := range sortedChildNodes {
		childNode := item
		if WorkflowNodeFinished(childNode.Status) {
			finishedChildren = append(finishedChildren, childNode)
		} else {
			activeChildren = append(activeChildren, childNode)
		}
	}
	return activeChildren, finishedChildren, nil
}

func getTaskNameFromGeneratedName(generatedNodeName string) string {
	index := strings.LastIndex(generatedNodeName, "-")
	if index < 0 {
		return generatedNodeName
	}
	return generatedNodeName[:index]
}

// setDifference return the set of elements which contained in former but not in latter
func setDifference(former []string, latter []string) []string {
	var result []string
	formerSet := make(map[string]struct{})
	latterSet := make(map[string]struct{})

	for _, item := range former {
		formerSet[item] = struct{}{}
	}
	for _, item := range latter {
		latterSet[item] = struct{}{}
	}
	for k := range formerSet {
		if _, ok := latterSet[k]; !ok {
			result = append(result, k)
		}
	}
	return result
}
