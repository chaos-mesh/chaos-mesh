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
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
												Namespaces: []string{ns},
												LabelSelectors: map[string]string{
													"app": "not-exist",
												},
											},
											Mode: v1alpha1.AllPodMode,
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
													Namespaces: []string{ns},
													LabelSelectors: map[string]string{
														"app": "not-exist",
													},
												},
												Mode: v1alpha1.AllPodMode,
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
													Namespaces: []string{ns},
													LabelSelectors: map[string]string{
														"app": "not-exist",
													},
												},
												Mode: v1alpha1.AllPodMode,
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
	})
})
