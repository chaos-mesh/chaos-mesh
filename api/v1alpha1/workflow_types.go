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
// +kubebuilder:resource:shortName=wf
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec WorkflowSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status WorkflowStatus `json:"status"`
}

// TODO: code generation
type EmbedChaos struct {
	// +optional
	DNSChaos *DNSChaosSpec `json:"dns_chaos,omitempty"`
	// +optional
	HTTPChaos *HTTPChaosSpec `json:"http_chaos,omitempty"`
	// +optional
	IoChaos *IoChaosSpec `json:"io_chaos,omitempty"`
	// +optional
	JVMChaos *JVMChaosSpec `json:"jvm_chaos,omitempty"`
	// +optional
	KernelChaos *KernelChaosSpec `json:"kernel_chaos,omitempty"`
	// +optional
	NetworkChaos *NetworkChaosSpec `json:"network_chaos,omitempty"`
	// +optional
	PodChaos *PodChaosSpec `json:"pod_chaos,omitempty"`
	// +optional
	StressChaos *StressChaosSpec `json:"stress_chaos,omitempty"`
	// +optional
	TimeChaos *TimeChaosSpec `json:"time_chaos,omitempty"`
}

type WorkflowSpec struct {
	Entry     string     `json:"entry"`
	Templates []Template `json:"templates"`
}

type WorkflowStatus struct {
	// +optional
	EntryNode *string `json:"entry_node,omitempty"`
	// +optional
	Nodes []corev1.LocalObjectReference `json:"nodes,omitempty"`
}

type TemplateType string

const (
	TypeTask         TemplateType = "Task"
	TypeSerial       TemplateType = "Serial"
	TypeParallel     TemplateType = "Parallel"
	TypeSuspend      TemplateType = "Suspend"
	TypeIoChaos      TemplateType = "IoChaos"
	TypeNetworkChaos TemplateType = "NetworkChaos"
	TypeStressChaos  TemplateType = "StressChaos"
	TypePodChaos     TemplateType = "PodChaos"
	TypeTimeChaos    TemplateType = "TimeChaos"
	TypeKernelChaos  TemplateType = "KernelChaos"
	TypeDnsChaos     TemplateType = "DnsChaos"
	TypeHttpChaos    TemplateType = "HttpChaos"
	TypeJvmChaos     TemplateType = "JvmChaos"
)

type Template struct {
	Name     string       `json:"name"`
	Type     TemplateType `json:"template_type"`
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
