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

package controllers

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// integration tests
var _ = Describe("Workflow", func() {
	var ns string
	BeforeEach(func() {
		ctx := context.TODO()
		newNs := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "chaos-mesh-",
			},
			Spec: corev1.NamespaceSpec{},
		}
		Expect(kubeClient.Create(ctx, &newNs)).To(Succeed())
		ns = newNs.Name
		By(fmt.Sprintf("create new namespace %s", ns))
	})

	AfterEach(func() {
		ctx := context.TODO()
		nsToDelete := corev1.Namespace{}
		Expect(kubeClient.Get(ctx, types.NamespacedName{Name: ns}, &nsToDelete)).To(Succeed())
		Expect(kubeClient.Delete(ctx, &nsToDelete)).To(Succeed())
		By(fmt.Sprintf("cleanup namespace %s", ns))
	})

	Context("with deadline", func() {
		Context("on suspend", func() {
			It("should do nothing except waiting", func() {
				ctx := context.TODO()
				now := time.Now()
				duration := 5 * time.Second
				toleratedJitter := 3 * time.Second

				By("create simple suspend node")
				startTime := metav1.NewTime(now)
				deadline := metav1.NewTime(now.Add(duration))
				node := v1alpha1.WorkflowNode{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:    ns,
						GenerateName: "suspend-node-",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						WorkflowName: "",
						Type:         v1alpha1.TypeSuspend,
						StartTime:    &startTime,
						Deadline:     &deadline,
					},
				}
				Expect(kubeClient.Create(ctx, &node)).To(Succeed())

				// TODO: no other side effects

				By("assert this node is finished")
				Eventually(func() bool {
					updatedNode := v1alpha1.WorkflowNode{}
					Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: node.Name}, &updatedNode)).To(Succeed())
					return ConditionEqualsTo(updatedNode.Status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionTrue)
				}, duration+toleratedJitter, time.Second).Should(BeTrue())

				Eventually(func() bool {
					updatedNode := v1alpha1.WorkflowNode{}
					Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: node.Name}, &updatedNode)).To(Succeed())
					return WorkflowNodeFinished(updatedNode.Status)
				}, toleratedJitter, time.Second).Should(BeTrue())
			})
		})

		Context("on chaos node with chaos", func() {
			It("should delete chaos as soon as deadline exceed", func() {
				ctx := context.TODO()
				now := time.Now()
				duration := 5 * time.Second
				toleratedJitter := 3 * time.Second

				By("create simple chaos node with pod chaos")
				startTime := metav1.NewTime(now)
				deadline := metav1.NewTime(now.Add(duration))
				node := v1alpha1.WorkflowNode{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:    ns,
						GenerateName: "pod-chaos-",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						WorkflowName: "",
						Type:         v1alpha1.TypePodChaos,
						StartTime:    &startTime,
						Deadline:     &deadline,
						EmbedChaos: &v1alpha1.EmbedChaos{
							PodChaos: &v1alpha1.PodChaosSpec{
								ContainerSelector: v1alpha1.ContainerSelector{
									PodSelector: v1alpha1.PodSelector{
										Selector: v1alpha1.PodSelectorSpec{
											Namespaces: []string{ns},
											LabelSelectors: map[string]string{
												"app": "not-actually-exist",
											},
										},
										Mode: v1alpha1.AllPodMode,
									},
									ContainerNames: nil,
								},
								Action: v1alpha1.PodKillAction,
							},
						},
					},
				}
				Expect(kubeClient.Create(ctx, &node)).To(Succeed())

				By("assert that pod chaos CR is created")
				Eventually(func() bool {
					updatedNode := v1alpha1.WorkflowNode{}
					Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: node.Name}, &updatedNode)).To(Succeed())
					if !ConditionEqualsTo(updatedNode.Status, v1alpha1.ConditionChaosInjected, corev1.ConditionTrue) {
						return false
					}
					chaos := v1alpha1.PodChaos{}
					err := kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: updatedNode.Status.ChaosResource.Name}, &chaos)
					return err == nil
				}, toleratedJitter, time.Second).Should(BeTrue())

				By("assert that pod chaos should be purged")
				Eventually(func() bool {
					podChaosList := v1alpha1.PodChaosList{}
					Expect(kubeClient.List(ctx, &podChaosList, &client.ListOptions{Namespace: ns})).To(Succeed())
					return len(podChaosList.Items) == 0
				}, duration+toleratedJitter, time.Second).Should(BeTrue())
			})
		})

		Context("on chaos node with schedule", func() {
			It("should delete schedule as soon as deadline exceed", func() {
				ctx := context.TODO()
				now := time.Now()
				duration := 5 * time.Second
				toleratedJitter := 3 * time.Second

				By("create simple chaos node with pod chaos")
				startTime := metav1.NewTime(now)
				deadline := metav1.NewTime(now.Add(duration))
				node := v1alpha1.WorkflowNode{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:    ns,
						GenerateName: "pod-chaos-",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						WorkflowName: "",
						Type:         v1alpha1.TypeSchedule,
						StartTime:    &startTime,
						Deadline:     &deadline,
						Schedule: &v1alpha1.ScheduleSpec{
							Schedule:                "@every 1s",
							StartingDeadlineSeconds: nil,
							ConcurrencyPolicy:       v1alpha1.AllowConcurrent,
							HistoryLimit:            5,
							Type:                    v1alpha1.ScheduleTypePodChaos,
							ScheduleItem: v1alpha1.ScheduleItem{
								EmbedChaos: v1alpha1.EmbedChaos{
									PodChaos: &v1alpha1.PodChaosSpec{
										ContainerSelector: v1alpha1.ContainerSelector{
											PodSelector: v1alpha1.PodSelector{
												Selector: v1alpha1.PodSelectorSpec{
													Namespaces: []string{ns},
													LabelSelectors: map[string]string{
														"app": "not-actually-exist",
													},
												},
												Mode: v1alpha1.AllPodMode,
											},
											ContainerNames: nil,
										},
										Action: v1alpha1.PodKillAction,
									},
								},
							},
						},
					},
				}
				Expect(kubeClient.Create(ctx, &node)).To(Succeed())

				By("assert that schedule CR is created")
				Eventually(func() bool {
					updatedNode := v1alpha1.WorkflowNode{}
					Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: node.Name}, &updatedNode)).To(Succeed())
					if !ConditionEqualsTo(updatedNode.Status, v1alpha1.ConditionChaosInjected, corev1.ConditionTrue) {
						return false
					}
					schedule := v1alpha1.Schedule{}
					err := kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: updatedNode.Status.ChaosResource.Name}, &schedule)
					return err == nil
				}, toleratedJitter, time.Second).Should(BeTrue())

				By("assert that schedule should be purged")
				Eventually(func() bool {
					scheduleList := v1alpha1.ScheduleList{}
					Expect(kubeClient.List(ctx, &scheduleList, &client.ListOptions{Namespace: ns})).To(Succeed())
					return len(scheduleList.Items) == 0
				}, duration+toleratedJitter, time.Second).Should(BeTrue())
			})
		})

		Context("on serial", func() {
			It("should shutdown all children of serial", func() {
				// TODO: unfinished test case
			})
		})

		Context("on parallel", func() {
			It("should shutdown all children of parallel", func() {
				// TODO: unfinished test case
			})
		})

		Context("nested serial or parallel", func() {
			It("should shutdown children recursively", func() {
				// TODO: unfinished test case
			})
		})
	})
})
