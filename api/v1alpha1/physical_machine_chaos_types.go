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
// +chaos-mesh:experiment

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
	// +kubebuilder:validation:Enum=stress-cpu;stress-mem;disk-read-payload;disk-write-payload;disk-fill;network-corrupt;network-duplicate;network-loss;network-delay;network-partition;network-dns;process;jvm-exception;jvm-gc;jvm-latency;jvm-return;jvm-stress;jvm-rule-data;clock
	Action PhysicalMachineChaosAction `json:"action"`

	PhysicalMachineSelector `json:",inline"`

	// ExpInfo string `json:"expInfo"`
	ExpInfo `json:",inline"`

	// Duration represents the duration of the chaos action
	// +optional
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
	NetworkPartition *NetworkPartitionSpec `json:"network-partition,omitempty"`

	// +optional
	NetworkDNS *NetworkDNSSpec `json:"network-dns,omitempty"`

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

	// +optional
	Clock *ClockSpec `json:"clock,omitempty"`
}

type StressCPUSpec struct {
	// specifies P percent loading per CPU worker. 0 is effectively a sleep (no load) and 100 is full loading.
	Load    int `json:"load,omitempty"`
	// specifies N workers to apply the stressor.
	Workers int `json:"workers,omitempty"`
}

type StressMemorySpec struct {
	// specifies N bytes consumed per vm worker, default is the total available memory.
	// One can specify the size as % of total available memory or in units of B, KB/KiB, MB/MiB, GB/GiB, TB/TiB..
	Size string `json:"size,omitempty"`
}

type DiskFileSpec struct {
	// specifies how many units of data will write into the file path. support unit: c=1, w=2, b=512, kB=1000,
	// K=1024, MB=1000*1000, M=1024*1024, GB=1000*1000*1000, G=1024*1024*1024 BYTES. example : 1M | 512kB
	Size string `json:"size,omitempty"`
	// specifies the location to fill data in. if path not provided,
	// payload will write into a temp file, temp file will be deleted after writing
	Path string `json:"path,omitempty"`
}

type DiskPayloadSpec struct {
	DiskFileSpec      `json:",inline"`

	// specifies the number of process work on writing, default 1, only 1-255 is valid value
	PayloadProcessNum uint8 `json:"payload_process_num,omitempty"`
}

type DiskFillSpec struct {
	DiskFileSpec    `json:",inline"`

	// fill disk by fallocate
	FillByFallocate bool `json:"fill_by_fallocate,omitempty"`
}

type NetworkCommonSpec struct {
	// correlation is percentage (10 is 10%)
	Correlation string `json:"correlation,omitempty"`
	// the network interface to impact
	Device      string `json:"device,omitempty"`
	// only impact egress traffic from these source ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010.
	// it can only be used in conjunction with -p tcp or -p udp
	SourcePort  string `json:"source-port,omitempty"`
	// only impact egress traffic to these destination ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010.
	// it can only be used in conjunction with -p tcp or -p udp
	EgressPort  string `json:"egress-port,omitempty"`
	// only impact egress traffic to these IP addresses
	IPAddress   string `json:"ip-address,omitempty"`
	// only impact traffic using this IP protocol, supported: tcp, udp, icmp, all
	IPProtocol  string `json:"ip-protocol,omitempty"`
	// only impact traffic to these hostnames
	Hostname    string `json:"hostname,omitempty"`
}

type NetworkCorruptSpec struct {
	NetworkCommonSpec `json:",inline"`

	// percentage of packets to corrupt (10 is 10%)
	Percent string `json:"percent,omitempty"`
}

type NetworkDuplicateSpec struct {
	NetworkCommonSpec `json:",inline"`

	// percentage of packets to duplicate (10 is 10%)
	Percent string `json:"percent,omitempty"`
}

type NetworkLossSpec struct {
	NetworkCommonSpec `json:",inline"`

	// percentage of packets to loss (10 is 10%)
	Percent string `json:"percent,omitempty"`
}

type NetworkDelaySpec struct {
	NetworkCommonSpec `json:",inline"`

	// jitter time, time units: ns, us (or µs), ms, s, m, h.
	Jitter  string `json:"jitter,omitempty"`
	// delay egress time, time units: ns, us (or µs), ms, s, m, h.
	Latency string `json:"latency,omitempty"`
}

type NetworkPartitionSpec struct {
	// the network interface to impact
	Device string `json:"device,omitempty"`
	// only impact traffic to these hostnames
	Hostname  string `json:"hostname,omitempty"`
	// only impact egress traffic to these IP addresses
	IPAddress string `json:"ip-address,omitempty"`
	// specifies the partition direction, values can be 'from', 'to'.
	// 'from' means packets coming from the 'IPAddress' or 'Hostname' and going to your server,
	// 'to' means packets originating from your server and going to the 'IPAddress' or 'Hostname'.
	Direction string `json:"direction,omitempty"`
	// only impact egress traffic to these IP addresses
	IPProtocol string `json:"ip-protocol,omitempty"`
	// only the packet which match the tcp flag can be accepted, others will be dropped.
	// only set when the IPProtocol is tcp, used for partition.
	AcceptTCPFlags string `json:"accept-tcp-flags,omitempty"`
}

type NetworkDNSSpec struct {
	// update the DNS server in /etc/resolv.conf with this value
	DNSServer     string `json:"dns-server,omitempty"`
	// map specified host to this IP address
	DNSIp         string `json:"dns-ip,omitempty"`
	// map this host to specified IP
	DNSDomainName string `json:"dns-domain-name,omitempty"`
}

type ProcessSpec struct {
	// the process name or the process ID
	Process string `json:"process,omitempty"`
	// the signal number to send
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

	// the memory type need to locate, only set it when action is stress, the value can be 'stack' or 'heap'
	MemoryType int `json:"mem-type,omitempty"`
}

type JVMRuleDataSpec struct {
	JVMCommonSpec `json:",inline"`

	// RuleData used to save the rule file's data, will use it when recover
	RuleData string `json:"rule-data,omitempty"`
}

type ClockSpec struct {
	// the pid of target program.
	Pid int `json:"pid,omitempty"`
	// specifies the length of time offset.
	TimeOffset string `json:"time-offset,omitempty"`
	// the identifier of the particular clock on which to act.
	// More clock description in linux kernel can be found in man page of clock_getres, clock_gettime, clock_settime.
	// Muti clock ids should be split with ","
	ClockIdsSlice string `json:"clock-ids-slice,omitempty"`
}
