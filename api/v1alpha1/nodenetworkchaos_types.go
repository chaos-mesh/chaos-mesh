// Copyright 2026 Chaos Mesh Authors.
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

const KindNodeNetworkChaos = "NodeNetworkChaos"

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// NodeNetworkChaos holds the network chaos rules applied in the host network
// namespace of a node. Its name is the name of the node.
type NodeNetworkChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the rules applied in the host network namespace
	Spec NodeNetworkChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the rules on the node
	Status NodeNetworkChaosStatus `json:"status,omitempty"`
}

// NodeNetworkChaosSpec defines the desired state of NodeNetworkChaos
type NodeNetworkChaosSpec struct {
	// The ipset in the host network namespace
	// +optional
	IPSets []RawIPSet `json:"ipsets,omitempty"`

	// The iptables rules in the host network namespace
	// +optional
	Iptables []RawIptables `json:"iptables,omitempty"`
}

// NodeNetworkChaosStatus defines the observed state of NodeNetworkChaos
type NodeNetworkChaosStatus struct {
	FailedMessage string `json:"failedMessage,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// NodeNetworkChaosList contains a list of NodeNetworkChaos
type NodeNetworkChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeNetworkChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeNetworkChaos{}, &NodeNetworkChaosList{})
}
