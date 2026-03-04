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
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +chaos-mesh:base
// RemoteCluster defines a remote cluster
type RemoteCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RemoteClusterSpec `json:"spec,omitempty"`

	// +optional
	Status RemoteClusterStatus `json:"status,omitempty"`
}

// RemoteClusterSpec defines the specification of a remote cluster
type RemoteClusterSpec struct {
	Namespace string `json:"namespace"`
	Version   string `json:"version"`

	KubeConfig RemoteClusterKubeConfig `json:"kubeConfig"`

	// +optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	ConfigOverride json.RawMessage `json:"configOverride,omitempty"`
}

// RemoteClusterKubeConfig refers to a secret by which we'll use to connect remote cluster
type RemoteClusterKubeConfig struct {
	SecretRef RemoteClusterSecretRef `json:"secretRef"`
}

// RemoteClusterSecretRef refers to a secret in any namespaces
type RemoteClusterSecretRef struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`

	Key string `json:"key"`
}

type RemoteClusterStatus struct {
	CurrentVersion string `json:"currentVersion"`

	// Conditions represents the current condition of the remote cluster
	// +optional
	Conditions         []RemoteClusterCondition `json:"conditions,omitempty"`
	ObservedGeneration int64                    `json:"observedGeneration,omitempty"`
}

// AnnotationManagedHelmLifecycle is the annotation key used to control whether the
// RemoteCluster controller manages the Helm lifecycle (install/upgrade/uninstall) of
// Chaos Mesh on the remote cluster. When set to "false", the controller skips all Helm
// operations and only performs controller coordination. This is useful for air-gapped
// environments or when Chaos Mesh is pre-installed and managed independently.
// Default behavior (annotation absent or any value other than "false"): Helm lifecycle is managed.
//
// When Helm lifecycle is disabled, the kubeconfig secret must still grant the following
// minimum RBAC permissions on the remote cluster for controller coordination:
//   - chaos-mesh.org chaos resources (all 14 types): get, list, watch, create, update, patch, delete
//   - chaos-mesh.org chaos resources /status subresource: get, update, patch
//   - core pods: get, list, watch
const AnnotationManagedHelmLifecycle = "chaos-mesh.org/managed-helm-lifecycle"

type RemoteClusterConditionType string

var (
	RemoteClusterConditionInstalled RemoteClusterConditionType = "Installed"
	RemoteClusterConditionReady     RemoteClusterConditionType = "Ready"
)

type RemoteClusterCondition struct {
	Type   RemoteClusterConditionType `json:"type"`
	Status corev1.ConditionStatus     `json:"status"`
	// +optional
	Reason string `json:"reason"`
}

// RemoteClusterList contains a list of RemoteCluster
// +kubebuilder:object:root=true
type RemoteClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteCluster `json:"items"`
}
