// Copyright 2019 PingCAP, Inc.
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

// ChaosAction represents the chaos action about pods.
type NetworkChaosAction string

const (
	// DelayAction represents the chaos action of adding delay on pods.
	DelayAction NetworkChaosAction = "delay"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkChaos is the control script's spec.
type NetworkChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a pod chaos experiment
	Spec NetworkChaosSpec `json:"spec"`

	// Most recently observed status of the chaos experiment about pods
	Status NetworkChaosStatus `json:"status"`
}

// NetworkChaosSpec defines the attributes that a user creates on a chaos experiment about network.
type NetworkChaosSpec struct {
	// Action defines the specific network chaos action.
	// Supported action: delay
	// Default action: delay
	Action NetworkChaosAction `json:"action"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Duration represents the duration of the chaos action
	Duration string `json:"duration"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about network.
	Scheduler SchedulerSpec `json:"scheduler"`

	// Delay represetns the detail about delay action
	// +optional
	Delay *DelaySpec `json:"delay"`
}

// DelaySpec defines detail of a delay action
type DelaySpec struct {
	Latency     string  `json:"latency"`
	Correlation float32 `json:"correlation"`
	Jitter      string  `json:"jitter"`
}

// NetworkChaosStatus represents the current status of the chaos experiment about network.
type NetworkChaosStatus struct {
	// TODO: add status
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkChaosList is NetworkChaos list.
type NetworkChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NetworkChaos `json:"items"`
}
