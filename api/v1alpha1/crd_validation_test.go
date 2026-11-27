// Copyright 2026 Chaos Mesh Authors.
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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// These specs run the CEL rules from the CRD schemas against the envtest API server,
// without the validating webhook.
var _ = Describe("CRD validation rules", func() {
	const (
		containerKillMessage = "the name of container is required on container-kill action"
		blockDelayMessage    = "delay should be set on delay action"
		externalTargetsMsg   = "external targets cannot be used with `from` and `both` direction in netem action yet"
		emptyTargetMessage   = "`from` and `both` direction cannot be used when targets is empty in netem action"
		ebsVolumeMessage     = "the ID of EBS volume is required on detach-volume action"
		awsDeviceMessage     = "the name of device is required on detach-volume action"
		azureDiskMessage     = "the name of data disk should not be empty on disk-detach action"
		azureLUNMessage      = "the LUN of data disk should not be empty on disk-detach action"
		gcpDeviceMessage     = "at least one device name is required on disk-loss action"
		ioErrnoMessage       = "action fault: errno 0 is not supported"
		ioPercentMessage     = "percent field should be in 0-100"
		httpPortMessage      = "port must be greater than 0"
		httpMethodMessage    = "method is not a supported HTTP method"
		dnsStaticMessage     = "at least one domain=IP mapping is required for the static action"
		jvmStressOneMessage  = "must set one of cpu-count and mem-type when action is 'stress'"
		jvmStressBothMessage = "inject stress on both CPU and memory is not support now"
		jvmMemTypeMessage    = "value should be 'stack' or 'heap'"
		jvmClassMessage      = "class not provided"
		jvmMethodMessage     = "method not provided"
		jvmExceptionMessage  = "exception not provided"
		jvmValueMessage      = "value not provided"
		jvmLatencyMessage    = "latency not provided"
		jvmRuleDataMessage   = "rule data not provide"
		jvmMySQLMessage      = "MySQL connector version not provided"
		jvmMySQLVerMessage   = "MySQL connector version only support 5.X.X(set to 5) and 8.X.X(set to 8)"
		jvmMySQLFaultMessage = "must set one of exception or latency"
		missingStressors     = "missing stressors"
	)

	meta := func(name string) metav1.ObjectMeta {
		return metav1.ObjectMeta{Name: name, Namespace: "default"}
	}

	containerKill := func(names ...string) *PodChaosSpec {
		return &PodChaosSpec{
			Action: ContainerKillAction,
			ContainerSelector: ContainerSelector{
				PodSelector:    PodSelector{Mode: OneMode},
				ContainerNames: names,
			},
		}
	}

	blockChaos := func(delay *BlockDelaySpec) *BlockChaosSpec {
		return &BlockChaosSpec{
			Action: BlockDelay,
			Delay:  delay,
			ContainerNodeVolumePathSelector: ContainerNodeVolumePathSelector{
				ContainerSelector: ContainerSelector{PodSelector: PodSelector{Mode: OneMode}},
				VolumeName:        "data",
			},
		}
	}

	networkChaos := func(action NetworkChaosAction, direction Direction, target *PodSelector, externalTargets ...string) *NetworkChaosSpec {
		return &NetworkChaosSpec{
			PodSelector:     PodSelector{Mode: OneMode},
			Action:          action,
			TcParameter:     TcParameter{Delay: &DelaySpec{Latency: "10ms"}},
			Direction:       direction,
			Target:          target,
			ExternalTargets: externalTargets,
		}
	}

	target := &PodSelector{Mode: AllMode}

	str := func(s string) *string { return &s }

	awsChaos := func(action AWSChaosAction, volumeID, deviceName *string) *AWSChaosSpec {
		return &AWSChaosSpec{
			Action: action,
			AWSSelector: AWSSelector{
				AWSRegion:   "us-east-1",
				Ec2Instance: "i-0123456789",
				EbsVolume:   volumeID,
				DeviceName:  deviceName,
			},
		}
	}

	azureChaos := func(action AzureChaosAction, diskName *string, lun *int) *AzureChaosSpec {
		return &AzureChaosSpec{
			Action: action,
			AzureSelector: AzureSelector{
				SubscriptionID:    "sub",
				ResourceGroupName: "rg",
				VMName:            "vm",
				DiskName:          diskName,
				LUN:               lun,
			},
		}
	}

	gcpChaos := func(action GCPChaosAction, deviceNames ...string) *GCPChaosSpec {
		return &GCPChaosSpec{
			Action:      action,
			GCPSelector: GCPSelector{Project: "p", Zone: "z", Instance: "i", DeviceNames: deviceNames},
		}
	}

	ioChaos := func(action IOChaosType, errno uint32) *IOChaosSpec {
		return &IOChaosSpec{
			ContainerSelector: ContainerSelector{PodSelector: PodSelector{Mode: OneMode}},
			Action:            action,
			Delay:             "10ms",
			Errno:             errno,
			VolumePath:        "/data",
		}
	}

	httpChaos := func(target PodHttpChaosTarget, port int32, method *string) *HTTPChaosSpec {
		return &HTTPChaosSpec{
			PodSelector: PodSelector{Mode: OneMode},
			Target:      target,
			Port:        port,
			Method:      method,
		}
	}

	dnsChaos := func(action DNSChaosAction, mappings ...string) *DNSChaosSpec {
		return &DNSChaosSpec{
			Action:              action,
			ContainerSelector:   ContainerSelector{PodSelector: PodSelector{Mode: OneMode}},
			DomainIPMappingList: mappings,
		}
	}

	jvmChaos := func(action JVMChaosAction, param JVMParameter) *JVMChaosSpec {
		return &JVMChaosSpec{
			ContainerSelector: ContainerSelector{PodSelector: PodSelector{Mode: OneMode}},
			Action:            action,
			JVMParameter:      param,
		}
	}

	classMethod := JVMClassMethodSpec{Class: "Main", Method: "run"}

	stressChaos := func(stressors *Stressors, stressng string) *StressChaosSpec {
		return &StressChaosSpec{
			ContainerSelector: ContainerSelector{PodSelector: PodSelector{Mode: OneMode}},
			Stressors:         stressors,
			StressngStressors: stressng,
		}
	}

	cpuStressors := &Stressors{CPUStressor: &CPUStressor{Stressor: Stressor{Workers: 1}}}

	pmChaos := func(action PhysicalMachineChaosAction, info ExpInfo) *PhysicalMachineChaosSpec {
		return &PhysicalMachineChaosSpec{
			Action: action,
			PhysicalMachineSelector: PhysicalMachineSelector{
				Address: []string{"http://192.0.2.1:31767"},
				Mode:    OneMode,
			},
			ExpInfo: info,
		}
	}

	expectRejected := func(obj client.Object, message string) {
		err := k8sClient.Create(context.TODO(), obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(message))
	}

	expectAccepted := func(obj client.Object) {
		Expect(k8sClient.Create(context.TODO(), obj)).To(Succeed())
		Expect(k8sClient.Delete(context.TODO(), obj)).To(Succeed())
	}

	Context("PodChaos", func() {
		It("rejects container-kill without container names", func() {
			expectRejected(&PodChaos{ObjectMeta: meta("cel-container-kill-empty"), Spec: *containerKill()}, containerKillMessage)
		})

		It("accepts container-kill with container names", func() {
			expectAccepted(&PodChaos{ObjectMeta: meta("cel-container-kill"), Spec: *containerKill("app")})
		})

		It("accepts pod-kill without container names", func() {
			spec := containerKill()
			spec.Action = PodKillAction
			expectAccepted(&PodChaos{ObjectMeta: meta("cel-pod-kill"), Spec: *spec})
		})
	})

	Context("BlockChaos", func() {
		It("rejects the delay action without a delay spec", func() {
			expectRejected(&BlockChaos{ObjectMeta: meta("cel-block-no-delay"), Spec: *blockChaos(nil)}, blockDelayMessage)
		})

		It("accepts the delay action with a delay spec", func() {
			expectAccepted(&BlockChaos{ObjectMeta: meta("cel-block-delay"), Spec: *blockChaos(&BlockDelaySpec{Latency: "10ms"})})
		})
	})

	Context("NetworkChaos", func() {
		It("rejects the from direction without a target", func() {
			expectRejected(&NetworkChaos{ObjectMeta: meta("cel-net-from-no-target"), Spec: *networkChaos(DelayAction, From, nil)}, emptyTargetMessage)
		})

		It("rejects the both direction with external targets", func() {
			expectRejected(&NetworkChaos{ObjectMeta: meta("cel-net-both-external"), Spec: *networkChaos(DelayAction, Both, target, "8.8.8.8")}, externalTargetsMsg)
		})

		It("rejects the both direction with an empty external targets list", func() {
			// The typed client drops an empty list, so send it as unstructured like kubectl would.
			obj := &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": SchemeGroupVersion.String(),
				"kind":       "NetworkChaos",
				"metadata":   map[string]any{"name": "cel-net-both-empty-external", "namespace": "default"},
				"spec": map[string]any{
					"mode":            string(OneMode),
					"selector":        map[string]any{},
					"action":          string(DelayAction),
					"delay":           map[string]any{"latency": "10ms"},
					"direction":       string(Both),
					"target":          map[string]any{"mode": string(AllMode), "selector": map[string]any{}},
					"externalTargets": []any{},
				},
			}}
			expectRejected(obj, externalTargetsMsg)
		})

		It("accepts the both direction with a target", func() {
			expectAccepted(&NetworkChaos{ObjectMeta: meta("cel-net-both-target"), Spec: *networkChaos(DelayAction, Both, target)})
		})

		It("accepts the to direction with external targets and no target", func() {
			expectAccepted(&NetworkChaos{ObjectMeta: meta("cel-net-to-external"), Spec: *networkChaos(DelayAction, To, nil, "8.8.8.8")})
		})

		It("accepts an absent direction with external targets and no target", func() {
			expectAccepted(&NetworkChaos{ObjectMeta: meta("cel-net-default-direction"), Spec: *networkChaos(DelayAction, "", nil, "8.8.8.8")})
		})

		It("accepts partition with the both direction and no target", func() {
			expectAccepted(&NetworkChaos{ObjectMeta: meta("cel-net-partition-both"), Spec: *networkChaos(PartitionAction, Both, nil)})
		})
	})

	Context("AWSChaos", func() {
		It("rejects detach-volume without a volume ID", func() {
			expectRejected(&AWSChaos{ObjectMeta: meta("cel-aws-no-volume"), Spec: *awsChaos(DetachVolume, nil, str("/dev/sdf"))}, ebsVolumeMessage)
		})

		It("rejects detach-volume without a device name", func() {
			expectRejected(&AWSChaos{ObjectMeta: meta("cel-aws-no-device"), Spec: *awsChaos(DetachVolume, str("vol-1"), nil)}, awsDeviceMessage)
		})

		It("accepts detach-volume with a volume ID and a device name", func() {
			expectAccepted(&AWSChaos{ObjectMeta: meta("cel-aws-detach"), Spec: *awsChaos(DetachVolume, str("vol-1"), str("/dev/sdf"))})
		})

		It("accepts ec2-stop without a volume", func() {
			expectAccepted(&AWSChaos{ObjectMeta: meta("cel-aws-stop"), Spec: *awsChaos(Ec2Stop, nil, nil)})
		})
	})

	Context("AzureChaos", func() {
		lun := 0

		It("rejects disk-detach without a disk name", func() {
			expectRejected(&AzureChaos{ObjectMeta: meta("cel-azure-no-disk"), Spec: *azureChaos(AzureDiskDetach, nil, &lun)}, azureDiskMessage)
		})

		It("rejects disk-detach without a LUN", func() {
			expectRejected(&AzureChaos{ObjectMeta: meta("cel-azure-no-lun"), Spec: *azureChaos(AzureDiskDetach, str("data"), nil)}, azureLUNMessage)
		})

		It("accepts disk-detach with a disk name and LUN 0", func() {
			expectAccepted(&AzureChaos{ObjectMeta: meta("cel-azure-detach"), Spec: *azureChaos(AzureDiskDetach, str("data"), &lun)})
		})

		It("accepts vm-stop without a disk", func() {
			expectAccepted(&AzureChaos{ObjectMeta: meta("cel-azure-stop"), Spec: *azureChaos(AzureVmStop, nil, nil)})
		})
	})

	Context("GCPChaos", func() {
		It("rejects disk-loss without device names", func() {
			expectRejected(&GCPChaos{ObjectMeta: meta("cel-gcp-no-devices"), Spec: *gcpChaos(DiskLoss)}, gcpDeviceMessage)
		})

		It("accepts disk-loss with device names", func() {
			expectAccepted(&GCPChaos{ObjectMeta: meta("cel-gcp-disk-loss"), Spec: *gcpChaos(DiskLoss, "disk-1")})
		})

		It("accepts node-stop without device names", func() {
			expectAccepted(&GCPChaos{ObjectMeta: meta("cel-gcp-stop"), Spec: *gcpChaos(NodeStop)})
		})
	})

	Context("IOChaos", func() {
		It("rejects fault without an errno", func() {
			expectRejected(&IOChaos{ObjectMeta: meta("cel-io-no-errno"), Spec: *ioChaos(IoFaults, 0)}, ioErrnoMessage)
		})

		It("accepts fault with an errno", func() {
			expectAccepted(&IOChaos{ObjectMeta: meta("cel-io-fault"), Spec: *ioChaos(IoFaults, 5)})
		})

		It("rejects a percent above 100", func() {
			spec := ioChaos(IoLatency, 0)
			spec.Percent = 101
			expectRejected(&IOChaos{ObjectMeta: meta("cel-io-percent-high"), Spec: *spec}, ioPercentMessage)
		})

		It("rejects a negative percent", func() {
			spec := ioChaos(IoLatency, 0)
			spec.Percent = -1
			expectRejected(&IOChaos{ObjectMeta: meta("cel-io-percent-negative"), Spec: *spec}, ioPercentMessage)
		})

		It("accepts latency without an errno at 100 percent", func() {
			spec := ioChaos(IoLatency, 0)
			spec.Percent = 100
			expectAccepted(&IOChaos{ObjectMeta: meta("cel-io-percent"), Spec: *spec})
		})
	})

	Context("HTTPChaos", func() {
		It("rejects a missing port", func() {
			expectRejected(&HTTPChaos{ObjectMeta: meta("cel-http-no-port"), Spec: *httpChaos(PodHttpRequest, 0, nil)}, httpPortMessage)
		})

		It("rejects a negative port", func() {
			expectRejected(&HTTPChaos{ObjectMeta: meta("cel-http-negative-port"), Spec: *httpChaos(PodHttpRequest, -1, nil)}, httpPortMessage)
		})

		It("accepts a positive port", func() {
			expectAccepted(&HTTPChaos{ObjectMeta: meta("cel-http-port"), Spec: *httpChaos(PodHttpRequest, 80, nil)})
		})

		It("rejects an unknown method on the request target", func() {
			expectRejected(&HTTPChaos{ObjectMeta: meta("cel-http-bad-method"), Spec: *httpChaos(PodHttpRequest, 80, str("FOO"))}, httpMethodMessage)
		})

		It("accepts a known method on the request target", func() {
			expectAccepted(&HTTPChaos{ObjectMeta: meta("cel-http-method"), Spec: *httpChaos(PodHttpRequest, 80, str("GET"))})
		})

		It("accepts any method on the response target", func() {
			expectAccepted(&HTTPChaos{ObjectMeta: meta("cel-http-response-method"), Spec: *httpChaos(PodHttpResponse, 80, str("FOO"))})
		})
	})

	Context("DNSChaos", func() {
		It("rejects static without mappings", func() {
			expectRejected(&DNSChaos{ObjectMeta: meta("cel-dns-static-empty"), Spec: *dnsChaos(StaticAction)}, dnsStaticMessage)
		})

		It("accepts static with a mapping", func() {
			expectAccepted(&DNSChaos{ObjectMeta: meta("cel-dns-static"), Spec: *dnsChaos(StaticAction, "example.com=192.0.2.1")})
		})

		It("accepts error without mappings", func() {
			expectAccepted(&DNSChaos{ObjectMeta: meta("cel-dns-error"), Spec: *dnsChaos(ErrorAction)})
		})
	})

	Context("JVMChaos", func() {
		jvm := func(name string, action JVMChaosAction, param JVMParameter) *JVMChaos {
			return &JVMChaos{ObjectMeta: meta(name), Spec: *jvmChaos(action, param)}
		}

		It("rejects stress without a CPU count or memory type", func() {
			expectRejected(jvm("cel-jvm-stress-empty", JVMStressAction, JVMParameter{}), jvmStressOneMessage)
		})

		It("accepts stress with a CPU count", func() {
			expectAccepted(jvm("cel-jvm-stress-cpu", JVMStressAction, JVMParameter{JVMStressCfgSpec: JVMStressCfgSpec{CPUCount: 1}}))
		})

		It("rejects stress with both a CPU count and a memory type", func() {
			expectRejected(jvm("cel-jvm-stress-both", JVMStressAction, JVMParameter{JVMStressCfgSpec: JVMStressCfgSpec{CPUCount: 1, MemoryType: "heap"}}), jvmStressBothMessage)
		})

		It("accepts stress with a heap memory type", func() {
			expectAccepted(jvm("cel-jvm-stress-heap", JVMStressAction, JVMParameter{JVMStressCfgSpec: JVMStressCfgSpec{MemoryType: "heap"}}))
		})

		It("rejects stress with an unknown memory type", func() {
			expectRejected(jvm("cel-jvm-stress-memtype", JVMStressAction, JVMParameter{JVMStressCfgSpec: JVMStressCfgSpec{MemoryType: "disk"}}), jvmMemTypeMessage)
		})

		It("rejects exception without a class", func() {
			expectRejected(jvm("cel-jvm-no-class", JVMExceptionAction, JVMParameter{JVMClassMethodSpec: JVMClassMethodSpec{Method: "run"}, ThrowException: "java.io.IOException"}), jvmClassMessage)
		})

		It("rejects latency without a method", func() {
			expectRejected(jvm("cel-jvm-no-method", JVMLatencyAction, JVMParameter{JVMClassMethodSpec: JVMClassMethodSpec{Class: "Main"}, LatencyDuration: 100}), jvmMethodMessage)
		})

		It("rejects exception without an exception", func() {
			expectRejected(jvm("cel-jvm-no-exception", JVMExceptionAction, JVMParameter{JVMClassMethodSpec: classMethod}), jvmExceptionMessage)
		})

		It("accepts exception with a class, method and exception", func() {
			expectAccepted(jvm("cel-jvm-exception", JVMExceptionAction, JVMParameter{JVMClassMethodSpec: classMethod, ThrowException: "java.io.IOException"}))
		})

		It("rejects return without a value", func() {
			expectRejected(jvm("cel-jvm-no-value", JVMReturnAction, JVMParameter{JVMClassMethodSpec: classMethod}), jvmValueMessage)
		})

		It("accepts return with a value", func() {
			expectAccepted(jvm("cel-jvm-return", JVMReturnAction, JVMParameter{JVMClassMethodSpec: classMethod, ReturnValue: "0"}))
		})

		It("rejects latency without a latency", func() {
			expectRejected(jvm("cel-jvm-no-latency", JVMLatencyAction, JVMParameter{JVMClassMethodSpec: classMethod}), jvmLatencyMessage)
		})

		It("accepts latency with a latency", func() {
			expectAccepted(jvm("cel-jvm-latency", JVMLatencyAction, JVMParameter{JVMClassMethodSpec: classMethod, LatencyDuration: 100}))
		})

		It("rejects ruleData without rule data", func() {
			expectRejected(jvm("cel-jvm-no-rule", JVMRuleDataAction, JVMParameter{}), jvmRuleDataMessage)
		})

		It("accepts ruleData with rule data", func() {
			expectAccepted(jvm("cel-jvm-rule", JVMRuleDataAction, JVMParameter{RuleData: "RULE test"}))
		})

		It("rejects mysql without a connector version", func() {
			expectRejected(jvm("cel-jvm-mysql-no-version", JVMMySQLAction, JVMParameter{LatencyDuration: 100}), jvmMySQLMessage)
		})

		It("rejects mysql with an unsupported connector version", func() {
			expectRejected(jvm("cel-jvm-mysql-version", JVMMySQLAction, JVMParameter{JVMMySQLSpec: JVMMySQLSpec{MySQLConnectorVersion: "6"}, LatencyDuration: 100}), jvmMySQLVerMessage)
		})

		It("rejects mysql without an exception or latency", func() {
			expectRejected(jvm("cel-jvm-mysql-no-fault", JVMMySQLAction, JVMParameter{JVMMySQLSpec: JVMMySQLSpec{MySQLConnectorVersion: "8"}}), jvmMySQLFaultMessage)
		})

		It("accepts mysql with a connector version and an exception", func() {
			expectAccepted(jvm("cel-jvm-mysql", JVMMySQLAction, JVMParameter{JVMMySQLSpec: JVMMySQLSpec{MySQLConnectorVersion: "5"}, ThrowException: "boom"}))
		})

		It("accepts gc without parameters", func() {
			expectAccepted(jvm("cel-jvm-gc", JVMGCAction, JVMParameter{}))
		})
	})

	Context("StressChaos", func() {
		stress := func(name string, stressors *Stressors, stressng string) *StressChaos {
			return &StressChaos{ObjectMeta: meta(name), Spec: *stressChaos(stressors, stressng)}
		}

		It("rejects a spec without stressors", func() {
			expectRejected(stress("cel-stress-none", nil, ""), missingStressors)
		})

		It("accepts stress-ng stressors only", func() {
			expectAccepted(stress("cel-stress-ng", nil, "--cpu 1"))
		})

		It("rejects empty stressors", func() {
			expectRejected(stress("cel-stress-empty", &Stressors{}, ""), missingStressors)
		})

		It("accepts a CPU stressor", func() {
			expectAccepted(stress("cel-stress-cpu", cpuStressors, ""))
		})
	})

	Context("PhysicalMachineChaos", func() {
		pm := func(name string, action PhysicalMachineChaosAction, info ExpInfo) *PhysicalMachineChaos {
			return &PhysicalMachineChaos{ObjectMeta: meta(name), Spec: *pmChaos(action, info)}
		}

		It("ignores an incomplete configuration that belongs to another action", func() {
			expectAccepted(pm("cel-pm-other-config", PMStressMemAction, ExpInfo{StressMemory: &StressMemorySpec{Size: "1GB"}, StressCPU: &StressCPUSpec{}}))
		})

		DescribeTable("rejects actions missing required fields",
			func(name string, action PhysicalMachineChaosAction, info ExpInfo, message string) {
				expectRejected(pm(name, action, info), message)
			},
			Entry("stress-cpu load", "cel-pm-cpu-load", PMStressCPUAction,
				ExpInfo{StressCPU: &StressCPUSpec{Workers: 1}}, "load can't be 0"),
			Entry("stress-cpu workers", "cel-pm-cpu-workers", PMStressCPUAction,
				ExpInfo{StressCPU: &StressCPUSpec{Load: 10}}, "workers can't be 0"),
			Entry("stress-mem size", "cel-pm-mem-size", PMStressMemAction,
				ExpInfo{StressMemory: &StressMemorySpec{}}, "size is required"),
			Entry("disk-read-payload process num", "cel-pm-read-num", PMDiskReadPayloadAction,
				ExpInfo{DiskReadPayload: &DiskPayloadSpec{DiskFileSpec: DiskFileSpec{Size: "1GB"}}}, "payload-process-num can't be 0"),
			Entry("disk-read-payload size", "cel-pm-read-size", PMDiskReadPayloadAction,
				ExpInfo{DiskReadPayload: &DiskPayloadSpec{PayloadProcessNum: 1}}, "size is required"),
			Entry("disk-write-payload process num", "cel-pm-write-num", PMDiskWritePayloadAction,
				ExpInfo{DiskWritePayload: &DiskPayloadSpec{DiskFileSpec: DiskFileSpec{Size: "1GB"}}}, "payload-process-num can't be 0"),
			Entry("disk-write-payload size", "cel-pm-write-size", PMDiskWritePayloadAction,
				ExpInfo{DiskWritePayload: &DiskPayloadSpec{PayloadProcessNum: 1}}, "size is required"),
			Entry("disk-fill size", "cel-pm-fill-size", PMDiskFillAction,
				ExpInfo{DiskFill: &DiskFillSpec{}}, "size is required"),
			Entry("process name", "cel-pm-process-name", PMProcessAction,
				ExpInfo{Process: &ProcessSpec{Signal: 9}}, "process is required"),
			Entry("process signal", "cel-pm-process-signal", PMProcessAction,
				ExpInfo{Process: &ProcessSpec{Process: "nginx"}}, "signal is required"),
			Entry("user_defined attack command", "cel-pm-attack", PMUserDefinedAction,
				ExpInfo{UserDefined: &UserDefinedSpec{RecoverCmd: "true"}}, "attack command not provided"),
			Entry("user_defined recover command", "cel-pm-recover", PMUserDefinedAction,
				ExpInfo{UserDefined: &UserDefinedSpec{AttackCmd: "true"}}, "recover command not provided"),
		)

		DescribeTable("accepts actions with their required fields",
			func(name string, action PhysicalMachineChaosAction, info ExpInfo) {
				expectAccepted(pm(name, action, info))
			},
			Entry("stress-cpu", "cel-pm-cpu-ok", PMStressCPUAction,
				ExpInfo{StressCPU: &StressCPUSpec{Load: 10, Workers: 1}}),
			Entry("stress-mem", "cel-pm-mem-ok", PMStressMemAction,
				ExpInfo{StressMemory: &StressMemorySpec{Size: "1GB"}}),
			Entry("disk-read-payload", "cel-pm-read-ok", PMDiskReadPayloadAction,
				ExpInfo{DiskReadPayload: &DiskPayloadSpec{DiskFileSpec: DiskFileSpec{Size: "1GB"}, PayloadProcessNum: 1}}),
			Entry("disk-write-payload", "cel-pm-write-ok", PMDiskWritePayloadAction,
				ExpInfo{DiskWritePayload: &DiskPayloadSpec{DiskFileSpec: DiskFileSpec{Size: "1GB"}, PayloadProcessNum: 1}}),
			Entry("disk-fill", "cel-pm-fill-ok", PMDiskFillAction,
				ExpInfo{DiskFill: &DiskFillSpec{DiskFileSpec: DiskFileSpec{Size: "1GB"}}}),
			Entry("process", "cel-pm-process-ok", PMProcessAction,
				ExpInfo{Process: &ProcessSpec{Process: "nginx", Signal: 9}}),
			Entry("user_defined", "cel-pm-user-defined-ok", PMUserDefinedAction,
				ExpInfo{UserDefined: &UserDefinedSpec{AttackCmd: "true", RecoverCmd: "true"}}),
		)
	})

	Context("Schedule", func() {
		schedule := func(name string, embed EmbedChaos, scheduleType ScheduleTemplateType) *Schedule {
			return &Schedule{
				ObjectMeta: meta(name),
				Spec: ScheduleSpec{
					Schedule:          "* * * * 1",
					ScheduleItem:      ScheduleItem{EmbedChaos: embed},
					Type:              scheduleType,
					ConcurrencyPolicy: ForbidConcurrent,
				},
			}
		}

		It("rejects an embedded container-kill without container names", func() {
			expectRejected(schedule("cel-schedule-container-kill", EmbedChaos{PodChaos: containerKill()}, ScheduleTypePodChaos), containerKillMessage)
		})

		It("rejects an embedded delay action without a delay spec", func() {
			expectRejected(schedule("cel-schedule-block", EmbedChaos{BlockChaos: blockChaos(nil)}, ScheduleTypeBlockChaos), blockDelayMessage)
		})

		It("rejects an embedded from direction without a target", func() {
			expectRejected(schedule("cel-schedule-net", EmbedChaos{NetworkChaos: networkChaos(DelayAction, From, nil)}, ScheduleTypeNetworkChaos), emptyTargetMessage)
		})

		It("rejects an embedded detach-volume without a volume ID", func() {
			expectRejected(schedule("cel-schedule-aws", EmbedChaos{AWSChaos: awsChaos(DetachVolume, nil, str("/dev/sdf"))}, ScheduleTypeAWSChaos), ebsVolumeMessage)
		})

		It("rejects an embedded fault without an errno", func() {
			expectRejected(schedule("cel-schedule-io", EmbedChaos{IOChaos: ioChaos(IoFaults, 0)}, ScheduleTypeIOChaos), ioErrnoMessage)
		})

		It("rejects an embedded static DNSChaos without mappings", func() {
			expectRejected(schedule("cel-schedule-dns", EmbedChaos{DNSChaos: dnsChaos(StaticAction)}, ScheduleTypeDNSChaos), dnsStaticMessage)
		})

		It("accepts an embedded static DNSChaos with a mapping", func() {
			expectAccepted(schedule("cel-schedule-dns-ok", EmbedChaos{DNSChaos: dnsChaos(StaticAction, "example.com=192.0.2.1")}, ScheduleTypeDNSChaos))
		})

		It("accepts an embedded partition with the both direction and no target", func() {
			expectAccepted(schedule("cel-schedule-net-ok", EmbedChaos{NetworkChaos: networkChaos(PartitionAction, Both, nil)}, ScheduleTypeNetworkChaos))
		})

		It("rejects an embedded JVM stress without a CPU count or memory type", func() {
			expectRejected(schedule("cel-schedule-jvm", EmbedChaos{JVMChaos: jvmChaos(JVMStressAction, JVMParameter{})}, ScheduleTypeJVMChaos), jvmStressOneMessage)
		})

		It("rejects an embedded StressChaos without stressors", func() {
			expectRejected(schedule("cel-schedule-stress", EmbedChaos{StressChaos: stressChaos(nil, "")}, ScheduleTypeStressChaos), missingStressors)
		})

		It("accepts an embedded StressChaos with a CPU stressor", func() {
			expectAccepted(schedule("cel-schedule-stress-ok", EmbedChaos{StressChaos: stressChaos(cpuStressors, "")}, ScheduleTypeStressChaos))
		})

		It("rejects an embedded process action without a signal", func() {
			expectRejected(schedule("cel-schedule-pm", EmbedChaos{PhysicalMachineChaos: pmChaos(PMProcessAction, ExpInfo{Process: &ProcessSpec{Process: "nginx"}})}, ScheduleTypePhysicalMachineChaos), "signal is required")
		})

		It("accepts a valid embedded container-kill", func() {
			expectAccepted(schedule("cel-schedule-ok", EmbedChaos{PodChaos: containerKill("app")}, ScheduleTypePodChaos))
		})
	})

	Context("Workflow", func() {
		workflow := func(name string, embed EmbedChaos, templateType TemplateType) *Workflow {
			return &Workflow{
				ObjectMeta: meta(name),
				Spec: WorkflowSpec{
					Entry: "entry",
					Templates: []Template{{
						Name:       "entry",
						Type:       templateType,
						EmbedChaos: &embed,
					}},
				},
			}
		}

		It("rejects an embedded container-kill without container names", func() {
			expectRejected(workflow("cel-workflow-container-kill", EmbedChaos{PodChaos: containerKill()}, TypePodChaos), containerKillMessage)
		})

		It("rejects an embedded delay action without a delay spec", func() {
			expectRejected(workflow("cel-workflow-block", EmbedChaos{BlockChaos: blockChaos(nil)}, TypeBlockChaos), blockDelayMessage)
		})

		It("rejects an embedded both direction with external targets", func() {
			expectRejected(workflow("cel-workflow-net", EmbedChaos{NetworkChaos: networkChaos(DelayAction, Both, target, "8.8.8.8")}, TypeNetworkChaos), externalTargetsMsg)
		})

		It("rejects an embedded disk-detach without a LUN", func() {
			expectRejected(workflow("cel-workflow-azure", EmbedChaos{AzureChaos: azureChaos(AzureDiskDetach, str("data"), nil)}, TypeAzureChaos), azureLUNMessage)
		})

		It("rejects an embedded disk-loss without device names", func() {
			expectRejected(workflow("cel-workflow-gcp", EmbedChaos{GCPChaos: gcpChaos(DiskLoss)}, TypeGCPChaos), gcpDeviceMessage)
		})

		It("rejects an embedded HTTPChaos without a port", func() {
			expectRejected(workflow("cel-workflow-http", EmbedChaos{HTTPChaos: httpChaos(PodHttpRequest, 0, nil)}, TypeHTTPChaos), httpPortMessage)
		})

		It("accepts an embedded HTTPChaos with a port", func() {
			expectAccepted(workflow("cel-workflow-http-ok", EmbedChaos{HTTPChaos: httpChaos(PodHttpRequest, 80, nil)}, TypeHTTPChaos))
		})

		It("rejects an embedded JVM return without a value", func() {
			expectRejected(workflow("cel-workflow-jvm", EmbedChaos{JVMChaos: jvmChaos(JVMReturnAction, JVMParameter{JVMClassMethodSpec: classMethod})}, TypeJVMChaos), jvmValueMessage)
		})

		It("accepts an embedded JVM return with a value", func() {
			expectAccepted(workflow("cel-workflow-jvm-ok", EmbedChaos{JVMChaos: jvmChaos(JVMReturnAction, JVMParameter{JVMClassMethodSpec: classMethod, ReturnValue: "0"})}, TypeJVMChaos))
		})

		It("rejects embedded empty stressors", func() {
			expectRejected(workflow("cel-workflow-stress", EmbedChaos{StressChaos: stressChaos(&Stressors{}, "")}, TypeStressChaos), missingStressors)
		})

		It("rejects an embedded stress-cpu without workers", func() {
			expectRejected(workflow("cel-workflow-pm", EmbedChaos{PhysicalMachineChaos: pmChaos(PMStressCPUAction, ExpInfo{StressCPU: &StressCPUSpec{Load: 10}})}, TypePhysicalMachineChaos), "workers can't be 0")
		})

		It("accepts an embedded stress-cpu with load and workers", func() {
			expectAccepted(workflow("cel-workflow-pm-ok", EmbedChaos{PhysicalMachineChaos: pmChaos(PMStressCPUAction, ExpInfo{StressCPU: &StressCPUSpec{Load: 10, Workers: 1}})}, TypePhysicalMachineChaos))
		})

		It("accepts a valid embedded container-kill", func() {
			expectAccepted(workflow("cel-workflow-ok", EmbedChaos{PodChaos: containerKill("app")}, TypePodChaos))
		})
	})
})
