// Copyright 2019 Chaos Mesh Authors.
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

package pod

import (
	"context"
	"testing"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

	. "github.com/onsi/gomega"

	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSelectPods(t *testing.T) {
	g := NewGomegaWithT(t)

	objects, pods := GenerateNPods("p", 5, PodArg{Labels: map[string]string{"l1": "l1"}, Nodename: "az1-node1"})
	objects2, pods2 := GenerateNPods("s", 2, PodArg{Namespace: "test-s", Labels: map[string]string{"l2": "l2"}, Nodename: "az2-node1"})

	objects3, _ := GenerateNNodes("az1-node", 3, map[string]string{"disktype": "ssd", "zone": "az1"})
	objects4, _ := GenerateNNodes("az2-node", 2, map[string]string{"disktype": "hdd", "zone": "az2"})

	objects = append(objects, objects2...)
	objects = append(objects, objects3...)
	objects = append(objects, objects4...)

	pods = append(pods, pods2...)

	c := fake.NewClientBuilder().WithRuntimeObjects(objects...).Build()
	var r client.Reader

	type TestCase struct {
		name         string
		selector     v1alpha1.PodSelectorSpec
		expectedPods []v1.Pod
	}

	tcs := []TestCase{
		{
			name: "filter specified pods",
			selector: v1alpha1.PodSelectorSpec{
				Pods: map[string][]string{
					metav1.NamespaceDefault: {"p3", "p4"},
					"test-s":                {"s1"},
				},
			},
			expectedPods: []v1.Pod{pods[3], pods[4], pods[6]},
		},
		{
			name: "filter labels pods",
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"l2": "l2"},
				},
			},
			expectedPods: []v1.Pod{pods[5], pods[6]},
		},
		{
			name: "filter pods by label expressions",
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					ExpressionSelectors: []metav1.LabelSelectorRequirement{
						{
							Key:      "l2",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"l2"},
						},
					},
				},
			},
			expectedPods: []v1.Pod{pods[5], pods[6]},
		},
		{
			name: "filter pods by label selectors and expression selectors",
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"l1": "l1"},
					ExpressionSelectors: []metav1.LabelSelectorRequirement{
						{
							Key:      "l2",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"l2"},
						},
					},
				},
			},
			expectedPods: nil,
		},
		{
			name: "filter namespace and labels",
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces:     []string{"test-s"},
					LabelSelectors: map[string]string{"l2": "l2"},
				},
			},
			expectedPods: []v1.Pod{pods[5], pods[6]},
		},
		{
			name: "filter namespace and labels",
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces:     []string{metav1.NamespaceDefault},
					LabelSelectors: map[string]string{"l2": "l2"},
				},
			},
			expectedPods: nil,
		},
		{
			name: "filter by specified node",
			selector: v1alpha1.PodSelectorSpec{
				Nodes: []string{"az1-node1"},
			},
			expectedPods: []v1.Pod{pods[0], pods[1], pods[2], pods[3], pods[4]},
		},
		{
			name: "filter node and labels",
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"l1": "l1"},
				},
				Nodes: []string{"az2-node1"},
			},
			expectedPods: nil,
		},
		{
			name: "filter pods by nodeSelector",
			selector: v1alpha1.PodSelectorSpec{
				NodeSelectors: map[string]string{"disktype": "hdd"},
			},
			expectedPods: []v1.Pod{pods[5], pods[6]},
		},
		{
			name: "filter pods by node and nodeSelector",
			selector: v1alpha1.PodSelectorSpec{
				NodeSelectors: map[string]string{"zone": "az1"},
				Nodes:         []string{"az2-node1"},
			},
			expectedPods: []v1.Pod{pods[0], pods[1], pods[2], pods[3], pods[4], pods[5], pods[6]},
		},
	}

	var (
		testCfgClusterScoped   = true
		testCfgTargetNamespace = ""
	)

	for _, tc := range tcs {
		filteredPods, err := SelectPods(context.Background(), c, r, tc.selector, testCfgClusterScoped, testCfgTargetNamespace, false)
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(len(filteredPods)).To(Equal(len(tc.expectedPods)), tc.name)
	}
}

func TestCheckPodMeetSelector(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		selector      v1alpha1.PodSelectorSpec
		pod           v1.Pod
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name: "meet label",
			pod:  NewPod(PodArg{Name: "t1", Status: v1.PodPending, Labels: map[string]string{"app": "tikv", "ss": "t1"}}),
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"app": "tikv"},
				},
			},
			expectedValue: true,
		},
		{
			name: "not meet label",
			pod:  NewPod(PodArg{Name: "t1", Labels: map[string]string{"app": "tidb", "ss": "t1"}}),
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"app": "tikv"},
				},
			},
			expectedValue: false,
		},
		{
			name: "pod labels is empty",
			pod:  NewPod(PodArg{Name: "t1"}),
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"app": "tikv"},
				},
			},
			expectedValue: false,
		},
		{
			name:          "selector is empty",
			pod:           NewPod(PodArg{Name: "t1", Labels: map[string]string{"app": "tidb"}}),
			selector:      v1alpha1.PodSelectorSpec{},
			expectedValue: true,
		},
		{
			name: "meet labels and meet expressions",
			pod:  NewPod(PodArg{Name: "t1", Status: v1.PodPending, Labels: map[string]string{"app": "tikv", "ss": "t1"}}),
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"app": "tikv"},
					ExpressionSelectors: []metav1.LabelSelectorRequirement{
						{
							Key:      "ss",
							Operator: metav1.LabelSelectorOpExists,
						},
					},
				},
			},
			expectedValue: true,
		},
		{
			name: "meet labels and not meet expressions",
			pod:  NewPod(PodArg{Name: "t1", Status: v1.PodPending, Labels: map[string]string{"app": "tikv", "ss": "t1"}}),
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"app": "tikv"},
					ExpressionSelectors: []metav1.LabelSelectorRequirement{
						{
							Key:      "ss",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"t1"},
						},
					},
				},
			},
			expectedValue: false,
		},
		{
			name: "meet namespace",
			pod:  NewPod(PodArg{Name: "t1"}),
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces: []string{metav1.NamespaceDefault},
				},
			},
			expectedValue: true,
		},
		{
			name: "meet namespace and meet labels",
			pod:  NewPod(PodArg{Name: "t1", Labels: map[string]string{"app": "tikv"}}),
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces:     []string{metav1.NamespaceDefault},
					LabelSelectors: map[string]string{"app": "tikv"},
				},
			},
			expectedValue: true,
		},
		{
			name: "meet namespace and not meet labels",
			pod:  NewPod(PodArg{Name: "t1", Labels: map[string]string{"app": "tidb"}}),
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces:     []string{metav1.NamespaceDefault},
					LabelSelectors: map[string]string{"app": "tikv"},
				},
			},
			expectedValue: false,
		},
		{
			name: "meet pods",
			pod:  NewPod(PodArg{Name: "t1", Labels: map[string]string{"app": "tidb"}}),
			selector: v1alpha1.PodSelectorSpec{
				Pods: map[string][]string{
					metav1.NamespaceDefault: {"t1"},
				},
			},
			expectedValue: true,
		},
		{
			name: "meet annotation",
			pod:  NewPod(PodArg{Name: "t1", Ans: map[string]string{"an": "n1", "an2": "n2"}, Labels: map[string]string{"app": "tidb"}}),
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces: []string{metav1.NamespaceDefault},
					AnnotationSelectors: map[string]string{
						"an": "n1",
					},
				},
			},
			expectedValue: true,
		},
		{
			name: "not meet annotation",
			pod:  NewPod(PodArg{Name: "t1", Ans: map[string]string{"an": "n1"}, Labels: map[string]string{"app": "tidb"}}),
			selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces: []string{metav1.NamespaceDefault},
					AnnotationSelectors: map[string]string{
						"an": "n2",
					},
				},
			},
			expectedValue: false,
		},
		{
			name: "meet pod selector",
			pod:  NewPod(PodArg{Name: "t1", Labels: map[string]string{"app": "tidb"}}),
			selector: v1alpha1.PodSelectorSpec{
				Pods: map[string][]string{
					metav1.NamespaceDefault: {"t1", "t2"},
				},
			},
			expectedValue: true,
		},
		{
			name: "not meet pod selector",
			pod:  NewPod(PodArg{Name: "t1", Labels: map[string]string{"app": "tidb"}}),
			selector: v1alpha1.PodSelectorSpec{
				Pods: map[string][]string{
					metav1.NamespaceDefault: {"t2"},
				},
			},
			expectedValue: false,
		},
		{
			name: "meet pod selector and not meet labels",
			pod:  NewPod(PodArg{Name: "t1", Labels: map[string]string{"app": "tidb"}}),
			selector: v1alpha1.PodSelectorSpec{
				Pods: map[string][]string{
					metav1.NamespaceDefault: {"t1", "t2"},
				},
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"app": "tikv"},
				},
			},
			expectedValue: false,
		},
	}

	for _, tc := range tcs {
		meet, err := CheckPodMeetSelector(tc.pod, tc.selector)
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(meet).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestRandomFixedIndexes(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name              string
		start             uint
		end               uint
		count             uint
		expectedOutputLen int
	}

	tcs := []TestCase{
		{
			name:              "start 0, end 10, count 3",
			start:             0,
			end:               10,
			count:             3,
			expectedOutputLen: 3,
		},
		{
			name:              "start 0, end 10, count 12",
			start:             0,
			end:               10,
			count:             12,
			expectedOutputLen: 10,
		},
		{
			name:              "start 5, end 10, count 3",
			start:             5,
			end:               10,
			count:             3,
			expectedOutputLen: 3,
		},
	}

	for _, tc := range tcs {
		values := RandomFixedIndexes(tc.start, tc.end, tc.count)
		g.Expect(len(values)).To(Equal(tc.expectedOutputLen), tc.name)

		for _, v := range values {
			g.Expect(v).Should(BeNumerically(">=", tc.start), tc.name)
			g.Expect(v).Should(BeNumerically("<", tc.end), tc.name)
		}
	}
}
