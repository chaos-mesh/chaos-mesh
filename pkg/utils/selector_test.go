// Copyright 2019 PingCAP, Inc.
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

package utils

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/pkg/label"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSelectPods(t *testing.T) {
	g := NewGomegaWithT(t)

	objects, pods := generateNPods("p", 5, v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"l1": "l1"})

	objects2, pods2 := generateNPods("s", 2, v1.PodRunning, "test-s", nil, map[string]string{"l2": "l2"})

	objects = append(objects, objects2...)
	pods = append(pods, pods2...)

	c := fake.NewFakeClient(objects...)

	type TestCase struct {
		name         string
		selector     v1alpha1.SelectorSpec
		expectedPods []v1.Pod
	}

	tcs := []TestCase{
		{
			name: "filter specified pods",
			selector: v1alpha1.SelectorSpec{
				Pods: map[string][]string{
					metav1.NamespaceDefault: {"p3", "p4"},
					"test-s":                {"s1"},
				},
			},
			expectedPods: []v1.Pod{pods[3], pods[4], pods[6]},
		},
		{
			name: "filter labels pods",
			selector: v1alpha1.SelectorSpec{
				LabelSelectors: map[string]string{"l2": "l2"},
			},
			expectedPods: []v1.Pod{pods[5], pods[6]},
		},
		{
			name: "filter namespace and labels",
			selector: v1alpha1.SelectorSpec{
				Namespaces:     []string{"test-s"},
				LabelSelectors: map[string]string{"l2": "l2"},
			},
			expectedPods: []v1.Pod{pods[5], pods[6]},
		},
		{
			name: "filter namespace and labels",
			selector: v1alpha1.SelectorSpec{
				Namespaces:     []string{metav1.NamespaceDefault},
				LabelSelectors: map[string]string{"l2": "l2"},
			},
			expectedPods: nil,
		},
	}

	for _, tc := range tcs {
		filteredPods, err := selectPods(context.Background(), c, tc.selector)
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(len(filteredPods)).To(Equal(len(tc.expectedPods)), tc.name)
	}
}

func TestCheckPodMeetSelector(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		selector      v1alpha1.SelectorSpec
		pod           v1.Pod
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name: "meet label",
			pod:  newPod("t1", v1.PodPending, metav1.NamespaceDefault, nil, map[string]string{"app": "tikv", "ss": "t1"}),
			selector: v1alpha1.SelectorSpec{
				LabelSelectors: map[string]string{"app": "tikv"},
			},
			expectedValue: true,
		},
		{
			name: "not meet label",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"app": "tidb", "ss": "t1"}),
			selector: v1alpha1.SelectorSpec{
				LabelSelectors: map[string]string{"app": "tikv"},
			},
			expectedValue: false,
		},
		{
			name: "pod labels is empty",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, nil, nil),
			selector: v1alpha1.SelectorSpec{
				LabelSelectors: map[string]string{"app": "tikv"},
			},
			expectedValue: false,
		},
		{
			name:          "selector is empty",
			pod:           newPod("t1", v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"app": "tidb"}),
			selector:      v1alpha1.SelectorSpec{},
			expectedValue: true,
		},
		{
			name: "meet namespace",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, nil, nil),
			selector: v1alpha1.SelectorSpec{
				Namespaces: []string{metav1.NamespaceDefault},
			},
			expectedValue: true,
		},
		{
			name: "meet namespace and meet labels",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"app": "tikv"}),
			selector: v1alpha1.SelectorSpec{
				Namespaces:     []string{metav1.NamespaceDefault},
				LabelSelectors: map[string]string{"app": "tikv"},
			},
			expectedValue: true,
		},
		{
			name: "meet namespace and not meet labels",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"app": "tidb"}),
			selector: v1alpha1.SelectorSpec{
				Namespaces:     []string{metav1.NamespaceDefault},
				LabelSelectors: map[string]string{"app": "tikv"},
			},
			expectedValue: false,
		},
		{
			name: "meet pods",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"app": "tidb"}),
			selector: v1alpha1.SelectorSpec{
				Pods: map[string][]string{
					metav1.NamespaceDefault: {"t1"},
				},
			},
			expectedValue: true,
		},
		{
			name: "meet annotation",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, map[string]string{"an": "n1", "an2": "n2"}, map[string]string{"app": "tidb"}),
			selector: v1alpha1.SelectorSpec{
				Namespaces: []string{metav1.NamespaceDefault},
				AnnotationSelectors: map[string]string{
					"an": "n1",
				},
			},
			expectedValue: true,
		},
		{
			name: "not meet annotation",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, map[string]string{"an": "n1"}, map[string]string{"app": "tidb"}),
			selector: v1alpha1.SelectorSpec{
				Namespaces: []string{metav1.NamespaceDefault},
				AnnotationSelectors: map[string]string{
					"an": "n2",
				},
			},
			expectedValue: false,
		},
		{
			name: "meet pod selector",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"app": "tidb"}),
			selector: v1alpha1.SelectorSpec{
				Pods: map[string][]string{
					metav1.NamespaceDefault: {"t1", "t2"},
				},
			},
			expectedValue: true,
		},
		{
			name: "not meet pod selector",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"app": "tidb"}),
			selector: v1alpha1.SelectorSpec{
				Pods: map[string][]string{
					metav1.NamespaceDefault: {"t2"},
				},
			},
			expectedValue: false,
		},
		{
			name: "meet pod selector and not meet labels",
			pod:  newPod("t1", v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"app": "tidb"}),
			selector: v1alpha1.SelectorSpec{
				Pods: map[string][]string{
					metav1.NamespaceDefault: {"t1", "t2"},
				},
				LabelSelectors: map[string]string{"app": "tikv"},
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

func TestFilterByPhaseSelector(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name           string
		pods           []v1.Pod
		filterSelector labels.Selector
		filteredPods   []v1.Pod
	}

	pods := []v1.Pod{
		newPod("p1", v1.PodRunning, metav1.NamespaceDefault, nil, nil),
		newPod("p2", v1.PodRunning, metav1.NamespaceDefault, nil, nil),
		newPod("p3", v1.PodPending, metav1.NamespaceDefault, nil, nil),
		newPod("p4", v1.PodFailed, metav1.NamespaceDefault, nil, nil),
	}

	var tcs []TestCase

	runningSelector, err := parseSelector(string(pods[1].Status.Phase))
	g.Expect(err).ShouldNot(HaveOccurred())

	tcs = append(tcs, TestCase{
		name:           "filter n2",
		pods:           pods,
		filterSelector: runningSelector,
		filteredPods:   []v1.Pod{pods[0], pods[1]},
	})

	emptySelector, err := parseSelector("")
	g.Expect(err).ShouldNot(HaveOccurred())
	tcs = append(tcs, TestCase{
		name:           "filter empty selector",
		pods:           pods,
		filterSelector: emptySelector,
		filteredPods:   pods,
	})

	tcs = append(tcs, TestCase{
		name:           "filter no pods",
		pods:           []v1.Pod{},
		filterSelector: runningSelector,
		filteredPods:   nil,
	})

	runningAndPendingSelector, err := parseSelector("Running,Pending")
	g.Expect(err).ShouldNot(HaveOccurred())

	tcs = append(tcs, TestCase{
		name:           "filter running and pending",
		pods:           pods,
		filterSelector: runningAndPendingSelector,
		filteredPods:   []v1.Pod{pods[0], pods[1], pods[2]},
	})

	failedSelector, err := parseSelector("Failed")
	g.Expect(err).ShouldNot(HaveOccurred())

	tcs = append(tcs, TestCase{
		name:           "filter failed",
		pods:           pods,
		filterSelector: failedSelector,
		filteredPods:   []v1.Pod{pods[3]},
	})

	unknownSelector, err := parseSelector("Unknown")
	g.Expect(err).ShouldNot(HaveOccurred())
	tcs = append(tcs, TestCase{
		name:           "filter Unknown",
		pods:           pods,
		filterSelector: unknownSelector,
		filteredPods:   nil,
	})

	for _, tc := range tcs {
		g.Expect(filterByPhaseSelector(tc.pods, tc.filterSelector)).To(Equal(tc.filteredPods), tc.name)
	}
}

func TestFilterByAnnotations(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name           string
		pods           []v1.Pod
		filterSelector labels.Selector
		filteredPods   []v1.Pod
	}

	pods := []v1.Pod{
		newPod("p1", v1.PodRunning, metav1.NamespaceDefault, map[string]string{"p1": "p1"}, nil),
		newPod("p2", v1.PodRunning, metav1.NamespaceDefault, map[string]string{"p2": "p2"}, nil),
		newPod("p3", v1.PodRunning, metav1.NamespaceDefault, map[string]string{"t": "t"}, nil),
		newPod("p4", v1.PodRunning, metav1.NamespaceDefault, map[string]string{"t": "t"}, nil),
	}

	var tcs []TestCase
	p2Selector, err := parseSelector(label.Label(pods[1].Annotations).String())
	g.Expect(err).ShouldNot(HaveOccurred())

	tcs = append(tcs, TestCase{
		name:           "filter p2",
		pods:           pods,
		filterSelector: p2Selector,
		filteredPods:   []v1.Pod{pods[1]},
	})

	emptySelector, err := parseSelector(label.Label(map[string]string{}).String())
	g.Expect(err).ShouldNot(HaveOccurred())
	tcs = append(tcs, TestCase{
		name:           "filter empty selector",
		pods:           pods,
		filterSelector: emptySelector,
		filteredPods:   pods,
	})

	tcs = append(tcs, TestCase{
		name:           "filter no pods",
		pods:           []v1.Pod{},
		filterSelector: p2Selector,
		filteredPods:   nil,
	})

	for _, tc := range tcs {
		g.Expect(filterByAnnotations(tc.pods, tc.filterSelector)).To(Equal(tc.filteredPods), tc.name)
	}
}

func TestIsAllowedNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)
	type TestCase struct {
		name   string
		pods   []v1.Pod
		ret    []bool
		allow  []string
		ignore []string
	}
	pods := []v1.Pod{
		newPod("p1", v1.PodRunning, "allow", nil, nil),
		newPod("p1", v1.PodRunning, "allow-app", nil, nil),
		newPod("p1", v1.PodRunning, "app-allow", nil, nil),
		newPod("p1", v1.PodRunning, "ignore", nil, nil),
		newPod("p1", v1.PodRunning, "ignore-app", nil, nil),
		newPod("p1", v1.PodRunning, "app-ignore", nil, nil),
	}

	allowRet := []bool{true, true, true, false, false, false}

	var tcs []TestCase
	tcs = append(tcs, TestCase{
		name:  "only set allow",
		pods:  pods,
		ret:   allowRet,
		allow: []string{"allow"},
	})

	tcs = append(tcs, TestCase{
		name:   "only set ignore",
		pods:   pods,
		ret:    allowRet,
		ignore: []string{"ignore"},
	})

	tcs = append(tcs, TestCase{
		name:   "only set allow",
		pods:   pods,
		ret:    allowRet,
		allow:  []string{"allow"},
		ignore: []string{"ignore"},
	})

	setRule := func(allow []string, ignore []string) {
		AllowedNamespaces = allow
		IgnoredNamespaces = ignore
	}

	clean := func() {
		AllowedNamespaces = []string{}
		IgnoredNamespaces = []string{}
	}

	for _, tc := range tcs {
		setRule(tc.allow, tc.ignore)
		for index, pod := range tc.pods {
			g.Expect(IsAllowedNamespaces(pod.Namespace)).Should(Equal(tc.ret[index]))
		}
		clean()
	}
}

func TestFilterNamespaceSelector(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name           string
		pods           []v1.Pod
		filterSelector labels.Selector
		filteredPods   []v1.Pod
	}

	pods := []v1.Pod{
		newPod("p1", v1.PodRunning, "n1", nil, nil),
		newPod("p2", v1.PodRunning, "n2", nil, nil),
		newPod("p3", v1.PodRunning, "n2", nil, nil),
		newPod("p4", v1.PodRunning, "n4", nil, nil),
	}

	var tcs []TestCase
	n2Selector, err := parseSelector(pods[1].Namespace)
	g.Expect(err).ShouldNot(HaveOccurred())

	tcs = append(tcs, TestCase{
		name:           "filter n2",
		pods:           pods,
		filterSelector: n2Selector,
		filteredPods:   []v1.Pod{pods[1], pods[2]},
	})

	emptySelector, err := parseSelector("")
	g.Expect(err).ShouldNot(HaveOccurred())
	tcs = append(tcs, TestCase{
		name:           "filter empty selector",
		pods:           pods,
		filterSelector: emptySelector,
		filteredPods:   pods,
	})

	tcs = append(tcs, TestCase{
		name:           "filter no pods",
		pods:           []v1.Pod{},
		filterSelector: n2Selector,
		filteredPods:   nil,
	})

	n2AndN3Selector, err := parseSelector("n2,n3")
	g.Expect(err).ShouldNot(HaveOccurred())

	tcs = append(tcs, TestCase{
		name:           "filter n2 and n3",
		pods:           pods,
		filterSelector: n2AndN3Selector,
		filteredPods:   []v1.Pod{pods[1], pods[2]},
	})

	n2AndN4Selector, err := parseSelector("n2,n4")
	g.Expect(err).ShouldNot(HaveOccurred())

	tcs = append(tcs, TestCase{
		name:           "filter n2 and n4",
		pods:           pods,
		filterSelector: n2AndN4Selector,
		filteredPods:   []v1.Pod{pods[1], pods[2], pods[3]},
	})

	for _, tc := range tcs {
		g.Expect(filterByNamespaceSelector(tc.pods, tc.filterSelector)).To(Equal(tc.filteredPods), tc.name)
	}
}

func newPod(
	name string,
	status v1.PodPhase,
	namespace string,
	ans map[string]string,
	ls map[string]string,
) v1.Pod {
	return v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      ls,
			Annotations: ans,
		},
		Status: v1.PodStatus{
			Phase: status,
		},
	}
}

func generateNPods(
	namePrefix string,
	n int,
	status v1.PodPhase,
	ns string,
	ans map[string]string,
	ls map[string]string,
) ([]runtime.Object, []v1.Pod) {
	var podObjects []runtime.Object
	var pods []v1.Pod
	for i := 0; i < n; i++ {
		pod := newPod(fmt.Sprintf("%s%d", namePrefix, i), status, ns, ans, ls)
		podObjects = append(podObjects, &pod)
		pods = append(pods, pod)
	}

	return podObjects, pods
}
