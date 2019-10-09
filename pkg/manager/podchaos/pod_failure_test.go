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

package podchaos

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned/fake"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func newPodFailureJob(pc *v1alpha1.PodChaos, objects ...runtime.Object) PodFailureJob {
	kubeCli := kubefake.NewSimpleClientset(objects...)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeCli, 0)

	podLister := kubeInformerFactory.Core().V1().Pods().Lister()

	return PodFailureJob{
		podChaos:  pc,
		kubeCli:   kubeCli,
		podLister: podLister,
	}
}

func newPodFailurePodChaos(name string) *v1alpha1.PodChaos {
	return &v1alpha1.PodChaos{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodChaos",
			APIVersion: "pingcap.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector: v1alpha1.SelectorSpec{
				Namespaces: []string{"chaos-testing"},
			},
			Scheduler: v1alpha1.SchedulerSpec{
				Cron: "@every 1m",
			},
			Action:   v1alpha1.PodFailureAction,
			Duration: "30s",
		},
	}
}

func newPodChaosDiffSelector(name string, selector v1alpha1.SelectorSpec) *v1alpha1.PodChaos {
	return &v1alpha1.PodChaos{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodChaos",
			APIVersion: "pingcap.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector: selector,
			Scheduler: v1alpha1.SchedulerSpec{
				Cron: "@every 1m",
			},
			Action:   v1alpha1.PodFailureAction,
			Duration: "30s",
		},
	}
}

func newPodChaosDiffScheduler(name string, scheduler v1alpha1.SchedulerSpec) *v1alpha1.PodChaos {
	return &v1alpha1.PodChaos{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodChaos",
			APIVersion: "pingcap.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector: v1alpha1.SelectorSpec{
				Namespaces: []string{"chaos-testing"},
			},
			Scheduler: scheduler,
			Action:    v1alpha1.PodFailureAction,
			Duration:  "30s",
		},
	}
}

func newPodChaosDiffDuration(name string, duration string) *v1alpha1.PodChaos {
	return &v1alpha1.PodChaos{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodChaos",
			APIVersion: "pingcap.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector: v1alpha1.SelectorSpec{
				Namespaces: []string{"chaos-testing"},
			},
			Scheduler: v1alpha1.SchedulerSpec{
				Cron: "@every 1m",
			},
			Duration: duration,
			Action:   v1alpha1.PodFailureAction,
		},
	}
}

func TestPodFailureJobEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		job1PodChaos  *v1alpha1.PodChaos
		job2PodChaos  *v1alpha1.PodChaos
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name:          "same podChaos",
			job1PodChaos:  newPodFailurePodChaos("job1"),
			job2PodChaos:  newPodFailurePodChaos("job1"),
			expectedValue: true,
		},
		{
			name:          "diff name",
			job1PodChaos:  newPodFailurePodChaos("job1"),
			job2PodChaos:  newPodFailurePodChaos("job2"),
			expectedValue: false,
		},
		{
			name:          "diff selector",
			job1PodChaos:  newPodChaosDiffSelector("job", v1alpha1.SelectorSpec{Namespaces: []string{"p1"}}),
			job2PodChaos:  newPodChaosDiffSelector("job", v1alpha1.SelectorSpec{Namespaces: []string{"p2"}}),
			expectedValue: false,
		},
		{
			name:          "diff scheduler",
			job1PodChaos:  newPodChaosDiffScheduler("job", v1alpha1.SchedulerSpec{Cron: "@every 1m"}),
			job2PodChaos:  newPodChaosDiffScheduler("job", v1alpha1.SchedulerSpec{Cron: "@every 2m"}),
			expectedValue: false,
		},
		{
			name:          "diff duration",
			job1PodChaos:  newPodChaosDiffDuration("job1", "1m"),
			job2PodChaos:  newPodChaosDiffDuration("job2", "2m"),
			expectedValue: false,
		},
	}

	for _, tc := range tcs {
		job1 := newPodFailureJob(tc.job1PodChaos)
		job2 := newPodFailureJob(tc.job2PodChaos)
		g.Expect(job1.Equal(&job2)).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestPodFailureFailPod(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name           string
		pod            v1.Pod
		expectedPod    v1.Pod
		expectedResult resultF
	}

	tcs := []TestCase{
		{
			name: "one container",
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "t1",
					Namespace:   NAMESPACE,
					Annotations: make(map[string]string),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: "pingcap.com/image1", Name: "t1"}},
				},
			},
			expectedPod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "t1",
					Namespace:   NAMESPACE,
					Annotations: map[string]string{GenAnnotationKeyForImage(newPodFailurePodChaos("test"), "t1"): "pingcap.com/image1"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: fakeImage, Name: "t1"}},
				},
			},
			expectedResult: Succeed,
		},
		{
			name: "two containers",
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "t1",
					Namespace:   NAMESPACE,
					Annotations: make(map[string]string),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: "pingcap.com/image1", Name: "t1"}, {Image: "pingcap.com/image2", Name: "t2"}},
				},
			},
			expectedPod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "t1",
					Namespace: NAMESPACE,
					Annotations: map[string]string{
						GenAnnotationKeyForImage(newPodFailurePodChaos("test"), "t1"): "pingcap.com/image1",
						GenAnnotationKeyForImage(newPodFailurePodChaos("test"), "t2"): "pingcap.com/image2",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: fakeImage, Name: "t1"}, {Image: fakeImage, Name: "t2"}},
				},
			},
			expectedResult: Succeed,
		},
		{
			name: "annotation already exist",
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "t1",
					Namespace:   NAMESPACE,
					Annotations: map[string]string{GenAnnotationKeyForImage(newPodFailurePodChaos("test"), "t1"): "pingcap.com/image1"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: fakeImage, Name: "t1"}},
				},
			},
			expectedPod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "t1",
					Namespace:   NAMESPACE,
					Annotations: map[string]string{GenAnnotationKeyForImage(newPodFailurePodChaos("test"), "t1"): "pingcap.com/image1"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: fakeImage, Name: "t1"}},
				},
			},
			expectedResult: HaveOccurred,
		},
	}

	for _, tc := range tcs {
		job := newPodFailureJob(newPodFailurePodChaos("test"), &tc.pod)
		g.Expect(job.failPod(tc.pod)).Should(tc.expectedResult(), tc.name)
		pod, err := job.kubeCli.CoreV1().Pods(tc.pod.Namespace).Get(tc.pod.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(pod.Annotations).To(Equal(tc.expectedPod.Annotations), tc.name)
		g.Expect(pod.Spec.Containers).To(Equal(tc.expectedPod.Spec.Containers), tc.name)
	}
}

func TestPodFailureJobRecoverPod(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name           string
		pod            v1.Pod
		expectedPod    v1.Pod
		expectedResult resultF
	}

	tcs := []TestCase{
		{
			name: "one container",
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "t1",
					Namespace:   NAMESPACE,
					Annotations: map[string]string{GenAnnotationKeyForImage(newPodFailurePodChaos("test"), "t1"): "pingcap.com/image1"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: fakeImage, Name: "t1"}},
				},
			},
			expectedPod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "t1",
					Namespace:   NAMESPACE,
					Annotations: make(map[string]string),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: "pingcap.com/image1", Name: "t1"}},
				},
			},
			expectedResult: Succeed,
		},
		{
			name: "two containers",
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "t1",
					Namespace: NAMESPACE,
					Annotations: map[string]string{
						GenAnnotationKeyForImage(newPodFailurePodChaos("test"), "t1"): "pingcap.com/image1",
						GenAnnotationKeyForImage(newPodFailurePodChaos("test"), "t2"): "pingcap.com/image2",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: fakeImage, Name: "t1"}, {Image: fakeImage, Name: "t2"}},
				},
			},
			expectedPod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "t1",
					Namespace:   NAMESPACE,
					Annotations: make(map[string]string),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: "pingcap.com/image1", Name: "t1"}, {Image: "pingcap.com/image2", Name: "t2"}},
				},
			},
			expectedResult: Succeed,
		},
		{
			name: "annotation already exist",
			pod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "t1",
					Namespace:   NAMESPACE,
					Annotations: make(map[string]string),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: fakeImage, Name: "t1"}},
				},
			},
			expectedPod: v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        "t1",
					Namespace:   NAMESPACE,
					Annotations: make(map[string]string),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Image: fakeImage, Name: "t1"}},
				},
			},
			expectedResult: Succeed,
		},
	}

	for _, tc := range tcs {
		job := newPodFailureJob(newPodFailurePodChaos("test"), &tc.pod)
		g.Expect(job.recoverPod(tc.pod)).Should(tc.expectedResult(), tc.name)
		pod, err := job.kubeCli.CoreV1().Pods(tc.pod.Namespace).Get(tc.pod.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(pod.Annotations).To(Equal(tc.expectedPod.Annotations), tc.name)
		g.Expect(pod.Spec.Containers).To(Equal(tc.expectedPod.Spec.Containers), tc.name)
	}
}

func TestPodFailureJobAddFinalizer(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name               string
		pod                v1.Pod
		podChaos           *v1alpha1.PodChaos
		expectedFinalizers []string
	}

	pc := newPodFailurePodChaos("t2")
	pc.Finalizers = []string{"default/t1"}

	tcs := []TestCase{
		{
			name:               "one finalizer",
			pod:                newPod("t1", v1.PodRunning),
			podChaos:           newPodFailurePodChaos("t1"),
			expectedFinalizers: []string{fmt.Sprintf("%s/t1", NAMESPACE)},
		},
		{
			name:               "two finalizers",
			pod:                newPod("t2", v1.PodRunning),
			podChaos:           pc,
			expectedFinalizers: []string{"default/t1", fmt.Sprintf("%s/t2", NAMESPACE)},
		},
	}

	for _, tc := range tcs {
		job := newPodFailureJob(tc.podChaos, &tc.pod)
		job.cli = fake.NewSimpleClientset(tc.podChaos)
		g.Expect(job.addPodFinalizer(tc.pod)).ShouldNot(HaveOccurred(), tc.name)

		tpc, err := job.cli.PingcapV1alpha1().PodChaoses(tc.podChaos.Namespace).Get(tc.podChaos.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(tpc.Finalizers).To(Equal(tc.expectedFinalizers), tc.name)
	}
}
