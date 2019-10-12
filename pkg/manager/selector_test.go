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

package manager

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/label"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

var (
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

func TestSelectPods(t *testing.T) {
	g := NewGomegaWithT(t)

	objects, pods := generateNPods("p", 5, v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"l1": "l1"})

	objects2, pods2 := generateNPods("s", 2, v1.PodRunning, "test-s", nil, map[string]string{"l2": "l2"})

	objects = append(objects, objects2...)
	pods = append(pods, pods2...)

	kubeCli := kubefake.NewSimpleClientset(objects...)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeCli, noResyncPeriodFunc())

	for _, pod := range pods {
		pod := pod
		err := kubeInformerFactory.Core().V1().Pods().Informer().GetIndexer().Add(&pod)
		g.Expect(err).ShouldNot(HaveOccurred())
	}

	podLister := kubeInformerFactory.Core().V1().Pods().Lister()

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
		filteredPods, err := SelectPods(tc.selector, podLister, kubeCli)
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		var fPods []v1.Pod
		fPods = append(fPods, filteredPods...)
		g.Expect(fPods).To(Equal(tc.expectedPods), tc.name)
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

func TestFilterByPhase(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name         string
		pods         []v1.Pod
		filterPhase  v1.PodPhase
		filteredPods []v1.Pod
	}

	pods := []v1.Pod{
		newPod("p1", v1.PodRunning, metav1.NamespaceDefault, nil, nil),
		newPod("p2", v1.PodRunning, metav1.NamespaceDefault, nil, nil),
		newPod("p3", v1.PodPending, metav1.NamespaceDefault, nil, nil),
		newPod("p4", v1.PodFailed, metav1.NamespaceDefault, nil, nil),
	}

	tcs := []TestCase{
		{
			name:         "filter running pods",
			pods:         pods,
			filterPhase:  v1.PodRunning,
			filteredPods: []v1.Pod{pods[0], pods[1]},
		},
		{
			name:         "filter pending pods",
			pods:         pods,
			filterPhase:  v1.PodPending,
			filteredPods: []v1.Pod{pods[2]},
		},
		{
			name:         "no pods",
			pods:         []v1.Pod{},
			filterPhase:  v1.PodPending,
			filteredPods: nil,
		},
		{
			name:         "filter unknown pods",
			pods:         pods,
			filterPhase:  v1.PodUnknown,
			filteredPods: nil,
		},
	}

	for _, tc := range tcs {
		g.Expect(filterByPhase(tc.pods, tc.filterPhase)).To(Equal(tc.filteredPods), tc.name)
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

func TestFilterNamespace(t *testing.T) {
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
		g.Expect(filterByNamespaces(tc.pods, tc.filterSelector)).To(Equal(tc.filteredPods), tc.name)
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
