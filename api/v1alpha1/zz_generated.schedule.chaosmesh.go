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
	ScheduleTypeAwsChaos ScheduleTemplateType = "AwsChaos"
	ScheduleTypeDNSChaos ScheduleTemplateType = "DNSChaos"
	ScheduleTypeGcpChaos ScheduleTemplateType = "GcpChaos"
	ScheduleTypeHTTPChaos ScheduleTemplateType = "HTTPChaos"
	ScheduleTypeIOChaos ScheduleTemplateType = "IOChaos"
	ScheduleTypeJVMChaos ScheduleTemplateType = "JVMChaos"
	ScheduleTypeKernelChaos ScheduleTemplateType = "KernelChaos"
	ScheduleTypeNetworkChaos ScheduleTemplateType = "NetworkChaos"
	ScheduleTypePodChaos ScheduleTemplateType = "PodChaos"
	ScheduleTypeStressChaos ScheduleTemplateType = "StressChaos"
	ScheduleTypeTimeChaos ScheduleTemplateType = "TimeChaos"
	ScheduleTypeWorkflow ScheduleTemplateType = "Workflow"

)

var allScheduleTemplateType = []ScheduleTemplateType{
	ScheduleTypeAwsChaos,
	ScheduleTypeDNSChaos,
	ScheduleTypeGcpChaos,
	ScheduleTypeHTTPChaos,
	ScheduleTypeIOChaos,
	ScheduleTypeJVMChaos,
	ScheduleTypeKernelChaos,
	ScheduleTypeNetworkChaos,
	ScheduleTypePodChaos,
	ScheduleTypeStressChaos,
	ScheduleTypeTimeChaos,
	ScheduleTypeWorkflow,

}

func (it *ScheduleItem) SpawnNewObject(templateType ScheduleTemplateType) (runtime.Object, metav1.Object, error) {

	switch templateType {
	case ScheduleTypeAwsChaos:
		result := AwsChaos{}
		result.Spec = *it.AwsChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypeDNSChaos:
		result := DNSChaos{}
		result.Spec = *it.DNSChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypeGcpChaos:
		result := GcpChaos{}
		result.Spec = *it.GcpChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypeHTTPChaos:
		result := HTTPChaos{}
		result.Spec = *it.HTTPChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypeIOChaos:
		result := IOChaos{}
		result.Spec = *it.IOChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypeJVMChaos:
		result := JVMChaos{}
		result.Spec = *it.JVMChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypeKernelChaos:
		result := KernelChaos{}
		result.Spec = *it.KernelChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypeNetworkChaos:
		result := NetworkChaos{}
		result.Spec = *it.NetworkChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypePodChaos:
		result := PodChaos{}
		result.Spec = *it.PodChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypeStressChaos:
		result := StressChaos{}
		result.Spec = *it.StressChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypeTimeChaos:
		result := TimeChaos{}
		result.Spec = *it.TimeChaos
		return &result, result.GetObjectMeta(), nil
	case ScheduleTypeWorkflow:
		result := Workflow{}
		result.Spec = *it.Workflow
		return &result, result.GetObjectMeta(), nil

	default:
		return nil, nil, fmt.Errorf("unsupported template type %s", templateType)
	}

	return nil, &metav1.ObjectMeta{}, nil
}

