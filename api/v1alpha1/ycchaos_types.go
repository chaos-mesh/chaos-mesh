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

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="action",type=string,JSONPath=`.spec.action`
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment
// +chaos-mesh:oneshot=in.Spec.Action==ComputeRestart
// +genclient

// YCChaos is the Schema for the ycchaos API
type YCChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   YCChaosSpec   `json:"spec"`
	Status YCChaosStatus `json:"status,omitempty"`
}

var _ InnerObjectWithSelector = (*YCChaos)(nil)
var _ InnerObject = (*YCChaos)(nil)

// YCChaosAction represents the chaos action about yc.
type YCChaosAction string

const (
	// ComputeStop represents the chaos action of stopping compute.
	ComputeStop YCChaosAction = "compute-stop"
	// ComputeRestart represents the chaos action of restarting compute.
	ComputeRestart YCChaosAction = "compute-restart"
)

// YCChaosSpec is the content of the specification for an YCChaos
type YCChaosSpec struct {
	// Action defines the specific yc chaos action.
	// Supported action: compute-stop / compute-restart
	// Default action: compute-stop
	// +kubebuilder:validation:Enum=compute-stop;compute-restart
	Action YCChaosAction `json:"action"`

	// Duration represents the duration of the chaos action.
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	// SecretName defines the name of kubernetes secret.
	// +optional
	SecretName *string `json:"secretName,omitempty" webhook:",nilable"`

	YCSelector `json:",inline"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

// YCChaosStatus represents the status of an YCChaos
type YCChaosStatus struct {
	ChaosStatus `json:",inline"`
}

type YCSelector struct {
	// ComputeInstance indicates the ID of the compute instance.
	ComputeInstance string `json:"computeInstance"`
}

func (obj *YCChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.YCSelector,
	}
}

func (selector *YCSelector) Id() string {
	json, _ := json.Marshal(selector)

	return string(json)
}
