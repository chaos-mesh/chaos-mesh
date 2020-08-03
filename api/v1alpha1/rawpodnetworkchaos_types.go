// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KindRawPodNetworkChaos is the kind for network chaos
const KindRawPodNetworkChaos = "RawPodNetworkChaos"

// +kubebuilder:object:root=true

// RawPodNetworkChaos is the Schema for the networkchaos API
type RawPodNetworkChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec RawPodNetworkChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment about pods
	Status RawPodNetworkChaosStatus `json:"status"`
}

// RawPodNetworkChaosSpec defines the desired state of RawPodNetworkChaos
type RawPodNetworkChaosSpec struct {
	// The ipset on the pod
	// +optional
	IpSets []RawIpSet `json:"ipsets,omitempty"`

	// The iptables rules on the pod
	// +optional
	Iptables []RawIpTables `json:"iptables,omitempty"`

	// The qdisc rules on the pod
	// +optional
	Qdiscs []RawQdisc `json:"qdiscs,omitempty"`
}

// RawIpSet represents an ipset on specific pod
type RawIpSet struct {
	// The name of ipset
	Name string `json:"name"`

	// The contents of ipset
	Cidrs []string `json:"items"`

	// The name and namespace of the source network chaos
	RawRuleSource `json:",inline"`
}

// RawIpTables represents the iptables rules on specific pod
type RawIpTables struct {
	// The name of related ipset
	IpSet string `json:"ipset"`

	// The block direction of this iptables rule
	Direction Direction `json:"direction"`

	RawRuleSource `json:",inline"`
}

// QdiscType the type of a qdisc
type QdiscType string

const (
	// Netem represents netem qdisc
	Netem QdiscType = "netem"

	// Bandwidth represents bandwidth shape qdisc
	Bandwidth QdiscType = "bandwidth"
)

// RawQdisc represents the qdiscs on specific pod
type RawQdisc struct {
	// The type of qdisc
	Type QdiscType `json:"type"`

	Parameters QdiscParameter `json:"parameters"`

	// The name and namespace of the source network chaos
	Source string `json:"source"`
}

// QdiscParameter represents the parameters for a qdisc
type QdiscParameter struct {
	// Delay represents the detail about delay action
	// +optional
	Delay *DelaySpec `json:"delay,omitempty"`

	// Loss represents the detail about loss action
	Loss *LossSpec `json:"loss,omitempty"`

	// DuplicateSpec represents the detail about loss action
	Duplicate *DuplicateSpec `json:"duplicate,omitempty"`

	// Corrupt represents the detail about corrupt action
	Corrupt *CorruptSpec `json:"corrupt,omitempty"`

	// Bandwidth represents the detail about bandwidth control action
	// +optional
	Bandwidth *BandwidthSpec `json:"bandwidth,omitempty"`
}

// RawRuleSource represents the name and namespace of the source network chaos
type RawRuleSource struct {
	Source string `json:"source"`
}

// RawPodNetworkChaosStatus defines the observed state of RawPodNetworkChaos
type RawPodNetworkChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// RawPodNetworkChaosList contains a list of NetworkChaos
type RawPodNetworkChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RawPodNetworkChaos `json:"items"`
}

// GetStatus returns the status of chaos
func (in *RawPodNetworkChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetChaos returns a chaos instance
func (in *RawPodNetworkChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindNetworkChaos,
		StartTime: in.CreationTimestamp.Time,
		Status:    string(in.GetStatus().Experiment.Phase),
	}
	if in.DeletionTimestamp != nil {
		instance.EndTime = in.DeletionTimestamp.Time
	}
	return instance
}

// ListChaos returns a list of network chaos
func (in *RawPodNetworkChaosList) ListChaos() []*ChaosInstance {
	if len(in.Items) == 0 {
		return nil
	}
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

func init() {
	SchemeBuilder.Register(&RawPodNetworkChaos{}, &RawPodNetworkChaosList{})
}
