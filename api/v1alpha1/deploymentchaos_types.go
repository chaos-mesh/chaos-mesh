// Copyright 2023 Chaos Mesh Authors.
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
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment

// DeploymentChaos is the Schema for the Kubernetes Chaos API
type DeploymentChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec DeploymentChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status DeploymentChaosStatus `json:"status,omitempty"`
}

// DeploymentChaosSpec defines the desired state of DeploymentChaos
type DeploymentChaosSpec struct {
	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	DeploymentSelector `json:",inline"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

type DeploymentSelector struct {
	// Namespace defines the namespace of the deployment.
	Namespace string `json:"namespace"`

	// Name defines the name of the deployment.
	Name string `json:"name"`
}

// DeploymentChaosStatus defines the observed state of DeploymentChaos
type DeploymentChaosStatus struct {
	ChaosStatus `json:",inline"`
}

func (obj *DeploymentChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.DeploymentSelector,
	}
}

func (selector *DeploymentSelector) Id() string {
	json, _ := json.Marshal(selector)
	return string(json)
}
