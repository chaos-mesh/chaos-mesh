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
