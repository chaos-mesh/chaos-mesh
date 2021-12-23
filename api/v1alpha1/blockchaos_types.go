// Copyright 2021 Chaos Mesh Authors.
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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="action",type=string,JSONPath=`.spec.action`
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment

// BlockChaos is the Schema for the blockchaos API
type BlockChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BlockChaosSpec   `json:"spec"`
	Status BlockChaosStatus `json:"status,omitempty"`
}

type BlockChaosAction string

const (
	BlockLimit BlockChaosAction = "limit"
	BlockDelay BlockChaosAction = "delay"
)

// BlockChaosSpec is the content of the specification for a BlockChaos
type BlockChaosSpec struct {
	// Action defines the specific block chaos action.
	// Supported action: limit / delay
	// +kubebuilder:validation:Enum=limit;delay
	Action BlockChaosAction `json:"action"`

	// IOPS defines the limit of IO frequency.
	IOPS int `json:"iops"`

	// Delay defines the latency of every io request.
	Delay string `json:"delay" webhook:"Duration"`

	// +optional
	Correlation string `json:"correlation,omitempty" default:"0" webhook:"FloatStr"`

	// +optional
	Jitter string `json:"jitter,omitempty" default:"0ms" webhook:"Duration"`

	NodePVSelector `json:",inline"`

	// Duration represents the duration of the chaos action.
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`
}

// NodePVSelector is the selector to select a node and a PV on it
type NodePVSelector struct {
	PodSelector `json:",inline"`

	VolumeName string `json:"volumeName"`
}

// BlockChaosStatus represents the status of a BlockChaos
type BlockChaosStatus struct {
	ChaosStatus `json:",inline"`

	// InjectionIds always specifies the number of injected chaos action
	// +optional
	InjectionIds map[string]int `json:"ids,omitempty"`
}

func (obj *BlockChaos) GetCustomStatus() interface{} {
	return &obj.Status.InjectionIds
}
