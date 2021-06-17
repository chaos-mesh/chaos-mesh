// Copyright 2020 Chaos Mesh Authors.
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
// +chaos-mesh:oneshot=in.Spec.Action==Ec2Restart

// AwsChaos is the Schema for the awschaos API
type AwsChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsChaosSpec   `json:"spec"`
	Status AwsChaosStatus `json:"status,omitempty"`
}

// AwsChaosAction represents the chaos action about aws.
type AwsChaosAction string

const (
	// Ec2Stop represents the chaos action of stopping ec2.
	Ec2Stop AwsChaosAction = "ec2-stop"
	// Ec2Restart represents the chaos action of restarting ec2.
	Ec2Restart AwsChaosAction = "ec2-restart"
	// DetachVolume represents the chaos action of detaching the volume of ec2.
	DetachVolume AwsChaosAction = "detach-volume"
)

// AwsChaosSpec is the content of the specification for an AwsChaos
type AwsChaosSpec struct {
	// Action defines the specific aws chaos action.
	// Supported action: ec2-stop / ec2-restart / detach-volume
	// Default action: ec2-stop
	// +kubebuilder:validation:Enum=ec2-stop;ec2-restart;detach-volume
	Action AwsChaosAction `json:"action"`

	// Duration represents the duration of the chaos action.
	// +optional
	Duration *string `json:"duration,omitempty"`

	// SecretName defines the name of kubernetes secret.
	// +optional
	SecretName *string `json:"secretName,omitempty"`

	AwsSelector `json:",inline"`
}

// AwsChaosStatus represents the status of an AwsChaos
type AwsChaosStatus struct {
	ChaosStatus `json:",inline"`
}

type AwsSelector struct {
	// TODO: it would be better to split them into multiple different selector and implementation
	// but to keep the minimal modification on current implementation, it hasn't been splited.

	// Endpoint indicates the endpoint of the aws server. Just used it in test now.
	// +optional
	Endpoint *string `json:"endpoint,omitempty"`

	// AwsRegion defines the region of aws.
	AwsRegion string `json:"awsRegion"`

	// Ec2Instance indicates the ID of the ec2 instance.
	Ec2Instance string `json:"ec2Instance"`

	// EbsVolume indicates the ID of the EBS volume.
	// Needed in detach-volume.
	// +optional
	EbsVolume *string `json:"volumeID,omitempty"`

	// DeviceName indicates the name of the device.
	// Needed in detach-volume.
	// +optional
	DeviceName *string `json:"deviceName,omitempty"`
}

func (obj *AwsChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.AwsSelector,
	}
}

func (selector *AwsSelector) Id() string {
	// TODO: handle the error here
	// or ignore it is enough ?
	json, _ := json.Marshal(selector)

	return string(json)
}
