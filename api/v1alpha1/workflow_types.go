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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	TypeSchedule TemplateType = "Schedule"
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
	Deadline *string      `json:"deadline,omitempty"`
	Tasks    []string     `json:"tasks,omitempty"`
	// +optional
	*EmbedChaos `json:",inline"`
	// +optional
	Schedule *ChaosOnlyScheduleSpec `json:"schedule,omitempty"`
}

// ChaosOnlyScheduleSpec is very similar with ScheduleSpec, but it could not schedule Workflow
// because we could not resolve nested CRD now
type ChaosOnlyScheduleSpec struct {
	Schedule string `json:"schedule"`

	// +optional
	// +nullable
	// +kubebuilder:validation:Minimum=0
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds"`

	// +optional
	// +kubebuilder:validation:Enum=Forbid;Allow
	ConcurrencyPolicy ConcurrencyPolicy `json:"concurrencyPolicy"`

	// +optional
	// +kubebuilder:validation:Minimum=1
	HistoryLimit int `json:"historyLimit,omitempty"`

	// TODO: use a custom type, as `TemplateType` contains other possible values
	Type ScheduleTemplateType `json:"type"`

	EmbedChaos `json:",inline"`
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
