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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type WorkflowNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec WorkflowNodeSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status WorkflowNodeStatus `json:"status"`
}

type WorkflowNodeSpec struct {
	TemplateName string       `json:"template_name"`
	WorkflowName string       `json:"workflow_name"`
	Type         TemplateType `json:"type"`
	StartTime    *metav1.Time `json:"start_time,omitempty"`
	Deadline     *metav1.Time `json:"deadline,omitempty"`
	Tasks        []string     `json:"tasks,omitempty"`
	EmbedChaos   `json:",inline"`
}

type WorkflowNodeStatus struct {

	// ExpectedChildren means the expected children to execute
	ExpectedChildren int `json:"expected_children"`

	// ChaosResource refs to the real chaos CR object.
	// +optional
	ChaosResource *corev1.TypedLocalObjectReference `json:"chaos_resource,omitempty"`

	// ActiveChildren means the created children node
	// +optional
	ActiveChildren []corev1.LocalObjectReference `json:"active_children"`

	// Children is necessary for representing the order when replicated child template references by parent template.
	// +optional
	FinishedChildren []corev1.LocalObjectReference `json:"finished_children,omitempty"`

	// Represents the latest available observations of a worklfow node's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []WorkflowNodeCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

type WorkflowNodeConditionType string

const (
	Accomplished   WorkflowNodeConditionType = "Accomplished"
	DeadlineExceed WorkflowNodeConditionType = "DeadlineExceed"
)

type WorkflowNodeCondition struct {
	Type   WorkflowNodeConditionType `json:"type"`
	Status corev1.ConditionStatus    `json:"status"`
	Reason string                    `json:"reason"`
}

// +kubebuilder:object:root=true
type WorkflowNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkflowNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkflowNode{}, &WorkflowNodeList{})
}

// Reasons
const (
	EntryCreated          string = "EntryCreated"
	InvalidEntry          string = "InvalidEntry"
	NodeAccomplished      string = "NodeAccomplished"
	NodeDeadlineExceed    string = "NodeDeadlineExceed"
	NodeDeadlineNotExceed string = "NodeDeadlineNotExceed"
	NodeDeadlineOmitted   string = "NodeDeadlineOmitted"
)
