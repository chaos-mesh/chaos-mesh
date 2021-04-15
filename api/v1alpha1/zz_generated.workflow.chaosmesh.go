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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)


const (
	TypeAwsChaos TemplateType = "AwsChaos"
	TypeDNSChaos TemplateType = "DNSChaos"
	TypeGcpChaos TemplateType = "GcpChaos"
	TypeHTTPChaos TemplateType = "HTTPChaos"
	TypeIoChaos TemplateType = "IoChaos"
	TypeJVMChaos TemplateType = "JVMChaos"
	TypeKernelChaos TemplateType = "KernelChaos"
	TypeNetworkChaos TemplateType = "NetworkChaos"
	TypePodChaos TemplateType = "PodChaos"
	TypeStressChaos TemplateType = "StressChaos"
	TypeTimeChaos TemplateType = "TimeChaos"

)

var allChaosTemplateType = []TemplateType{
	TypeAwsChaos,
	TypeDNSChaos,
	TypeGcpChaos,
	TypeHTTPChaos,
	TypeIoChaos,
	TypeJVMChaos,
	TypeKernelChaos,
	TypeNetworkChaos,
	TypePodChaos,
	TypeStressChaos,
	TypeTimeChaos,

}

type EmbedChaos struct {
	// +optional
	AwsChaos *AwsChaosSpec `json:"aws_chaos,omitempty" mapstructure:"aws_chaos"`
	// +optional
	DNSChaos *DNSChaosSpec `json:"dns_chaos,omitempty" mapstructure:"dns_chaos"`
	// +optional
	GcpChaos *GcpChaosSpec `json:"gcp_chaos,omitempty" mapstructure:"gcp_chaos"`
	// +optional
	HTTPChaos *HTTPChaosSpec `json:"http_chaos,omitempty" mapstructure:"http_chaos"`
	// +optional
	IoChaos *IoChaosSpec `json:"io_chaos,omitempty" mapstructure:"io_chaos"`
	// +optional
	JVMChaos *JVMChaosSpec `json:"jvm_chaos,omitempty" mapstructure:"jvm_chaos"`
	// +optional
	KernelChaos *KernelChaosSpec `json:"kernel_chaos,omitempty" mapstructure:"kernel_chaos"`
	// +optional
	NetworkChaos *NetworkChaosSpec `json:"network_chaos,omitempty" mapstructure:"network_chaos"`
	// +optional
	PodChaos *PodChaosSpec `json:"pod_chaos,omitempty" mapstructure:"pod_chaos"`
	// +optional
	StressChaos *StressChaosSpec `json:"stress_chaos,omitempty" mapstructure:"stress_chaos"`
	// +optional
	TimeChaos *TimeChaosSpec `json:"time_chaos,omitempty" mapstructure:"time_chaos"`
}

func (it *EmbedChaos) SpawnNewObject(templateType TemplateType) (runtime.Object, metav1.Object, error) {

	switch templateType {
	case TypeAwsChaos:
		result := AwsChaos{}
		result.Spec = *it.AwsChaos
		return &result, result.GetObjectMeta(), nil
	case TypeDNSChaos:
		result := DNSChaos{}
		result.Spec = *it.DNSChaos
		return &result, result.GetObjectMeta(), nil
	case TypeGcpChaos:
		result := GcpChaos{}
		result.Spec = *it.GcpChaos
		return &result, result.GetObjectMeta(), nil
	case TypeHTTPChaos:
		result := HTTPChaos{}
		result.Spec = *it.HTTPChaos
		return &result, result.GetObjectMeta(), nil
	case TypeIoChaos:
		result := IoChaos{}
		result.Spec = *it.IoChaos
		return &result, result.GetObjectMeta(), nil
	case TypeJVMChaos:
		result := JVMChaos{}
		result.Spec = *it.JVMChaos
		return &result, result.GetObjectMeta(), nil
	case TypeKernelChaos:
		result := KernelChaos{}
		result.Spec = *it.KernelChaos
		return &result, result.GetObjectMeta(), nil
	case TypeNetworkChaos:
		result := NetworkChaos{}
		result.Spec = *it.NetworkChaos
		return &result, result.GetObjectMeta(), nil
	case TypePodChaos:
		result := PodChaos{}
		result.Spec = *it.PodChaos
		return &result, result.GetObjectMeta(), nil
	case TypeStressChaos:
		result := StressChaos{}
		result.Spec = *it.StressChaos
		return &result, result.GetObjectMeta(), nil
	case TypeTimeChaos:
		result := TimeChaos{}
		result.Spec = *it.TimeChaos
		return &result, result.GetObjectMeta(), nil

	default:
		return nil, nil, fmt.Errorf("unsupported template type %s", templateType)
	}

	return nil, &metav1.ObjectMeta{}, nil
}
