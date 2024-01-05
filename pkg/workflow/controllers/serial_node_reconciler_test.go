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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
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

	Context("with one serial node", func() {
		Context("with one simple serial node", func() {

			It("should spawn all the children one by one", func() {
				By("create simple workflow")
				ctx := context.TODO()

				networkChaosDuration := 5 * time.Second
				networkChaosDurationString := networkChaosDuration.String()
				podChaosDuration := 7 * time.Second
				podChaosDurationString := podChaosDuration.String()
				stressChaosDuration := 9 * time.Second
				stressChaosDurationString := stressChaosDuration.String()

				toleratedJitter := 10 * time.Second

				simpleSerialWorkflow := v1alpha1.Workflow{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple-serial",
						Namespace: ns,
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "serial",
						Templates: []v1alpha1.Template{
							{
								Name: "serial",
								Type: v1alpha1.TypeSerial,
								Children: []string{
									"network-chaos",
									"pod-chaos",
									"stress-chaos",
								},
							}, {
								Name:     "network-chaos",
								Type:     v1alpha1.TypeNetworkChaos,
								Deadline: &networkChaosDurationString,
								EmbedChaos: &v1alpha1.EmbedChaos{
									NetworkChaos: &v1alpha1.NetworkChaosSpec{
										PodSelector: v1alpha1.PodSelector{
											Selector: v1alpha1.PodSelectorSpec{
												GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
													Namespaces: []string{ns},
													LabelSelectors: map[string]string{
														"app": "not-exist",
													},
												},
											},
											Mode: v1alpha1.AllMode,
										},
										Action: v1alpha1.PartitionAction,
									},
								},
							}, {
								Name:     "pod-chaos",
								Type:     v1alpha1.TypePodChaos,
								Deadline: &podChaosDurationString,
								EmbedChaos: &v1alpha1.EmbedChaos{
									PodChaos: &v1alpha1.PodChaosSpec{
										ContainerSelector: v1alpha1.ContainerSelector{
											PodSelector: v1alpha1.PodSelector{
												Selector: v1alpha1.PodSelectorSpec{
													GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
														Namespaces: []string{ns},
														LabelSelectors: map[string]string{
															"app": "not-exist",
														},
													},
												},
												Mode: v1alpha1.AllMode,
											},
										},
										Action: v1alpha1.PodKillAction,
									},
								},
							},
							{
								Name:     "stress-chaos",
								Type:     v1alpha1.TypeStressChaos,
								Deadline: &stressChaosDurationString,
								EmbedChaos: &v1alpha1.EmbedChaos{
									StressChaos: &v1alpha1.StressChaosSpec{
										ContainerSelector: v1alpha1.ContainerSelector{
											PodSelector: v1alpha1.PodSelector{
												Selector: v1alpha1.PodSelectorSpec{
													GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
														Namespaces: []string{ns},
														LabelSelectors: map[string]string{
															"app": "not-exist",
														},
													},
												},
												Mode: v1alpha1.AllMode,
											},
										},
										Stressors: &v1alpha1.Stressors{
											CPUStressor: &v1alpha1.CPUStressor{
												Stressor: v1alpha1.Stressor{
													Workers: 2,
												},
											}},
									},
								},
							},
						},
					}}
				Expect(kubeClient.Create(ctx, &simpleSerialWorkflow)).To(Succeed())

				By("assert that all resource has been created")

				By("assert that entry node created")
				Eventually(func() int {
					workflowNodeList := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodeList, &client.ListOptions{Namespace: ns})).To(Succeed())
					return len(workflowNodeList.Items)
				}, 10*time.Second, time.Second).Should(BeNumerically(">=", 1))

				By("assert that network chaos has been created")
				Eventually(func() bool {
					chaosList := v1alpha1.NetworkChaosList{}
					Expect(kubeClient.List(ctx, &chaosList, &client.ListOptions{Namespace: ns})).To(Succeed())
					if len(chaosList.Items) != 1 {
						return false
					}
					return strings.HasPrefix(chaosList.Items[0].Name, "network-chaos")
				}, toleratedJitter, time.Second).Should(BeTrue())

				By("assert that network chaos has been deleted")
				Eventually(func() int {
					chaosList := v1alpha1.NetworkChaosList{}
					Expect(kubeClient.List(ctx, &chaosList, &client.ListOptions{Namespace: ns})).To(Succeed())
					return len(chaosList.Items)
				}, networkChaosDuration+toleratedJitter, time.Second).Should(BeZero())

				By("assert that pod chaos has been created")
				Eventually(func() bool {
					chaosList := v1alpha1.PodChaosList{}
					Expect(kubeClient.List(ctx, &chaosList, &client.ListOptions{Namespace: ns})).To(Succeed())
					if len(chaosList.Items) != 1 {
						return false
					}
					return strings.HasPrefix(chaosList.Items[0].Name, "pod-chaos")
				}, toleratedJitter, time.Second).Should(BeTrue())

				By("assert that pod chaos has been deleted")
				Eventually(func() int {
					chaosList := v1alpha1.PodChaosList{}
					Expect(kubeClient.List(ctx, &chaosList, &client.ListOptions{Namespace: ns})).To(Succeed())
					return len(chaosList.Items)
				}, podChaosDuration+toleratedJitter, time.Second).Should(BeZero())

				By("assert that stress chaos has been created")
				Eventually(func() bool {
					chaosList := v1alpha1.StressChaosList{}
					Expect(kubeClient.List(ctx, &chaosList, &client.ListOptions{Namespace: ns})).To(Succeed())
					if len(chaosList.Items) != 1 {
						return false
					}
					return strings.HasPrefix(chaosList.Items[0].Name, "stress-chaos")
				}, toleratedJitter, time.Second).Should(BeTrue())

				By("assert that stress chaos has been deleted")
				Eventually(func() int {
					chaosList := v1alpha1.StressChaosList{}
					Expect(kubeClient.List(ctx, &chaosList, &client.ListOptions{Namespace: ns})).To(Succeed())
					return len(chaosList.Items)
				}, stressChaosDuration+toleratedJitter, time.Second).Should(BeZero())

				By("assert that serial node marked as finished")
				Eventually(func() bool {
					workflowNodeList := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodeList, &client.ListOptions{Namespace: ns})).To(Succeed())
					if len(workflowNodeList.Items) != 4 {
						return false
					}
					entryFounded := false
					var entry *v1alpha1.WorkflowNode = nil
					for _, item := range workflowNodeList.Items {
						item := item
						if item.Spec.Type == v1alpha1.TypeSerial {
							entryFounded = true
							entry = &item
						}
					}
					if !entryFounded || entry == nil {
						return false
					}
					return ConditionEqualsTo(entry.Status, v1alpha1.ConditionAccomplished, corev1.ConditionTrue)
				}, toleratedJitter, time.Second).Should(BeTrue())
			})
		})

		Context("with statuscheck node", func() {
			Context("statuscheck aborted, AbortWithStatusCheck=true", func() {
				It("aborts workflow", func() {
					By("create workflow")
					ctx := context.TODO()

					statusCheckDuration := 5 * time.Second
					statusCheckDurationString := statusCheckDuration.String()
					podChaosDuration := 7 * time.Second
					podChaosDurationString := podChaosDuration.String()

					statusCheckWorkflow := v1alpha1.Workflow{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "statuscheck-abort",
							Namespace: ns,
						},
						Spec: v1alpha1.WorkflowSpec{
							Entry: "statuscheck-abort-serial",
							Templates: []v1alpha1.Template{
								{
									Name: "statuscheck-abort-serial",
									Type: v1alpha1.TypeSerial,
									Children: []string{
										"http-check",
										"pod-chaos",
									},
								}, {
									Name:                 "http-check",
									Type:                 v1alpha1.TypeStatusCheck,
									AbortWithStatusCheck: true,
									Deadline:             &statusCheckDurationString,
									StatusCheck: &v1alpha1.StatusCheckSpec{
										Mode:             v1alpha1.StatusCheckSynchronous,
										Type:             v1alpha1.TypeHTTP,
										TimeoutSeconds:   1,
										IntervalSeconds:  1,
										FailureThreshold: 1,
										SuccessThreshold: 1,
										EmbedStatusCheck: &v1alpha1.EmbedStatusCheck{
											HTTPStatusCheck: &v1alpha1.HTTPStatusCheck{
												RequestUrl: "http://127.0.0.1:63123",
												Criteria: v1alpha1.HTTPCriteria{
													StatusCode: "200",
												},
											},
										},
									},
								}, {
									Name:     "pod-chaos",
									Type:     v1alpha1.TypePodChaos,
									Deadline: &podChaosDurationString,
									EmbedChaos: &v1alpha1.EmbedChaos{
										PodChaos: &v1alpha1.PodChaosSpec{
											ContainerSelector: v1alpha1.ContainerSelector{
												PodSelector: v1alpha1.PodSelector{
													Selector: v1alpha1.PodSelectorSpec{
														GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
															Namespaces: []string{ns},
															LabelSelectors: map[string]string{
																"app": "not-exist",
															},
														},
													},
													Mode: v1alpha1.AllMode,
												},
											},
											Action: v1alpha1.PodKillAction,
										},
									},
								},
							},
						},
					}
					Expect(kubeClient.Create(ctx, &statusCheckWorkflow)).To(Succeed(), "workflow create")

					extractNodeType := func(n v1alpha1.WorkflowNode) v1alpha1.TemplateType {
						return n.Spec.Type
					}

					By("assert that expected workflow nodes have been created")
					Eventually(func(g Gomega) {
						workflowNodeList := v1alpha1.WorkflowNodeList{}
						g.Expect(kubeClient.List(ctx, &workflowNodeList, &client.ListOptions{Namespace: ns})).
							To(Succeed(), "failed to list workflownodes")

						g.Expect(workflowNodeList.Items).To(ConsistOf(
							WithTransform(extractNodeType, Equal(v1alpha1.TypeSerial)),
							WithTransform(extractNodeType, Equal(v1alpha1.TypeStatusCheck)),
						))
					}, 10*time.Second, time.Second).Should(Succeed())

					By("assert that status check is completed and failed")
					Eventually(func(g Gomega) {
						list := v1alpha1.StatusCheckList{}
						g.Expect(kubeClient.List(ctx, &list, &client.ListOptions{Namespace: ns})).To(Succeed())
						g.Expect(len(list.Items)).To(Equal(1))
						g.Expect(list.Items[0].Spec.Type).To(Equal(v1alpha1.TypeHTTP))
						g.Expect(list.Items[0].Name).To(HavePrefix("http-check"))

						g.Expect(list.Items[0].Status.Conditions).To(
							ConsistOf(
								MatchFields(IgnoreExtras, Fields{
									"Type":   Equal(v1alpha1.StatusCheckConditionCompleted),
									"Status": Equal(corev1.ConditionTrue),
								}),
								MatchFields(IgnoreExtras, Fields{
									"Type":   Equal(v1alpha1.StatusCheckConditionDurationExceed),
									"Status": Equal(corev1.ConditionFalse),
								}),
								MatchFields(IgnoreExtras, Fields{
									"Type":   Equal(v1alpha1.StatusCheckConditionFailureThresholdExceed),
									"Status": Equal(corev1.ConditionTrue),
								}),
								MatchFields(IgnoreExtras, Fields{
									"Type":   Equal(v1alpha1.StatusCheckConditionSuccessThresholdExceed),
									"Status": Equal(corev1.ConditionFalse),
								}),
							),
						)
					}, 10*time.Second, time.Second).Should(Succeed())

					By("assert that workflow aborts without creating further nodes")
					Eventually(func(g Gomega) {
						workflowNodeList := v1alpha1.WorkflowNodeList{}
						g.Expect(kubeClient.List(ctx, &workflowNodeList, &client.ListOptions{Namespace: ns})).
							To(Succeed(), "failed to list workflownodes")

						// pod chaos node should not have been created
						// if pod chaos node has not been created, and the entry node is aborted, assume it will
						// not be created
						g.Expect(workflowNodeList.Items).To(ConsistOf(
							WithTransform(extractNodeType, Equal(v1alpha1.TypeSerial)),
							WithTransform(extractNodeType, Equal(v1alpha1.TypeStatusCheck)),
						))

						var entryNode *v1alpha1.WorkflowNode
						var statusCheckNode *v1alpha1.WorkflowNode

						for _, item := range workflowNodeList.Items {
							item := item
							if item.Spec.Type == v1alpha1.TypeSerial {
								entryNode = &item
							} else if item.Spec.Type == v1alpha1.TypeStatusCheck {
								statusCheckNode = &item
							}
						}

						g.Expect(entryNode).NotTo(BeNil(), "entry workflow node not found")
						g.Expect(statusCheckNode).NotTo(BeNil(), "status check workflow node not found")

						g.Expect(ConditionEqualsTo(statusCheckNode.Status, v1alpha1.ConditionAborted, corev1.ConditionTrue)).
							To(BeTrue(), "status check node should be aborted")

						g.Expect(ConditionEqualsTo(entryNode.Status, v1alpha1.ConditionAborted, corev1.ConditionTrue)).
							To(BeTrue(), "entry node should be aborted")
					}, 10*time.Second, time.Second).Should(Succeed())
				})
			})
		})
	})
})
