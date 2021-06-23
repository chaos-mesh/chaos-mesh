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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)


const (
	TypeAwsChaos TemplateType = "AwsChaos"
	TypeDNSChaos TemplateType = "DNSChaos"
	TypeGcpChaos TemplateType = "GcpChaos"
	TypeHTTPChaos TemplateType = "HTTPChaos"
	TypeIOChaos TemplateType = "IOChaos"
	TypeJVMChaos TemplateType = "JVMChaos"
	TypeKernelChaos TemplateType = "KernelChaos"
	TypeNetworkChaos TemplateType = "NetworkChaos"
	TypePodChaos TemplateType = "PodChaos"
	TypeStressChaos TemplateType = "StressChaos"
	TypeTimeChaos TemplateType = "TimeChaos"

)

var allChaosTemplateType = []TemplateType{
	TypeSchedule,
	TypeAwsChaos,
	TypeDNSChaos,
	TypeGcpChaos,
	TypeHTTPChaos,
	TypeIOChaos,
	TypeJVMChaos,
	TypeKernelChaos,
	TypeNetworkChaos,
	TypePodChaos,
	TypeStressChaos,
	TypeTimeChaos,

}

type EmbedChaos struct {
	// +optional
	AwsChaos *AwsChaosSpec `json:"awsChaos,omitempty"`
	// +optional
	DNSChaos *DNSChaosSpec `json:"dnsChaos,omitempty"`
	// +optional
	GcpChaos *GcpChaosSpec `json:"gcpChaos,omitempty"`
	// +optional
	HTTPChaos *HTTPChaosSpec `json:"httpChaos,omitempty"`
	// +optional
	IOChaos *IOChaosSpec `json:"ioChaos,omitempty"`
	// +optional
	JVMChaos *JVMChaosSpec `json:"jvmChaos,omitempty"`
	// +optional
	KernelChaos *KernelChaosSpec `json:"kernelChaos,omitempty"`
	// +optional
	NetworkChaos *NetworkChaosSpec `json:"networkChaos,omitempty"`
	// +optional
	PodChaos *PodChaosSpec `json:"podChaos,omitempty"`
	// +optional
	StressChaos *StressChaosSpec `json:"stressChaos,omitempty"`
	// +optional
	TimeChaos *TimeChaosSpec `json:"timeChaos,omitempty"`

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
	case TypeIOChaos:
		result := IOChaos{}
		result.Spec = *it.IOChaos
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

func (it *EmbedChaos) SpawnNewList(templateType TemplateType) (GenericChaosList, error) {

	switch templateType {
	case TypeAwsChaos:
		result := AwsChaosList{}
		return &result, nil
	case TypeDNSChaos:
		result := DNSChaosList{}
		return &result, nil
	case TypeGcpChaos:
		result := GcpChaosList{}
		return &result, nil
	case TypeHTTPChaos:
		result := HTTPChaosList{}
		return &result, nil
	case TypeIOChaos:
		result := IOChaosList{}
		return &result, nil
	case TypeJVMChaos:
		result := JVMChaosList{}
		return &result, nil
	case TypeKernelChaos:
		result := KernelChaosList{}
		return &result, nil
	case TypeNetworkChaos:
		result := NetworkChaosList{}
		return &result, nil
	case TypePodChaos:
		result := PodChaosList{}
		return &result, nil
	case TypeStressChaos:
		result := StressChaosList{}
		return &result, nil
	case TypeTimeChaos:
		result := TimeChaosList{}
		return &result, nil

	default:
		return nil, fmt.Errorf("unsupported template type %s", templateType)
	}

	return nil, nil
}

func (in *AwsChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
func (in *DNSChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
func (in *GcpChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
func (in *HTTPChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
func (in *IOChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
func (in *JVMChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
func (in *KernelChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
func (in *NetworkChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
func (in *PodChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
func (in *StressChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}
func (in *TimeChaosList) GetItems() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

