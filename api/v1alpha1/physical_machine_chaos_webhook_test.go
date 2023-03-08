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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("physicalmachinechaos_webhook", func() {
	Context("webhook.Defaultor of physicalmachinechaos", func() {
		It("Default", func() {
			physicalMachineChaos := &PhysicalMachineChaos{
				Spec: PhysicalMachineChaosSpec{
					Action: "stress-cpu",
					PhysicalMachineSelector: PhysicalMachineSelector{
						Address: []string{
							"123.123.123.123:123",
							"234.234.234.234:234",
						},
					},
					ExpInfo: ExpInfo{
						UID: "",
						StressCPU: &StressCPUSpec{
							Load:    10,
							Workers: 1,
						},
					},
				},
			}
			physicalMachineChaos.Default()
			Expect(physicalMachineChaos.Spec.UID).ToNot(Equal(""))
			Expect(physicalMachineChaos.Spec.Address).To(BeEquivalentTo([]string{
				"http://123.123.123.123:123",
				"http://234.234.234.234:234",
			}))
		})
	})
	Context("webhook.Validator of physicalmachinechaos", func() {
		It("Validate common", func() {
			testCases := []struct {
				chaos PhysicalMachineChaos
				err   string
			}{
				{
					PhysicalMachineChaos{
						Spec: PhysicalMachineChaosSpec{
							PhysicalMachineSelector: PhysicalMachineSelector{
								Address: []string{
									"123.123.123.123:123",
									"234.234.234.234:234",
								},
							},
							Action:  "stress-cpu",
							ExpInfo: ExpInfo{},
						},
					},
					"the configuration corresponding to action is required",
				},
			}

			for _, testCase := range testCases {
				err := testCase.chaos.ValidateCreate()
				Expect(strings.Contains(err.Error(), testCase.err)).To(BeTrue())
			}
		})

		It("Validate selector", func() {
			testCases := []struct {
				chaos PhysicalMachineChaos
				err   string
			}{
				{
					PhysicalMachineChaos{
						Spec: PhysicalMachineChaosSpec{
							Action: "stress-cpu",
							PhysicalMachineSelector: PhysicalMachineSelector{
								Address: []string{
									"123.123.123.123:123",
									"234.234.234.234:234",
								},
								Selector: PhysicalMachineSelectorSpec{
									PhysicalMachines: map[string][]string{
										"default": {"physical-machine1"},
									},
								},
							},
							ExpInfo: ExpInfo{
								UID: "",
								StressCPU: &StressCPUSpec{
									Load:    10,
									Workers: 1,
								},
							},
						},
					},
					"only one of address or selector could be specified",
				},
				{
					PhysicalMachineChaos{
						Spec: PhysicalMachineChaosSpec{
							Action: "stress-cpu",
							PhysicalMachineSelector: PhysicalMachineSelector{
								Selector: PhysicalMachineSelectorSpec{
									PhysicalMachines: map[string][]string{
										"default": {"physical-machine1"},
									},
								},
							},
							ExpInfo: ExpInfo{
								UID: "",
								StressCPU: &StressCPUSpec{
									Load:    10,
									Workers: 1,
								},
							},
						},
					},
					"",
				},
				{
					PhysicalMachineChaos{
						Spec: PhysicalMachineChaosSpec{
							Action: "stress-cpu",
							PhysicalMachineSelector: PhysicalMachineSelector{
								Address: []string{
									"123.123.123.123:123",
									"234.234.234.234:234",
								},
							},
							ExpInfo: ExpInfo{
								UID: "",
								StressCPU: &StressCPUSpec{
									Load:    10,
									Workers: 1,
								},
							},
						},
					},
					"",
				},
				{
					PhysicalMachineChaos{
						Spec: PhysicalMachineChaosSpec{
							Action:                  "stress-cpu",
							PhysicalMachineSelector: PhysicalMachineSelector{},
							ExpInfo: ExpInfo{
								UID: "",
								StressCPU: &StressCPUSpec{
									Load:    10,
									Workers: 1,
								},
							},
						},
					},
					"one of address or selector should be specified",
				},
			}

			for _, testCase := range testCases {
				err := testCase.chaos.ValidateCreate()
				if len(testCase.err) != 0 {
					Expect(err).To(HaveOccurred())
					Expect(strings.Contains(err.Error(), testCase.err)).To(BeTrue())
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			}
		})

		It("Validate config for specified action", func() {
			testCases := []struct {
				action  PhysicalMachineChaosAction
				expInfo ExpInfo
				err     string
			}{
				{
					PMStressCPUAction,
					ExpInfo{
						StressCPU: &StressCPUSpec{
							Load: 0,
						},
					},
					"load can't be 0",
				},
				{
					PMStressCPUAction,
					ExpInfo{
						StressCPU: &StressCPUSpec{
							Load:    1,
							Workers: 0,
						},
					},
					"workers can't be 0",
				},
				{
					PMStressCPUAction,
					ExpInfo{
						StressCPU: &StressCPUSpec{
							Load:    1,
							Workers: 1,
						},
					},
					"",
				},
				{
					PMStressMemAction,
					ExpInfo{
						StressMemory: &StressMemorySpec{
							Size: "",
						},
					},
					"size is required",
				},
				{
					PMStressMemAction,
					ExpInfo{
						StressMemory: &StressMemorySpec{
							Size: "123HB",
						},
					},
					"unknown unit",
				},
				{
					PMStressMemAction,
					ExpInfo{
						StressMemory: &StressMemorySpec{
							Size: "123MB",
						},
					},
					"",
				},
				{
					PMDiskReadPayloadAction,
					ExpInfo{
						DiskReadPayload: &DiskPayloadSpec{
							PayloadProcessNum: 0,
						},
					},
					"payload-process-num can't be 0",
				},
				{
					PMDiskReadPayloadAction,
					ExpInfo{
						DiskReadPayload: &DiskPayloadSpec{
							PayloadProcessNum: 1,
							DiskFileSpec: DiskFileSpec{
								Size: "",
							},
						},
					},
					"size is required",
				},
				{
					PMDiskReadPayloadAction,
					ExpInfo{
						DiskReadPayload: &DiskPayloadSpec{
							PayloadProcessNum: 1,
							DiskFileSpec: DiskFileSpec{
								Size: "100HB",
							},
						},
					},
					"unknown unit",
				},
				{
					PMDiskReadPayloadAction,
					ExpInfo{
						DiskReadPayload: &DiskPayloadSpec{
							PayloadProcessNum: 1,
							DiskFileSpec: DiskFileSpec{
								Size: "100MB",
							},
						},
					},
					"",
				},
				{
					PMDiskWritePayloadAction,
					ExpInfo{
						DiskWritePayload: &DiskPayloadSpec{
							PayloadProcessNum: 0,
						},
					},
					"payload-process-num can't be 0",
				},
				{
					PMDiskWritePayloadAction,
					ExpInfo{
						DiskWritePayload: &DiskPayloadSpec{
							PayloadProcessNum: 1,
							DiskFileSpec: DiskFileSpec{
								Size: "",
							},
						},
					},
					"size is required",
				},
				{
					PMDiskWritePayloadAction,
					ExpInfo{
						DiskWritePayload: &DiskPayloadSpec{
							PayloadProcessNum: 1,
							DiskFileSpec: DiskFileSpec{
								Size: "100HB",
							},
						},
					},
					"unknown unit",
				},
				{
					PMDiskWritePayloadAction,
					ExpInfo{
						DiskWritePayload: &DiskPayloadSpec{
							PayloadProcessNum: 1,
							DiskFileSpec: DiskFileSpec{
								Size: "100MB",
							},
						},
					},
					"",
				},
				{
					PMDiskFillAction,
					ExpInfo{
						DiskFill: &DiskFillSpec{
							DiskFileSpec: DiskFileSpec{
								Size: "",
							},
						},
					},
					"size is required",
				},
				{
					PMDiskFillAction,
					ExpInfo{
						DiskFill: &DiskFillSpec{
							DiskFileSpec: DiskFileSpec{
								Size: "100HB",
							},
						},
					},
					"unknown unit",
				},
				{
					PMDiskFillAction,
					ExpInfo{
						DiskFill: &DiskFillSpec{
							DiskFileSpec: DiskFileSpec{
								Size: "100MB",
							},
						},
					},
					"",
				},
				{
					PMNetworkCorruptAction,
					ExpInfo{
						NetworkCorrupt: &NetworkCorruptSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "-1",
							},
						},
					},
					"correlation -1 is invalid",
				},
				{
					PMNetworkCorruptAction,
					ExpInfo{
						NetworkCorrupt: &NetworkCorruptSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "",
							},
						},
					},
					"device is required",
				},
				{
					PMNetworkCorruptAction,
					ExpInfo{
						NetworkCorrupt: &NetworkCorruptSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "eth0",
								IPAddress:   "123.123.123.123",
							},
							Percent: "0",
						},
					},
					"percent is invalid",
				},
				{
					PMNetworkCorruptAction,
					ExpInfo{
						NetworkCorrupt: &NetworkCorruptSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "eth0",
								IPAddress:   "123.123.123.123",
							},
							Percent: "10",
						},
					},
					"",
				},
				{
					PMNetworkDuplicateAction,
					ExpInfo{
						NetworkDuplicate: &NetworkDuplicateSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "-1",
							},
						},
					},
					"correlation -1 is invalid",
				},
				{
					PMNetworkDuplicateAction,
					ExpInfo{
						NetworkDuplicate: &NetworkDuplicateSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "",
							},
						},
					},
					"device is required",
				},
				{
					PMNetworkDuplicateAction,
					ExpInfo{
						NetworkDuplicate: &NetworkDuplicateSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "eth0",
								IPAddress:   "123.123.123.123",
							},
							Percent: "0",
						},
					},
					"percent is invalid",
				},
				{
					PMNetworkDuplicateAction,
					ExpInfo{
						NetworkDuplicate: &NetworkDuplicateSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "eth0",
								IPAddress:   "123.123.123.123",
							},
							Percent: "10",
						},
					},
					"",
				},
				{
					PMNetworkLossAction,
					ExpInfo{
						NetworkLoss: &NetworkLossSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "",
							},
						},
					},
					"device is required",
				},
				{
					PMNetworkLossAction,
					ExpInfo{
						NetworkLoss: &NetworkLossSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "eth0",
								IPAddress:   "123.123.123.123",
							},
							Percent: "0",
						},
					},
					"percent is invalid",
				},
				{
					PMNetworkLossAction,
					ExpInfo{
						NetworkLoss: &NetworkLossSpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "eth0",
								IPAddress:   "123.123.123.123",
							},
							Percent: "10",
						},
					},
					"",
				},
				{
					PMNetworkDelayAction,
					ExpInfo{
						NetworkDelay: &NetworkDelaySpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "-1",
							},
						},
					},
					"correlation -1 is invalid",
				},
				{
					PMNetworkDelayAction,
					ExpInfo{
						NetworkDelay: &NetworkDelaySpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "",
							},
						},
					},
					"device is required",
				},
				{
					PMNetworkDelayAction,
					ExpInfo{
						NetworkDelay: &NetworkDelaySpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "eth0",
								Hostname:    "chaos-mesh.org",
							},
						},
					},
					"latency is invalid",
				},
				{
					PMNetworkDelayAction,
					ExpInfo{
						NetworkDelay: &NetworkDelaySpec{
							NetworkCommonSpec: NetworkCommonSpec{
								Correlation: "100",
								Device:      "eth0",
								Hostname:    "chaos-mesh.org",
							},
							Latency: "10ms",
						},
					},
					"",
				},
				{
					PMNetworkPartitionAction,
					ExpInfo{
						NetworkPartition: &NetworkPartitionSpec{

							Device: "",
						},
					},
					"device is required",
				},
				{
					PMNetworkPartitionAction,
					ExpInfo{
						NetworkPartition: &NetworkPartitionSpec{
							Device: "eth0",
						},
					},
					"one of ip-address and hostname is required",
				},
				{
					PMNetworkPartitionAction,
					ExpInfo{
						NetworkPartition: &NetworkPartitionSpec{
							Device:    "eth0",
							Hostname:  "chaos-mesh.org",
							Direction: "nil",
						},
					},
					"direction should be one of 'to' and 'from'",
				},
				{
					PMNetworkPartitionAction,
					ExpInfo{
						NetworkPartition: &NetworkPartitionSpec{
							Device:         "eth0",
							Hostname:       "chaos-mesh.org",
							Direction:      "to",
							AcceptTCPFlags: "SYN,ACK SYN,ACK",
							IPProtocol:     "udp",
						},
					},
					"protocol should be 'tcp' when set accept-tcp-flags",
				},
				{
					PMNetworkPartitionAction,
					ExpInfo{
						NetworkPartition: &NetworkPartitionSpec{
							Device:         "eth0",
							Hostname:       "chaos-mesh.org",
							Direction:      "to",
							AcceptTCPFlags: "SYN,ACK SYN,ACK",
							IPProtocol:     "tcp",
						},
					},
					"",
				},
				{
					PMNetworkDNSAction,
					ExpInfo{
						NetworkDNS: &NetworkDNSSpec{
							DNSDomainName: "chaos-mesh.org",
							DNSIp:         "",
						},
					},
					"DNS host chaos-mesh.org must match a DNS ip",
				},
				{
					PMNetworkDNSAction,
					ExpInfo{
						NetworkDNS: &NetworkDNSSpec{
							DNSDomainName: "",
							DNSIp:         "123.123.123.123",
						},
					},
					"DNS host  must match a DNS ip 123.123.123.123",
				},
				{
					PMNetworkDNSAction,
					ExpInfo{
						NetworkDNS: &NetworkDNSSpec{
							DNSDomainName: "chaos-mesh.org",
							DNSIp:         "123.123.123.123",
						},
					},
					"",
				},
				{
					PMProcessAction,
					ExpInfo{
						Process: &ProcessSpec{
							Process: "",
						},
					},
					"process is required",
				},
				{
					PMProcessAction,
					ExpInfo{
						Process: &ProcessSpec{
							Process: "123",
							Signal:  0,
						},
					},
					"signal is required",
				},
				{
					PMProcessAction,
					ExpInfo{
						Process: &ProcessSpec{
							Process: "123",
							Signal:  19,
						},
					},
					"",
				},
				{
					PMJVMExceptionAction,
					ExpInfo{
						JVMException: &JVMExceptionSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 0,
							},
						},
					},
					"pid is required",
				},
				{
					PMJVMExceptionAction,
					ExpInfo{
						JVMException: &JVMExceptionSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class: "",
							},
						},
					},
					"class is required",
				},
				{
					PMJVMExceptionAction,
					ExpInfo{
						JVMException: &JVMExceptionSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class:  "Main",
								Method: "",
							},
						},
					},
					"method is required",
				},
				{
					PMJVMExceptionAction,
					ExpInfo{
						JVMException: &JVMExceptionSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class:  "Main",
								Method: "test",
							},
							ThrowException: "",
						},
					},
					"exception is required",
				},
				{
					PMJVMExceptionAction,
					ExpInfo{
						JVMException: &JVMExceptionSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class:  "Main",
								Method: "test",
							},
							ThrowException: "java.io.IOException(\"BOOM\")",
						},
					},
					"",
				},
				{
					PMJVMGCAction,
					ExpInfo{
						JVMGC: &JVMGCSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 0,
							},
						},
					},
					"pid is required",
				},
				{
					PMJVMGCAction,
					ExpInfo{
						JVMGC: &JVMGCSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 10,
							},
						},
					},
					"",
				},
				{
					PMJVMLatencyAction,
					ExpInfo{
						JVMLatency: &JVMLatencySpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 0,
							},
						},
					},
					"pid is required",
				},
				{
					PMJVMLatencyAction,
					ExpInfo{
						JVMLatency: &JVMLatencySpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class: "",
							},
						},
					},
					"class is required",
				},
				{
					PMJVMLatencyAction,
					ExpInfo{
						JVMLatency: &JVMLatencySpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class:  "Main",
								Method: "",
							},
						},
					},
					"method is required",
				},
				{
					PMJVMLatencyAction,
					ExpInfo{
						JVMLatency: &JVMLatencySpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class:  "Main",
								Method: "test",
							},
							LatencyDuration: 0,
						},
					},
					"latency is required",
				},
				{
					PMJVMLatencyAction,
					ExpInfo{
						JVMLatency: &JVMLatencySpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class:  "Main",
								Method: "test",
							},
							LatencyDuration: 1000,
						},
					},
					"",
				},
				{
					PMJVMReturnAction,
					ExpInfo{
						JVMReturn: &JVMReturnSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 0,
							},
						},
					},
					"pid is required",
				},
				{
					PMJVMReturnAction,
					ExpInfo{
						JVMReturn: &JVMReturnSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class: "",
							},
						},
					},
					"class is required",
				},
				{
					PMJVMReturnAction,
					ExpInfo{
						JVMReturn: &JVMReturnSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class:  "Main",
								Method: "",
							},
						},
					},
					"method is required",
				},
				{
					PMJVMReturnAction,
					ExpInfo{
						JVMReturn: &JVMReturnSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class:  "Main",
								Method: "test",
							},
							ReturnValue: "",
						},
					},
					"value is required",
				},
				{
					PMJVMReturnAction,
					ExpInfo{
						JVMReturn: &JVMReturnSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 123,
							},
							JVMClassMethodSpec: JVMClassMethodSpec{
								Class:  "Main",
								Method: "test",
							},
							ReturnValue: "123",
						},
					},
					"",
				},
				{
					PMJVMStressAction,
					ExpInfo{
						JVMStress: &JVMStressSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 0,
							},
						},
					},
					"pid is required",
				},
				{
					PMJVMStressAction,
					ExpInfo{
						JVMStress: &JVMStressSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 10,
							},

							CPUCount:   0,
							MemoryType: "",
						},
					},
					"one of cpu-count and mem-type is required",
				},
				{
					PMJVMStressAction,
					ExpInfo{
						JVMStress: &JVMStressSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 10,
							},
							CPUCount:   1,
							MemoryType: "heap",
						},
					},
					"inject stress on both CPU and memory is not support",
				},
				{
					PMJVMStressAction,
					ExpInfo{
						JVMStress: &JVMStressSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 10,
							},
							CPUCount:   0,
							MemoryType: "heap",
						},
					},
					"",
				},
				{
					PMJVMRuleDataAction,
					ExpInfo{
						JVMRuleData: &JVMRuleDataSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 0,
							},
						},
					},
					"pid is required",
				},
				{
					PMJVMRuleDataAction,
					ExpInfo{
						JVMRuleData: &JVMRuleDataSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 10,
							},
							RuleData: "",
						},
					},
					"rule-data is required",
				},
				{
					PMJVMRuleDataAction,
					ExpInfo{
						JVMRuleData: &JVMRuleDataSpec{
							JVMCommonSpec: JVMCommonSpec{
								Pid: 10,
							},
							RuleData: "RULE modify return value\nCLASS Main\nMETHOD getnum\nAT ENTRY\nIF true\nDO\n    return 9999\nENDRULE",
						},
					},
					"",
				},
				{
					PMClockAction,
					ExpInfo{
						Clock: &ClockSpec{
							Pid: 0,
						},
					},
					"pid is required",
				},
				{
					PMClockAction,
					ExpInfo{
						Clock: &ClockSpec{
							Pid:        123,
							TimeOffset: "",
						},
					},
					"time-offset is required",
				},
				{
					PMClockAction,
					ExpInfo{
						Clock: &ClockSpec{
							Pid:        123,
							TimeOffset: "10m",
						},
					},
					"",
				},
			}

			for _, testCase := range testCases {
				chaos := PhysicalMachineChaos{
					Spec: PhysicalMachineChaosSpec{
						PhysicalMachineSelector: PhysicalMachineSelector{
							Address: []string{
								"123.123.123.123:123",
								"234.234.234.234:234",
							},
						},
						Action:  testCase.action,
						ExpInfo: testCase.expInfo,
					},
				}
				err := chaos.ValidateCreate()
				if len(testCase.err) != 0 {
					Expect(err).To(HaveOccurred())
					Expect(strings.Contains(err.Error(), testCase.err)).To(BeTrue())
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			}
		})
	})
	Context("webhook.Validator of bandwidth physicalmachinechaos", func() {
		It("Validate", func() {
			testCases := []struct {
				chaos PhysicalMachineChaos
				err   string
			}{
				{
					PhysicalMachineChaos{
						Spec: PhysicalMachineChaosSpec{
							PhysicalMachineSelector: PhysicalMachineSelector{
								Address: []string{
									"123.123.123.123:123",
									"234.234.234.234:234",
								},
							},
							Action: "network",
							ExpInfo: ExpInfo{
								NetworkBandwidth: &NetworkBandwidthSpec{
									Rate:   "",
									Limit:  0,
									Buffer: 0,
								},
							},
						},
					},
					"rate is required",
				},
			}

			for _, testCase := range testCases {
				err := testCase.chaos.ValidateCreate()
				Expect(strings.Contains(err.Error(), testCase.err)).To(BeTrue())
			}
		})
	})
})
