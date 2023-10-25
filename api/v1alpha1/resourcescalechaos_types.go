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

type ResourceType string

const (
	ResourceTypeDaemonSet   ResourceType = "daemonset"
	ResourceTypeDeployment  ResourceType = "deployment"
	ResourceTypeReplicaSet  ResourceType = "replicaset"
	ResourceTypeStatefulSet ResourceType = "statefulset"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment

// ResourceScaleChaos is the Schema for the Kubernetes Chaos API
type ResourceScaleChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec ResourceScaleChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status ResourceScaleChaosStatus `json:"status,omitempty"`
}

// ResourceScaleChaosSpec defines the desired state of ResourceScaleChaos
type ResourceScaleChaosSpec struct {
	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration" webhook:"Duration"`

	ResourceScaleSelector `json:",inline"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

type ResourceScaleSelector struct {
	// Namespace defines the namespace of the ResourceScale.
	Namespace string `json:"namespace,omitempty"`

	// Name defines the name of the ResourceScale.
	Name string `json:"name,omitempty"`

	// Type of resource to scale
	ResourceType ResourceType `json:"resourceType,omitempty"`

	// ApplyReplicas is the amount of replicas the scale, defaults to 0
	// +optional
	ApplyReplicas *int32 `json:"applyReplicas,omitempty"`

	// RecoverReplicas is the amount of replicas the resources needs to scale to on recovery, defaults to initial replicas before applying chaos
	// +optional
	RecoverReplicas *int32 `json:"recoverReplicas,omitempty"`
}

// ResourceScaleChaosStatus defines the observed state of ResourceScaleChaos
type ResourceScaleChaosStatus struct {
	ChaosStatus `json:",inline"`
}

func (obj *ResourceScaleChaos) GetSelectorSpecs() map[string]interface{} {
	switch obj.Spec.ResourceType {
	case ResourceTypeDaemonSet, ResourceTypeDeployment, ResourceTypeReplicaSet:
		return map[string]interface{}{
			".": &obj.Spec.ResourceScaleSelector,
		}
	}
	return nil
}

func (selector *ResourceScaleSelector) Id() string {
	json, _ := json.Marshal(selector)
	return string(json)
}
