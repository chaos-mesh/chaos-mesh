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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=wf
// +kubebuilder:subresource:status
// +chaos-mesh:base
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a workflow
	Spec WorkflowSpec `json:"spec"`

	// +optional
	// Most recently observed status of the workflow
	Status WorkflowStatus `json:"status"`
}

func (in *Workflow) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindTimeChaos,
		StartTime: in.CreationTimestamp.Time,
		Action:    "",
		UID:       string(in.UID),
	}

	if in.DeletionTimestamp != nil {
		instance.EndTime = in.DeletionTimestamp.Time
	}
	return instance
}

func (in *Workflow) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

const KindWorkflow = "Workflow"

type WorkflowSpec struct {
	Entry     string     `json:"entry"`
	Templates []Template `json:"templates"`
}

type WorkflowStatus struct {
	// +optional
	EntryNode *string `json:"entryNode,omitempty"`
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	EndTime *metav1.Time `json:"endTime,omitempty"`
	// Represents the latest available observations of a workflow's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []WorkflowCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

type WorkflowConditionType string

const (
	WorkflowConditionAccomplished WorkflowConditionType = "Accomplished"
	WorkflowConditionScheduled    WorkflowConditionType = "Scheduled"
)

type WorkflowCondition struct {
	Type      WorkflowConditionType  `json:"type"`
	Status    corev1.ConditionStatus `json:"status"`
	Reason    string                 `json:"reason"`
	StartTime *metav1.Time           `json:"startTime,omitempty"`
}

type TemplateType string

const (
	TypeTask     TemplateType = "Task"
	TypeSerial   TemplateType = "Serial"
	TypeParallel TemplateType = "Parallel"
	TypeSuspend  TemplateType = "Suspend"
)

func IsChaosTemplateType(target TemplateType) bool {
	return contains(allChaosTemplateType, target)
}

func contains(arr []TemplateType, target TemplateType) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}

type Template struct {
	Name     string       `json:"name"`
	Type     TemplateType `json:"templateType"`
	Duration *string      `json:"duration,omitempty"`
	// +optional
	Task *Task `json:"task,omitempty"`
	// +optional
	Tasks []string `json:"tasks,omitempty"`
	// +optional
	ConditionalTasks []ConditionalTask `json:"conditionalTasks,omitempty"`
	// +optional
	*EmbedChaos `json:",inline"`
}

type Task struct {
	// Container is the main container image to run in the pod
	Container *corev1.Container `json:"container,omitempty"`

	// Volumes is a list of volumes that can be mounted by containers in a template.
	// +patchStrategy=merge
	// +patchMergeKey=name
	Volumes []corev1.Volume `json:"volumes,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// TODO: maybe we could specify parameters in other ways, like loading context from file
}

// +kubebuilder:object:root=true
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

func (in *WorkflowList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}

func FetchChaosByTemplateType(templateType TemplateType) (runtime.Object, error) {
	if kind, ok := all.kinds[string(templateType)]; ok {
		return kind.Chaos.DeepCopyObject(), nil
	}
	return nil, fmt.Errorf("no such kind refers to template type %s", templateType)
}
