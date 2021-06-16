// Copyright 2021 Chaos Mesh Authors.
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
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +chaos-mesh:base
// +chaos-mesh:oneshot=in.Spec.Action==NodeReset

// GcpChaos is the Schema for the gcpchaos API
type GcpChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpChaosSpec   `json:"spec"`
	Status GcpChaosStatus `json:"status,omitempty"`
}

// GcpChaosAction represents the chaos action about gcp.
type GcpChaosAction string

const (
	// NodeStop represents the chaos action of stopping the node.
	NodeStop GcpChaosAction = "node-stop"
	// NodeReset represents the chaos action of resetting the node.
	NodeReset GcpChaosAction = "node-reset"
	// DiskLoss represents the chaos action of detaching the disk.
	DiskLoss GcpChaosAction = "disk-loss"
)

// GcpChaosSpec is the content of the specification for a GcpChaos
type GcpChaosSpec struct {
	// Action defines the specific gcp chaos action.
	// Supported action: node-stop / node-reset / disk-loss
	// Default action: node-stop
	// +kubebuilder:validation:Enum=node-stop;node-reset;disk-loss
	Action GcpChaosAction `json:"action"`

	// Duration represents the duration of the chaos action.
	// +optional
	Duration *string `json:"duration,omitempty"`

	// SecretName defines the name of kubernetes secret. It is used for GCP credentials.
	// +optional
	SecretName *string `json:"secretName,omitempty"`

	GcpSelector `json:",inline"`
}

type GcpSelector struct {
	// Project defines the name of gcp project.
	Project string `json:"project"`

	// Zone defines the zone of gcp project.
	Zone string `json:"zone"`

	// Instance defines the name of the instance
	Instance string `json:"instance"`

	// The device name of disks to detach.
	// Needed in disk-loss.
	// +optional
	DeviceNames *[]string `json:"deviceNames,omitempty"`
}

func (obj *GcpChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.GcpSelector,
	}
}

func (selector *GcpSelector) Id() string {
	// TODO: handle the error here
	// or ignore it is enough ?
	json, _ := json.Marshal(selector)

	return string(json)
}

// GcpChaosStatus represents the status of a GcpChaos
type GcpChaosStatus struct {
	ChaosStatus `json:",inline"`

	// The attached disk info strings.
	// Needed in disk-loss.
	AttachedDisksStrings []string `json:"attachedDiskStrings,omitempty"`
}

func (obj *GcpChaos) GetCustomStatus() interface{} {
	return &obj.Status.AttachedDisksStrings
}
