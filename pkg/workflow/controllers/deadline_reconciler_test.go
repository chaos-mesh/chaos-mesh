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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/pointer"
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
											GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
												Namespaces: []string{ns},
												LabelSelectors: map[string]string{
													"app": "not-actually-exist",
												},
											},
										},
										Mode: v1alpha1.AllMode,
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
													GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
														Namespaces: []string{ns},
														LabelSelectors: map[string]string{
															"app": "not-actually-exist",
														},
													},
												},
												Mode: v1alpha1.AllMode,
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
				ctx := context.TODO()
				serialDuration := 3 * time.Second
				durationOfSubTask1 := time.Second
				durationOfSubTask2 := 5 * time.Second
				durationOfSubTask3 := 5 * time.Second
				toleratedJitter := 2 * time.Second

				maxConsisting := durationOfSubTask1 + durationOfSubTask2 + durationOfSubTask3

				workflow := v1alpha1.Workflow{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:    ns,
						GenerateName: "fake-workflow-serial-",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "entry-serial",
						Templates: []v1alpha1.Template{{
							Name:     "entry-serial",
							Type:     v1alpha1.TypeSerial,
							Deadline: pointer.StringPtr(serialDuration.String()),
							Children: []string{
								"serial-task-1",
								"serial-task-2",
								"serial-task-3",
							},
						}, {
							Name:     "serial-task-1",
							Type:     v1alpha1.TypeSuspend,
							Deadline: pointer.StringPtr(durationOfSubTask1.String()),
						}, {
							Name:     "serial-task-2",
							Type:     v1alpha1.TypeSuspend,
							Deadline: pointer.StringPtr(durationOfSubTask2.String()),
						}, {
							Name:     "serial-task-3",
							Type:     v1alpha1.TypeSuspend,
							Deadline: pointer.StringPtr(durationOfSubTask3.String()),
						}},
					},
					Status: v1alpha1.WorkflowStatus{},
				}

				By("create workflow with serial entry")
				Expect(kubeClient.Create(ctx, &workflow)).To(Succeed())

				By("task 1 should be created")
				task1Name := ""
				Eventually(func() bool {
					workflowNodes := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodes)).To(Succeed())
					for _, item := range workflowNodes.Items {
						if strings.HasPrefix(item.Name, "serial-task-1") {
							task1Name = item.Name
							return true
						}
					}
					return false
				}, toleratedJitter, 200*time.Millisecond).Should(BeTrue())

				Expect(task1Name).NotTo(BeEmpty())

				By("task 1 will be DeadlineExceed by itself")
				Eventually(func() bool {
					taskNode1 := v1alpha1.WorkflowNode{}
					Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: task1Name}, &taskNode1)).To(Succeed())
					condition := GetCondition(taskNode1.Status, v1alpha1.ConditionDeadlineExceed)
					if condition == nil {
						return false
					}
					if condition.Status != corev1.ConditionTrue {
						return false
					}
					if condition.Reason != v1alpha1.NodeDeadlineExceed {
						return false
					}
					return true
				}, durationOfSubTask1+toleratedJitter, 200*time.Millisecond).Should(BeTrue())

				By("task 2 should be created")
				task2Name := ""
				Eventually(func() bool {
					workflowNodes := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodes)).To(Succeed())
					for _, item := range workflowNodes.Items {
						if strings.HasPrefix(item.Name, "serial-task-2") {
							task2Name = item.Name
							return true
						}
					}
					return false
				}, toleratedJitter, 200*time.Millisecond).Should(BeTrue())
				Expect(task2Name).NotTo(BeEmpty())

				By("task 2 should be DeadlineExceed by parent")
				taskNode2 := v1alpha1.WorkflowNode{}
				Eventually(func() bool {
					Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: task2Name}, &taskNode2)).To(Succeed())
					condition := GetCondition(taskNode2.Status, v1alpha1.ConditionDeadlineExceed)
					if condition == nil {
						return false
					}
					if condition.Status != corev1.ConditionTrue {
						return false
					}
					if condition.Reason != v1alpha1.ParentNodeDeadlineExceed {
						return false
					}
					return true
				}, durationOfSubTask1+toleratedJitter, 200*time.Millisecond).Should(BeTrue())

				By("entry serial should also be DeadlineExceed by itself")
				entryNode := v1alpha1.WorkflowNode{}
				entryNodeName := taskNode2.Labels[v1alpha1.LabelControlledBy]
				Expect(entryNodeName).NotTo(BeEmpty())
				Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: entryNodeName}, &entryNode)).To(Succeed())
				condition := GetCondition(entryNode.Status, v1alpha1.ConditionDeadlineExceed)
				Expect(condition).NotTo(BeNil())
				Expect(condition.Status).To(Equal(corev1.ConditionTrue))
				Expect(condition.Reason).To(Equal(v1alpha1.NodeDeadlineExceed))

				By("task 3 should NEVER be created")
				Consistently(
					func() bool {
						workflowNodes := v1alpha1.WorkflowNodeList{}
						Expect(kubeClient.List(ctx, &workflowNodes)).To(Succeed())
						for _, item := range workflowNodes.Items {
							if strings.HasPrefix(item.Name, "serial-task-3") {
								return false
							}
						}
						return true
					},
					maxConsisting+toleratedJitter, time.Second).Should(BeTrue())
			})
		})

		Context("on parallel", func() {
			It("should shutdown all children of parallel", func() {
				ctx := context.TODO()
				parallelDuration := 3 * time.Second
				durationOfSubTask1 := time.Second
				durationOfSubTask2 := 5 * time.Second
				durationOfSubTask3 := 5 * time.Second
				toleratedJitter := 2 * time.Second

				workflow := v1alpha1.Workflow{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:    ns,
						GenerateName: "fake-workflow-parallel-",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "entry-parallel",
						Templates: []v1alpha1.Template{{
							Name:     "entry-parallel",
							Type:     v1alpha1.TypeParallel,
							Deadline: pointer.StringPtr(parallelDuration.String()),
							Children: []string{
								"parallel-task-1",
								"parallel-task-2",
								"parallel-task-3",
							},
						}, {
							Name:     "parallel-task-1",
							Type:     v1alpha1.TypeSuspend,
							Deadline: pointer.StringPtr(durationOfSubTask1.String()),
						}, {
							Name:     "parallel-task-2",
							Type:     v1alpha1.TypeSuspend,
							Deadline: pointer.StringPtr(durationOfSubTask2.String()),
						}, {
							Name:     "parallel-task-3",
							Type:     v1alpha1.TypeSuspend,
							Deadline: pointer.StringPtr(durationOfSubTask3.String()),
						}},
					},
					Status: v1alpha1.WorkflowStatus{},
				}

				By("create workflow with parallel entry")
				Expect(kubeClient.Create(ctx, &workflow)).To(Succeed())

				By("task 1,task 2,task 3 should be created")
				task1Name := ""
				task2Name := ""
				task3Name := ""
				Eventually(func() bool {
					workflowNodes := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodes)).To(Succeed())
					for _, item := range workflowNodes.Items {
						if strings.HasPrefix(item.Name, "parallel-task-1") {
							task1Name = item.Name
							return true
						}
					}
					return false
				}, toleratedJitter, 200*time.Millisecond).Should(BeTrue())
				Eventually(func() bool {
					workflowNodes := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodes)).To(Succeed())
					for _, item := range workflowNodes.Items {
						if strings.HasPrefix(item.Name, "parallel-task-2") {
							task2Name = item.Name
							return true
						}
					}
					return false
				}, toleratedJitter, 200*time.Millisecond).Should(BeTrue())
				Eventually(func() bool {
					workflowNodes := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodes)).To(Succeed())
					for _, item := range workflowNodes.Items {
						if strings.HasPrefix(item.Name, "parallel-task-3") {
							task3Name = item.Name
							return true
						}
					}
					return false
				}, toleratedJitter, 200*time.Millisecond).Should(BeTrue())

				Expect(task1Name).NotTo(BeEmpty())
				Expect(task2Name).NotTo(BeEmpty())
				Expect(task3Name).NotTo(BeEmpty())

				By("task 1 should be DeadlineExceed by itself")
				Eventually(func() bool {
					taskNode := v1alpha1.WorkflowNode{}
					Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: task1Name}, &taskNode)).To(Succeed())
					condition := GetCondition(taskNode.Status, v1alpha1.ConditionDeadlineExceed)
					if condition == nil {
						return false
					}
					if condition.Status != corev1.ConditionTrue {
						return false
					}
					if condition.Reason != v1alpha1.NodeDeadlineExceed {
						return false
					}
					return true
				}, durationOfSubTask1+toleratedJitter, 200*time.Millisecond).Should(BeTrue())

				By("task 2 and task 3 should be DeadlineExceed by parent")
				for _, nodeName := range []string{task2Name, task3Name} {
					Eventually(func() bool {
						taskNode := v1alpha1.WorkflowNode{}
						Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: nodeName}, &taskNode)).To(Succeed())
						condition := GetCondition(taskNode.Status, v1alpha1.ConditionDeadlineExceed)
						if condition == nil {
							return false
						}
						if condition.Status != corev1.ConditionTrue {
							return false
						}
						if condition.Reason != v1alpha1.ParentNodeDeadlineExceed {
							return false
						}
						return true
					}, parallelDuration+toleratedJitter, 200*time.Millisecond).Should(BeTrue())
				}
				By("entry parallel should also be DeadlineExceed by itself")
				updateWorkflow := v1alpha1.Workflow{}
				Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: workflow.Name}, &updateWorkflow)).To(Succeed())
				entryNodeName := updateWorkflow.Status.EntryNode
				Expect(entryNodeName).NotTo(BeNil())
				Expect(*entryNodeName).NotTo(BeEmpty())
				entryNode := v1alpha1.WorkflowNode{}
				Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: *entryNodeName}, &entryNode)).To(Succeed())
				condition := GetCondition(entryNode.Status, v1alpha1.ConditionDeadlineExceed)
				Expect(condition).NotTo(BeNil())
				Expect(condition.Status).To(Equal(corev1.ConditionTrue))
				Expect(condition.Reason).To(Equal(v1alpha1.NodeDeadlineExceed))
			})
		})

		Context("nested serial or parallel", func() {
			It("should shutdown children recursively", func() {
				ctx := context.TODO()
				parallelDuration := 3 * time.Second
				durationOfSuspend := 10 * time.Second
				toleratedJitter := 2 * time.Second

				workflow := v1alpha1.Workflow{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:    ns,
						GenerateName: "fake-workflow-parallel-",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "entry-parallel",
						Templates: []v1alpha1.Template{{
							Name:     "entry-parallel",
							Type:     v1alpha1.TypeParallel,
							Deadline: pointer.StringPtr(parallelDuration.String()),
							Children: []string{
								"parallel-level-1",
							},
						}, {
							Name: "parallel-level-1",
							Type: v1alpha1.TypeParallel,
							Children: []string{
								"parallel-level-2",
							},
						}, {
							Name: "parallel-level-2",
							Type: v1alpha1.TypeParallel,
							Children: []string{
								"suspend-task",
							},
						}, {
							Name:     "suspend-task",
							Type:     v1alpha1.TypeSuspend,
							Deadline: pointer.StringPtr(durationOfSuspend.String()),
						}},
					},
					Status: v1alpha1.WorkflowStatus{},
				}

				By("create workflow with parallel entry")
				Expect(kubeClient.Create(ctx, &workflow)).To(Succeed())

				By("all the node should be created")
				parallelLevel1NodeName := ""
				parallelLevel2NodeName := ""
				suspendTaskNodeName := ""
				Eventually(func() bool {
					workflowNodes := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodes)).To(Succeed())
					for _, item := range workflowNodes.Items {
						if strings.HasPrefix(item.Name, "parallel-level-1") {
							parallelLevel1NodeName = item.Name
							return true
						}
					}
					return false
				}, toleratedJitter, 200*time.Millisecond).Should(BeTrue())
				Eventually(func() bool {
					workflowNodes := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodes)).To(Succeed())
					for _, item := range workflowNodes.Items {
						if strings.HasPrefix(item.Name, "parallel-level-2") {
							parallelLevel2NodeName = item.Name
							return true
						}
					}
					return false
				}, toleratedJitter, 200*time.Millisecond).Should(BeTrue())
				Eventually(func() bool {
					workflowNodes := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodes)).To(Succeed())
					for _, item := range workflowNodes.Items {
						if strings.HasPrefix(item.Name, "suspend-task") {
							suspendTaskNodeName = item.Name
							return true
						}
					}
					return false
				}, toleratedJitter, 200*time.Millisecond).Should(BeTrue())

				Expect(parallelLevel1NodeName).NotTo(BeEmpty())
				Expect(parallelLevel2NodeName).NotTo(BeEmpty())
				Expect(suspendTaskNodeName).NotTo(BeEmpty())

				By("parallel level 1, parallel level 2 and suspend task should be DeadlineExceed by parent")
				for _, nodeName := range []string{parallelLevel1NodeName, parallelLevel2NodeName, suspendTaskNodeName} {
					Eventually(func() bool {
						taskNode := v1alpha1.WorkflowNode{}
						Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: nodeName}, &taskNode)).To(Succeed())
						condition := GetCondition(taskNode.Status, v1alpha1.ConditionDeadlineExceed)
						if condition == nil {
							return false
						}
						if condition.Status != corev1.ConditionTrue {
							return false
						}
						if condition.Reason != v1alpha1.ParentNodeDeadlineExceed {
							return false
						}
						return true
					}, parallelDuration+toleratedJitter, 200*time.Millisecond).Should(BeTrue())
				}

				By("entry parallel should also be DeadlineExceed by itself")
				updateWorkflow := v1alpha1.Workflow{}
				Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: workflow.Name}, &updateWorkflow)).To(Succeed())
				entryNodeName := updateWorkflow.Status.EntryNode
				Expect(entryNodeName).NotTo(BeNil())
				Expect(*entryNodeName).NotTo(BeEmpty())
				entryNode := v1alpha1.WorkflowNode{}
				Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: *entryNodeName}, &entryNode)).To(Succeed())
				condition := GetCondition(entryNode.Status, v1alpha1.ConditionDeadlineExceed)
				Expect(condition).NotTo(BeNil())
				Expect(condition.Status).To(Equal(corev1.ConditionTrue))
				Expect(condition.Reason).To(Equal(v1alpha1.NodeDeadlineExceed))

			})
		})

		Context("if this node is already in DeadlineExceed because of ParentNodeDeadlineExceed", func() {
			It("should omit the next coming deadline", func() {
				ctx := context.TODO()
				now := time.Now()
				duration := 5 * time.Second
				toleratedJitter := 3 * time.Second

				startTime := metav1.NewTime(now)
				deadline := metav1.NewTime(now.Add(duration))

				By("create one empty podchaos workflow node, with deadline: 3s")
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
											GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
												Namespaces: []string{ns},
												LabelSelectors: map[string]string{
													"app": "not-actually-exist",
												},
											},
										},
										Mode: v1alpha1.AllMode,
									},
									ContainerNames: nil,
								},
								Action: v1alpha1.PodKillAction,
							},
						},
					},
					Status: v1alpha1.WorkflowNodeStatus{},
				}
				Expect(kubeClient.Create(ctx, &node)).To(Succeed())
				By("manually set condition ConditionDeadlineExceed to true, because of v1alpha1.ParentNodeDeadlineExceed")
				updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					deadlineExceedNode := v1alpha1.WorkflowNode{}

					err := kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: node.Name}, &deadlineExceedNode)
					if err != nil {
						return err
					}
					deadlineExceedNode.Status.Conditions = []v1alpha1.WorkflowNodeCondition{
						{
							Type:   v1alpha1.ConditionDeadlineExceed,
							Status: corev1.ConditionTrue,
							Reason: v1alpha1.ParentNodeDeadlineExceed,
						},
					}
					err = kubeClient.Status().Update(ctx, &deadlineExceedNode)
					if err != nil {
						return err
					}
					return nil
				})
				Expect(updateError).To(BeNil())
				By("after 3 seconds, the condition ConditionDeadlineExceed should not be modified")
				Consistently(func() bool {
					updatedNode := v1alpha1.WorkflowNode{}
					Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: node.Name}, &updatedNode)).To(Succeed())

					condition := GetCondition(updatedNode.Status, v1alpha1.ConditionDeadlineExceed)
					if condition == nil {
						return false
					}
					if condition.Status != corev1.ConditionTrue {
						return false
					}
					if condition.Reason != v1alpha1.ParentNodeDeadlineExceed {
						return false
					}
					return true
				},
					duration+toleratedJitter, time.Second,
				).Should(BeTrue())
			})

			It("should NOT omit the next coming deadline otherwise", func() {
				ctx := context.TODO()
				now := time.Now()
				duration := 5 * time.Second
				toleratedJitter := 3 * time.Second

				startTime := metav1.NewTime(now)
				deadline := metav1.NewTime(now.Add(duration))

				By("create one empty podchaos workflow node, with deadline: 3s")
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
											GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
												Namespaces: []string{ns},
												LabelSelectors: map[string]string{
													"app": "not-actually-exist",
												},
											},
										},
										Mode: v1alpha1.AllMode,
									},
									ContainerNames: nil,
								},
								Action: v1alpha1.PodKillAction,
							},
						},
					},
					Status: v1alpha1.WorkflowNodeStatus{},
				}
				Expect(kubeClient.Create(ctx, &node)).To(Succeed())
				By("manually set condition ConditionDeadlineExceed to true, but NOT caused by v1alpha1.ParentNodeDeadlineExceed")
				updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					deadlineExceedNode := v1alpha1.WorkflowNode{}

					err := kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: node.Name}, &deadlineExceedNode)
					if err != nil {
						return err
					}
					deadlineExceedNode.Status.Conditions = []v1alpha1.WorkflowNodeCondition{
						{
							Type:   v1alpha1.ConditionDeadlineExceed,
							Status: corev1.ConditionTrue,
							Reason: v1alpha1.NodeDeadlineExceed,
						},
					}
					err = kubeClient.Status().Update(ctx, &deadlineExceedNode)
					if err != nil {
						return err
					}
					return nil
				})
				Expect(updateError).To(BeNil())
				By("condition ConditionDeadlineExceed should be corrected soon")
				Eventually(func() bool {
					updatedNode := v1alpha1.WorkflowNode{}
					Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: node.Name}, &updatedNode)).To(Succeed())
					return ConditionEqualsTo(updatedNode.Status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionFalse)
				},
					toleratedJitter,
					time.Second)
				By("after 5 seconds, the condition ConditionDeadlineExceed should not be modified, caused by NodeDeadlineExceed itself")
				Eventually(func() bool {
					updatedNode := v1alpha1.WorkflowNode{}
					Expect(kubeClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: node.Name}, &updatedNode)).To(Succeed())

					condition := GetCondition(updatedNode.Status, v1alpha1.ConditionDeadlineExceed)
					if condition == nil {
						return false
					}
					if condition.Status != corev1.ConditionTrue {
						return false
					}
					if condition.Reason != v1alpha1.NodeDeadlineExceed {
						return false
					}
					return true
				},
					duration+toleratedJitter, time.Second,
				).Should(BeTrue())
			})
		})
	})
})
