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
)


const (
	ScheduleTypeAWSChaos ScheduleTemplateType = "AWSChaos"
	ScheduleTypeDNSChaos ScheduleTemplateType = "DNSChaos"
	ScheduleTypeGCPChaos ScheduleTemplateType = "GCPChaos"
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
	ScheduleTypeAWSChaos,
	ScheduleTypeDNSChaos,
	ScheduleTypeGCPChaos,
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

func (it *ScheduleItem) SpawnNewObject(templateType ScheduleTemplateType) (GenericChaos, error) {

	switch templateType {
	case ScheduleTypeAWSChaos:
		result := AWSChaos{}
		result.Spec = *it.AWSChaos
		return &result, nil
	case ScheduleTypeDNSChaos:
		result := DNSChaos{}
		result.Spec = *it.DNSChaos
		return &result, nil
	case ScheduleTypeGCPChaos:
		result := GCPChaos{}
		result.Spec = *it.GCPChaos
		return &result, nil
	case ScheduleTypeHTTPChaos:
		result := HTTPChaos{}
		result.Spec = *it.HTTPChaos
		return &result, nil
	case ScheduleTypeIOChaos:
		result := IOChaos{}
		result.Spec = *it.IOChaos
		return &result, nil
	case ScheduleTypeJVMChaos:
		result := JVMChaos{}
		result.Spec = *it.JVMChaos
		return &result, nil
	case ScheduleTypeKernelChaos:
		result := KernelChaos{}
		result.Spec = *it.KernelChaos
		return &result, nil
	case ScheduleTypeNetworkChaos:
		result := NetworkChaos{}
		result.Spec = *it.NetworkChaos
		return &result, nil
	case ScheduleTypePodChaos:
		result := PodChaos{}
		result.Spec = *it.PodChaos
		return &result, nil
	case ScheduleTypeStressChaos:
		result := StressChaos{}
		result.Spec = *it.StressChaos
		return &result, nil
	case ScheduleTypeTimeChaos:
		result := TimeChaos{}
		result.Spec = *it.TimeChaos
		return &result, nil
	case ScheduleTypeWorkflow:
		result := Workflow{}
		result.Spec = *it.Workflow
		return &result, nil

	default:
		return nil, fmt.Errorf("unsupported template type %s", templateType)
	}

	return nil, nil
}

