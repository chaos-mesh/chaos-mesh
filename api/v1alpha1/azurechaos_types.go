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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="action",type=string,JSONPath=`.spec.action`
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment
// +chaos-mesh:oneshot=in.Spec.Action==AzureVmRestart

// AzureChaos is the Schema for the azurechaos API
type AzureChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureChaosSpec   `json:"spec"`
	Status AzureChaosStatus `json:"status,omitempty"`
}

var _ InnerObjectWithSelector = (*AzureChaos)(nil)
var _ InnerObject = (*AzureChaos)(nil)

// AzureChaosAction represents the chaos action about azure.
type AzureChaosAction string

const (
	// AzureVmStop represents the chaos action of stopping vm.
	AzureVmStop AzureChaosAction = "vm-stop"
	// AzureVmRestart represents the chaos action of restarting vm.
	AzureVmRestart AzureChaosAction = "vm-restart"
	// AzureDiskDetach represents the chaos action of detaching the disk from vm.
	AzureDiskDetach AzureChaosAction = "disk-detach"
)

// AzureChaosSpec is the content of the specification for an AzureChaos
type AzureChaosSpec struct {
	// Action defines the specific azure chaos action.
	// Supported action: vm-stop / vm-restart / disk-detach
	// Default action: vm-stop
	// +kubebuilder:validation:Enum=vm-stop;vm-restart;disk-detach
	Action AzureChaosAction `json:"action"`

	// Duration represents the duration of the chaos action.
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	AzureSelector `json:",inline"`
}

// AzureChaosStatus represents the status of an AzureChaos
type AzureChaosStatus struct {
	ChaosStatus `json:",inline"`
}

type AzureSelector struct {
	// SubscriptionID defines the id of Azure subscription.
	SubscriptionID string `json:"subscriptionID"`

	// ResourceGroupName defines the name of ResourceGroup
	ResourceGroupName string `json:"resourceGroupName"`

	// VMName defines the name of Virtual Machine
	VMName string `json:"vmName"`

	// DiskName indicates the name of the disk.
	// Needed in disk-detach.
	// +optional
	DiskName *string `json:"diskName,omitempty" webhook:"DiskName,nilable"`

	// LUN indicates the Logical Unit Number of the data disk.
	// Needed in disk-detach.
	// +optional
	LUN *int `json:"lun,omitempty" webhook:"LUN,nilable"`

	// SecretName defines the name of kubernetes secret. It is used for Azure credentials.
	// +optional
	SecretName *string `json:"secretName,omitempty"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

func (obj *AzureChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.AzureSelector,
	}
}

func (selector *AzureSelector) Id() string {
	// TODO: handle the error here
	// or ignore it is enough ?
	json, _ := json.Marshal(selector)

	return string(json)
}
