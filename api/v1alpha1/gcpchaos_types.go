// Copyright 2020 Chaos Mesh Authors.
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +chaos-mesh:base

// GcpChaos is the Schema for the gcpchaos API
type GcpChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpChaosSpec   `json:"spec"`
	Status GcpChaosStatus `json:"status,omitempty"`
}

// GcpChaosAction represents the chaos action about gcp.
type GcpChaosAction string

const (
	// NodeStop represents the chaos action of stopping the node.
	NodeStop GcpChaosAction = "node-stop"
)

// GcpChaosSpec is the content of the specification for a GcpChaos
type GcpChaosSpec struct {
	// Action defines the specific gcp chaos action.
	// Supported action: node-stop / node-reset
	// Default action: node-stop
	// +kubebuilder:validation:Enum=node-stop;node-reset
	Action GcpChaosAction `json:"action"`

	// Duration represents the duration of the chaos action.
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	// +optional
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// SecretName defines the name of kubernetes secret.
	// +optional
	SecretName *string `json:"secretName,omitempty"`
}

// GcpChaosStatus represents the status of a GcpChaos
type GcpChaosStatus struct {
	ChaosStatus `json:",inline"`
}
