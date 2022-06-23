// Copyright 2022 Chaos Mesh Authors.
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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment

// DiskChaos is the Schema for the diskchaos API
type DiskChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a disk chaos experiment
	Spec DiskChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the disk chaos experiment
	Status DiskChaosStatus `json:"status"`
}

var _ InnerObjectWithSelector = (*DiskChaos)(nil)
var _ InnerObject = (*DiskChaos)(nil)

type DiskChaosAction string

const (
	DFill  DiskChaosAction = "fill"
	DWrite DiskChaosAction = "write"
	DRead  DiskChaosAction = "read"
)

type DiskChaosSpec struct {
	ContainerSelector `json:",inline"`

	// +kubebuilder:validation:Enum=fill;write;read
	Action DiskChaosAction `json:"action" webhook:"DiskAction"`

	Path    string `json:"path,omitempty"`
	Size    string `json:"size,omitempty"`
	Percent string `json:"percent,omitempty"`

	// SpaceLockSize keeps a part of disk space before disk chaos execute and
	// delete at first when we recover the disk chaos.
	SpaceLockSize string `json:"space_lock_size,omitempty"`

	FillByFAllocate bool `json:"fill_by_fallocate,omitempty"`

	ProcessNum uint8 `json:"process_num,omitempty" webhook:"ProcessNum"`
	// Not implement.
	LoopExecution bool `json:"loop_execution,omitempty"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`
}

type DiskChaosStatus struct {
	ChaosStatus `json:",inline"`
}

func (in *DiskChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &in.Spec.ContainerSelector,
	}
}
