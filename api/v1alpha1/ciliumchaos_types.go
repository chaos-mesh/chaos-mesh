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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +chaos-mesh:experiment

// CiliumChaos is the control script's spec.
type CiliumChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a cilium chaos experiment
	Spec CiliumChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment
	Status CiliumChaosStatus `json:"status,omitempty"`
}

var _ InnerObject = (*CiliumChaos)(nil)

// CiliumChaosSpec defines the attributes that a user creates on a chaos experiment affecting cilium CNI.
type CiliumChaosSpec struct {
	NodeSelector `json:",inline"`

	// Duration represents the duration of the chaos action.
	Duration *string `json:"duration" webhook:"Duration"`

	// CiliumPodSelector provides a custom selector to find the cilium-agent pod for the node.
	//
	// If not specified, it will default to selecting pods from the `kube-system` namespace with labels
	// `app.kubernetes.io/name=cilium-agent` and `app.kubernetes.io/part-of=cilium` (which are used by default by the cilium
	// helm chart)
	CiliumPodSelector *CiliumPodSelectorSpec `json:"ciliumPodSelector"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

type CiliumPodSelectorSpec struct {
	// Namespace to restrict cilium-agent pod selection to
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Map of label selector expressions that can be used to select the cilium-agent pods..
	LabelSelectors map[string]string `json:"labelSelectors,omitempty"`
}

// CiliumChaosStatus represents the current status of the chaos experiment about pods.
type CiliumChaosStatus struct {
	ChaosStatus `json:",inline"`
}

func (obj *CiliumChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.NodeSelector,
	}
}
