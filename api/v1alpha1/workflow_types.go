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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=wf
// +kubebuilder:subresource:status
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a workflow
	Spec WorkflowSpec `json:"spec"`

	// +optional
	// Most recently observed status of the workflow
	Status WorkflowStatus `json:"status"`
}

type WorkflowSpec struct {
	Entry     string     `json:"entry"`
	Templates []Template `json:"templates"`
}

type WorkflowStatus struct {
	// +optional
	EntryNode *string `json:"entryNode,omitempty"`
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
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
	Tasks    []string     `json:"tasks,omitempty"`
	// +optional
	*EmbedChaos `json:",inline"`
}

// +kubebuilder:object:root=true
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
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
