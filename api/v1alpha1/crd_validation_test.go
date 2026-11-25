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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// These specs run the CEL rules from the CRD schemas against the envtest API server,
// without the validating webhook.
var _ = Describe("CRD validation rules", func() {
	const (
		containerKillMessage = "the name of container is required on container-kill action"
		blockDelayMessage    = "delay should be set on delay action"
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

		It("accepts a valid embedded container-kill", func() {
			expectAccepted(workflow("cel-workflow-ok", EmbedChaos{PodChaos: containerKill("app")}, TypePodChaos))
		})
	})
})
