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
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// unit tests
func Test_getTaskNameFromGeneratedName(t *testing.T) {
	type args struct {
		generatedNodeName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"common case",
			args{"name-1"},
			"name",
		}, {
			"common case",
			args{"name-1-2"},
			"name-1",
		}, {
			"common case",
			args{"name"},
			"name",
		}, {
			"common case",
			args{"name-"},
			"name",
		},
		{
			"common case",
			args{"-name"},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTaskNameFromGeneratedName(tt.args.generatedNodeName); got != tt.want {
				t.Errorf("getTaskNameFromGeneratedName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_relativeComplementSet(t *testing.T) {
	type args struct {
		former []string
		latter []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "common_case",
			args: args{
				former: []string{"a", "b", "c"},
				latter: []string{},
			},
			want: []string{"a", "b", "c"},
		}, {
			name: "common_case",
			args: args{
				former: []string{"a", "b", "c"},
				latter: []string{"b", "c"},
			},
			want: []string{"a"},
		}, {
			name: "common_case",
			args: args{
				former: []string{"a", "b", "c"},
				latter: []string{"c", "a"},
			},
			want: []string{"b"},
		}, {
			name: "common_case",
			args: args{
				former: []string{"a", "b", "c"},
				latter: []string{"c", "b", "d"},
			},
			want: []string{"a"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := setDifference(test.args.former, test.args.latter)
			sort.Strings(got)
			sort.Strings(test.want)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("getTaskNameFromGeneratedName() = %v, want %v", got, test.want)
			}
		})
	}
}

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

	Context("with one parallel node", func() {
		Context("with one simple parallel node", func() {

			It("should spawn all the children at the same time", func() {
				By("create simple workflow")
				ctx := context.TODO()
				simpleParallelWorkflow := v1alpha1.Workflow{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "simple-parallel",
						Namespace: ns,
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "parallel",
						Templates: []v1alpha1.Template{
							{
								Name: "parallel",
								Type: v1alpha1.TypeParallel,
								Children: []string{
									"network-chaos",
									"pod-chaos",
									"stress-chaos",
								},
							}, {
								Name: "network-chaos",
								Type: v1alpha1.TypeNetworkChaos,
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
								Name: "pod-chaos",
								Type: v1alpha1.TypePodChaos,
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
								Name: "stress-chaos",
								Type: v1alpha1.TypeStressChaos,
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
				Expect(kubeClient.Create(ctx, &simpleParallelWorkflow)).To(Succeed())

				By("assert that all resource has been created")

				By("assert that 1 entry node and 3 chaos nodes created")
				Eventually(func() int {
					workflowNodeList := v1alpha1.WorkflowNodeList{}
					Expect(kubeClient.List(ctx, &workflowNodeList, &client.ListOptions{Namespace: ns})).To(Succeed())
					return len(workflowNodeList.Items)
				}, 10*time.Second, time.Second).Should(Equal(4))

				By("assert that network chaos has been created")
				Eventually(func() bool {
					chaosList := v1alpha1.NetworkChaosList{}
					Expect(kubeClient.List(ctx, &chaosList, &client.ListOptions{Namespace: ns})).To(Succeed())
					if len(chaosList.Items) != 1 {
						return false
					}
					return strings.HasPrefix(chaosList.Items[0].Name, "network-chaos")
				}, 10*time.Second, time.Second).Should(BeTrue())

				By("assert that pod chaos has been created")
				Eventually(func() bool {
					chaosList := v1alpha1.PodChaosList{}
					Expect(kubeClient.List(ctx, &chaosList, &client.ListOptions{Namespace: ns})).To(Succeed())
					if len(chaosList.Items) != 1 {
						return false
					}
					return strings.HasPrefix(chaosList.Items[0].Name, "pod-chaos")
				}, 10*time.Second, time.Second).Should(BeTrue())

				By("assert that stress chaos has been created")
				Eventually(func() bool {
					chaosList := v1alpha1.StressChaosList{}
					Expect(kubeClient.List(ctx, &chaosList, &client.ListOptions{Namespace: ns})).To(Succeed())
					if len(chaosList.Items) != 1 {
						return false
					}
					return strings.HasPrefix(chaosList.Items[0].Name, "stress-chaos")
				}, 10*time.Second, time.Second).Should(BeTrue())
			})
		})
	})
})
