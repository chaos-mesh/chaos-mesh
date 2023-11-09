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

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=csvm
// +kubebuilder:printcolumn:name="action",type=string,JSONPath=`.spec.action`
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment
// +chaos-mesh:oneshot=in.Spec.Action==VMRestart

// CloudStackVMChaos is the Schema for the cloudstackchaos API.
type CloudStackVMChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudStackVMChaosSpec   `json:"spec"`
	Status CloudStackVMChaosStatus `json:"status,omitempty"`
}

var (
	_ InnerObjectWithSelector = (*CloudStackVMChaos)(nil)
	_ InnerObject             = (*CloudStackVMChaos)(nil)
)

// CloudStackVMChaosAction represents the chaos action about cloudstack.
type CloudStackVMChaosAction string

const (
	// VMStop represents the chaos action of stopping the VM.
	VMStop CloudStackVMChaosAction = "vm-stop"

	// VMRestart represents the chaos action of restarting the VM.
	VMRestart CloudStackVMChaosAction = "vm-restart"
)

// CloudStackVMChaosSpec is the content of the specification for a CloudStackChaos.
type CloudStackVMChaosSpec struct {
	// APIConfig defines the configuration ncessary to connect to the CloudStack API.
	APIConfig CloudStackAPIConfig `json:"apiConfig"`

	// Selector defines the parameters that can be used to select target VMs.
	Selector CloudStackVMChaosSelector `json:"selector"`

	// DryRun defines whether the chaos should run a dry-run mode.
	// +optional
	DryRun bool `json:"dryRun,omitempty"`

	// Action defines the specific cloudstack chaos action.
	// Supported action: vm-stop / vm-restart
	// Default action: vm-stop
	// +kubebuilder:validation:Enum=vm-stop;vm-restart
	Action CloudStackVMChaosAction `json:"action"`

	// Duration represents the duration of the chaos action.
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

type CloudStackAPIConfig struct {
	// Address defines the address of the CloudStack instsance.
	Address string `json:"address"`

	// VerifySSL defines whether certificates should be verified when connecting to the API.
	// +optional
	VerifySSL bool `json:"verifySSL,omitempty"`

	// SecretName defines the name of the secret where the API credentials are stored.
	SecretName string `json:"secretName"`

	// APIKeyField defines the key under which the value for API key is stored inside the secret.
	// +optional
	APIKeyField string `json:"apiKeyField,omitempty" default:"api-key"`

	// APISecretField defines the key under which the value for API secret is stored inside the secret.
	// +optional
	APISecretField string `json:"apiSecretField,omitempty" default:"api-secret"`
}

// CloudStackVMChaosStatus represents the status of a CloudStackChaos.
type CloudStackVMChaosStatus struct {
	ChaosStatus `json:",inline"`
}

type CloudStackVMChaosSelector struct {
	// Account defines account to list resources by. Must be used with the domainId parameter.
	// +optional
	Account *string `json:"account,omitempty"`

	// AffinityGroupID defines affinity group to list the VMs by.
	// +optional
	AffinityGroupID *string `json:"affinityGroupId,omitempty"`

	// DisplayVM defines a flag that indicates whether to list VMs by the display flag.
	// +optional
	DisplayVM bool `json:"displayVm,omitempty"`

	// DomainID defines domain ID the VMs belong to.
	// +optional
	DomainID *string `json:"domainId,omitempty"`

	// GroupID defines the ID of the group the VMs belong to.
	// +optional
	GroupID *string `json:"groupId,omitempty"`

	// HostID defines the ID of the host the VMs belong to.
	// +optional
	HostID *string `json:"hostId,omitempty"`

	// Hypervisor defines the target hypervisor.
	// +optional
	Hypervisor *string `json:"hypervisor,omitempty"`

	// ID defines the ID of the VM.
	// +optional
	ID *string `json:"id,omitempty"`

	// IDs defines a list of VM IDs, mutually exclusive with ID.
	// +optional
	IDs []string `json:"ids,omitempty"`

	// ISOID defines the ISO ID to list the VMs by.
	// +optional
	ISOID *string `json:"isoid,omitempty"`

	// IsRecursive defines whether VMs should be listed recursively from parent specified by DomainID.
	// +optional
	IsRecursive bool `json:"isRecursive,omitempty"`

	// KeyPair defines the SSH keypair name to list the VMs by.
	// +optional
	KeyPair *string `json:"keyPair,omitempty"`

	// Keyword defines the keyword to list the VMs by.
	// +optional
	Keyword *string `json:"keyword,omitempty"`

	// ListAll defines whether to list just the resources that belong to the caller or all the resources the caller is
	// authorised to see.
	// +optional
	ListAll bool `json:"listAll,omitempty"`

	// Name defines the name of the VM instance.
	// +optiional
	Name *string `json:"name,omitempty"`

	// NetworkID defines the ID of the network to list the VMs by.
	// +optional
	NetworkID *string `json:"networkId,omitempty"`

	// ProjectID defines the project ID to list the VMs by.
	// +optional
	ProjectID *string `json:"projectId,omitempty"`

	// ServiceOffering defines the service offering to list the VMs by.
	// +optional
	ServiceOffering *string `json:"serviceOffering,omitempty"`

	// State defines the state of the VM that should match.
	// +kubebuilder:validation:Enum=Running;Stopped;Present;Destroyed;Expunged
	// +optional
	State *string `json:"state,omitempty"`

	// StorageID defines the ID the storage where VM's volumes belong to.
	// +optional
	StorageID *string `json:"storageId,omitempty"`

	// Tags defines key/value pairs that should match the tags of the VMs.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`

	// TemplateID defines the ID of the template that was used to create the VMs.
	// +optional
	TempalteID *string `json:"templateId,omitempty"`

	// UserID defines the user ID that created the VM and is under the account that owns the VM.
	// +optional
	UserID *string `json:"userId,omitempty"`

	// VPCID defines the ID of the VPC the VM belongs to.
	// +optional
	VPCID *string `json:"vpcId,omitempty"`

	// ZoneID defines the availability zone the VM belongs to.
	// +optional
	ZoneID *string `json:"zoneId,omitempty"`
}

func (selector *CloudStackVMChaosSelector) Id() string {
	v, _ := json.Marshal(selector)
	return string(v)
}

func (obj *CloudStackVMChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{".": &obj.Spec.Selector}
}
