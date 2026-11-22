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

		It("accepts latency without an errno", func() {
			expectAccepted(&IOChaos{ObjectMeta: meta("cel-io-latency"), Spec: *ioChaos(IoLatency, 0)})
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

		It("accepts a percent of 100", func() {
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

		It("accepts a valid embedded container-kill", func() {
			expectAccepted(workflow("cel-workflow-ok", EmbedChaos{PodChaos: containerKill("app")}, TypePodChaos))
		})
	})
})
