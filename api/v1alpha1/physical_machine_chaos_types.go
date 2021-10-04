// Copyright 2021 Chaos Mesh Authors.
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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PhysicalMachineChaosAction represents the chaos action about physical machine.
type PhysicalMachineChaosAction string

// +kubebuilder:object:root=true
// +chaos-mesh:base

// PhysicalMachineChaos is the Schema for the physical machine chaos API
type PhysicalMachineChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a physical machine chaos experiment
	Spec PhysicalMachineChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment
	Status PhysicalMachineChaosStatus `json:"status"`
}

// PhysicalMachineChaosSpec defines the desired state of PhysicalMachineChaos
type PhysicalMachineChaosSpec struct {
	// +kubebuilder:validation:Enum=stress-cpu;stress-mem;disk-read-payload;disk-write-payload;disk-fill;network-corrupt;network-duplicate;network-loss;network-delay;process;jvm-exception;jvm-gc;jvm-latency;jvm-return;jvm-stress;jvm-rule-data
	Action PhysicalMachineChaosAction `json:"action"`

	PhysicalMachineSelector `json:",inline"`

	// ExpInfo string `json:"expInfo"`
	ExpInfo `json:",inline"`

	// Duration represents the duration of the chaos action
	// +optional
	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty" webhook:"Duration"`
}

// PhysicalMachineChaosStatus defines the observed state of PhysicalMachineChaos
type PhysicalMachineChaosStatus struct {
	ChaosStatus `json:",inline"`
}

func (obj *PhysicalMachineChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.PhysicalMachineSelector,
	}
}

type PhysicalMachineSelector struct {
	Address []string `json:"address"`
}

func (selector *PhysicalMachineSelector) Id() string {
	return strings.Join(selector.Address, ",")
}

type ExpInfo struct {
	// the experiment ID
	// +optional
	UID string `json:"uid,omitempty"`

	// the subAction, generate automatically
	// +optional
	Action string `json:"action,omitempty"`

	// +optional
	StressCPU *StressCPUSpec `json:"stress-cpu,omitempty"`

	// +optional
	StressMemory *StressMemorySpec `json:"stress-mem,omitempty"`

	// +optional
	DiskReadPayload *DiskPayloadSpec `json:"disk-read-payload,omitempty"`

	// +optional
	DiskWritePayload *DiskPayloadSpec `json:"disk-write-payload,omitempty"`

	// +optional
	DiskFill *DiskFillSpec `json:"disk-fill,omitempty"`

	// +optional
	NetworkCorrupt *NetworkCorruptSpec `json:"network-corrupt,omitempty"`

	// +optional
	NetworkDuplicate *NetworkDuplicateSpec `json:"network-duplicate,omitempty"`

	// +optional
	NetworkLoss *NetworkLossSpec `json:"network-loss,omitempty"`

	// +optional
	NetworkDelay *NetworkDelaySpec `json:"network-delay,omitempty"`

	// +optional
	Process *ProcessSpec `json:"process,omitempty"`

	// +optional
	JVMException *JVMExceptionSpec `json:"jvm-exception,omitempty"`

	// +optional
	JVMGC *JVMGCSpec `json:"jvm-gc,omitempty"`

	// +optional
	JVMLatency *JVMLatencySpec `json:"jvm-latency,omitempty"`

	// +optional
	JVMReturn *JVMReturnSpec `json:"jvm-return,omitempty"`

	// +optional
	JVMStress *JVMStressSpec `json:"jvm-stress,omitempty"`

	// +optional
	JVMRuleData *JVMRuleDataSpec `json:"jvm-rule-data,omitempty"`
}

type StressCPUSpec struct {
	Load    int `json:"load,omitempty"`
	Workers int `json:"workers,omitempty"`
}

type StressMemorySpec struct {
	Size string `json:"size,omitempty"`
}

type DiskFileSpec struct {
	Size string `json:"size,omitempty"`
	Path string `json:"path,omitempty"`
}

type DiskPayloadSpec struct {
	DiskFileSpec      `json:",inline"`
	PayloadProcessNum uint8 `json:"payload_process_num,omitempty"`
}

type DiskFillSpec struct {
	DiskFileSpec    `json:",inline"`
	FillByFallocate bool `json:"fill_by_fallocate,omitempty"`
}

type NetworkCommonSpec struct {
	Correlation string `json:"correlation,omitempty"`
	Device      string `json:"device,omitempty"`
	SourcePort  string `json:"source-port,omitempty"`
	EgressPort  string `json:"egress-port,omitempty"`
	IPAddress   string `json:"ip-address,omitempty"`
	IPProtocol  string `json:"ip-protocol,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
}

type NetworkCorruptSpec struct {
	NetworkCommonSpec `json:",inline"`

	Percent string `json:"percent,omitempty"`
}

type NetworkDuplicateSpec struct {
	NetworkCommonSpec `json:",inline"`

	Percent string `json:"percent,omitempty"`
}

type NetworkLossSpec struct {
	NetworkCommonSpec `json:",inline"`

	Percent string `json:"percent,omitempty"`
}

type NetworkDelaySpec struct {
	NetworkCommonSpec `json:",inline"`

	Jitter  string `json:"jitter,omitempty"`
	Latency string `json:"latency,omitempty"`
}

type ProcessSpec struct {
	Process string `json:"process,omitempty"`
	Signal  int    `json:"signal,omitempty"`
}

type JVMCommonSpec struct {
	// the port of agent server
	Port int `json:"port,omitempty"`

	// the pid of Java process which need to attach
	Pid int `json:"pid,omitempty"`
}

type JVMClassMethodSpec struct {
	// Java class
	Class string `json:"class,omitempty"`

	// the method in Java class
	Method string `json:"method,omitempty"`
}

type JVMExceptionSpec struct {
	JVMCommonSpec      `json:",inline"`
	JVMClassMethodSpec `json:",inline"`

	// the exception which needs to throw dor action `exception`
	ThrowException string `json:"exception,omitempty"`
}

type JVMGCSpec struct {
	JVMCommonSpec `json:",inline"`
}

type JVMLatencySpec struct {
	JVMCommonSpec      `json:",inline"`
	JVMClassMethodSpec `json:",inline"`

	// the latency duration for action 'latency', unit ms
	LatencyDuration int `json:"latency,omitempty"`
}

type JVMReturnSpec struct {
	JVMCommonSpec      `json:",inline"`
	JVMClassMethodSpec `json:",inline"`

	// the return value for action 'return'
	ReturnValue string `json:"value,omitempty"`
}

type JVMStressSpec struct {
	JVMCommonSpec `json:",inline"`

	// the CPU core number need to use, only set it when action is stress
	CPUCount int `json:"cpu-count,omitempty"`

	// the memory size need to locate, only set it when action is stress
	MemorySize int `json:"mem-size,omitempty"`
}

type JVMRuleDataSpec struct {
	JVMCommonSpec `json:",inline"`

	// RuleData used to save the rule file's data, will use it when recover
	RuleData string `json:"rule-data,omitempty"`
}
