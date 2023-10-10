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

// SelectorResourceType is the type of resource that is valid for rolling restart
type SelectorResourceType string

const (
	// DaemonSetResourceType represents the resource type daemonset
	DaemonSetResourceType SelectorResourceType = "daemonset"
	// DeploymentResource represents the resource type deployment
	DeploymentResourceType SelectorResourceType = "deployment"
	// StatefulSetResourceType represents the resource type statefulset
	StatefulSetResourceType SelectorResourceType = "statefulset"
)

// +kubebuilder:object:root=true
// +chaos-mesh:experiment
// +chaos-mesh:oneshot=true

// RollingRestartChaos is the Schema for the rolling restart API
type RollingRestartChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec RollingRestartChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status RollingRestartChaosStatus `json:"status,omitempty"`
}

// RollingRestartChaosSpec defines the desired state of RollingRestartChaos
type RollingRestartChaosSpec struct {
	RollingRestartSelector `json:",inline"`

	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

type RollingRestartSelector struct {
	// Namespace defines the namespace of the deployment.
	Namespace string `json:"namespace,omitempty"`

	// Name defines the name of the deployment.
	Name string `json:"name,omitempty"`

	// ResourceType defines the specific resource type to restart
	// Supported resource types: daemonset / deployment / statefulset
	// +kubebuilder:validation:Enum=daemonset;deployment;statefulset
	ResourceType SelectorResourceType `json:"resourceType,omitempty"`
}

// RollingRestartChaosStatus defines the observed state of RollingRestartChaos
type RollingRestartChaosStatus struct {
	ChaosStatus `json:",inline"`
}

func (obj *RollingRestartChaos) GetSelectorSpecs() map[string]interface{} {
	switch obj.Spec.ResourceType {
	case DaemonSetResourceType, DeploymentResourceType, StatefulSetResourceType:
		return map[string]interface{}{
			".": &obj.Spec.RollingRestartSelector,
		}
	}
	return nil
}

func (selector *RollingRestartSelector) Id() string {
	json, _ := json.Marshal(selector)
	return string(json)
}
