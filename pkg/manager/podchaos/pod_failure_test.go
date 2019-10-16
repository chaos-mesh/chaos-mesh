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

func newPodChaosWithFinalizers(name string, finalizers []string) *v1alpha1.PodChaos {
	return &v1alpha1.PodChaos{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodChaos",
			APIVersion: "pingcap.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  metav1.NamespaceDefault,
			Finalizers: finalizers,
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
		name string
		pod  v1.Pod
		// expectedPod    v1.Pod
		expectedResult       resultF
		expectedGetPodResult resultF
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
			expectedResult:       Succeed,
			expectedGetPodResult: HaveOccurred,
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
			expectedResult:       Succeed,
			expectedGetPodResult: HaveOccurred,
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
			expectedResult:       Succeed,
			expectedGetPodResult: HaveOccurred,
		},
	}

	for _, tc := range tcs {
		job := newPodFailureJob(newPodFailurePodChaos("test"), &tc.pod)
		g.Expect(job.recoverPod(tc.pod)).Should(tc.expectedResult(), tc.name)
		_, err := job.kubeCli.CoreV1().Pods(tc.pod.Namespace).Get(tc.pod.Name, metav1.GetOptions{})
		g.Expect(err).Should(tc.expectedGetPodResult(), tc.name)
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
			podChaos:           newPodChaosWithFinalizers("t2", []string{"default/t1"}),
			expectedFinalizers: []string{"default/t1", fmt.Sprintf("%s/t2", NAMESPACE)},
		},
	}

	for _, tc := range tcs {
		job := newPodFailureJob(tc.podChaos, &tc.pod)
		job.cli = fake.NewSimpleClientset(tc.podChaos)
		g.Expect(job.addPodFinalizer(tc.pod)).ShouldNot(HaveOccurred(), tc.name)

		tpc, err := job.cli.PingcapV1alpha1().PodChaoses(tc.podChaos.Namespace).
			Get(tc.podChaos.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(tpc.Finalizers).To(Equal(tc.expectedFinalizers), tc.name)
	}
}

func TestPodFailureJobFailFixedPods(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name                  string
		fixedValue            string
		Duration              string
		podsCount             int
		record                *v1alpha1.PodChaosExperimentStatus
		expectedFinalizersLen int
		expectedResultF       resultF
	}

	tcs := []TestCase{
		{
			name:                  "fixed 2, pods 5",
			fixedValue:            "2",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 2,
			expectedResultF:       Succeed,
		},
		{
			name:                  "fixed 3, pods 2",
			fixedValue:            "3",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             2,
			expectedFinalizersLen: 2,
			expectedResultF:       Succeed,
		},
		{
			name:                  "invalid duration",
			fixedValue:            "3",
			Duration:              "1",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             2,
			expectedFinalizersLen: 0,
			expectedResultF:       HaveOccurred,
		},
	}

	for _, tc := range tcs {
		pc := newPodFailurePodChaos("pc-test")
		pc.Spec.Value = tc.fixedValue
		pc.Spec.Duration = tc.Duration
		objects, pods := generateNRunningPods("pc-test", tc.podsCount)
		job := newPodFailureJob(pc, objects...)
		job.cli = fake.NewSimpleClientset(pc)
		g.Expect(job.failFixedPods(pods, tc.record)).Should(tc.expectedResultF(), tc.name)
		tpc, err := job.cli.PingcapV1alpha1().PodChaoses(pc.Namespace).Get(pc.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(len(tpc.Finalizers)).To(Equal(tc.expectedFinalizersLen), tc.name)
		g.Expect(len(tc.record.Pods)).To(Equal(tc.expectedFinalizersLen), tc.name)
	}
}

func TestPodFailureJobFailFixedPercentagePods(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name                  string
		fixedValue            string
		Duration              string
		record                *v1alpha1.PodChaosExperimentStatus
		podsCount             int
		expectedFinalizersLen int
		expectedResultF       resultF
	}

	tcs := []TestCase{
		{
			name:                  "fixed 0%%, pods 5",
			fixedValue:            "0",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 0,
			expectedResultF:       Succeed,
		},
		{
			name:                  "fixed 100%%, pods 5",
			fixedValue:            "100",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 5,
			expectedResultF:       Succeed,
		},
		{
			name:                  "fixed 100%%, pods 0",
			fixedValue:            "100",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             0,
			expectedFinalizersLen: 0,
			expectedResultF:       Succeed,
		},
		{
			name:                  "fixed 50%%, pods 5",
			fixedValue:            "50",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 2,
			expectedResultF:       Succeed,
		},
		{
			name:                  "fixed 200%%, pods 5",
			fixedValue:            "200",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 0,
			expectedResultF:       HaveOccurred,
		},
		{
			name:                  "fixed -100%%, pods 5",
			fixedValue:            "-100",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 0,
			expectedResultF:       HaveOccurred,
		},
		{
			name:                  "invalid duration",
			fixedValue:            "3",
			Duration:              "1",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             2,
			expectedFinalizersLen: 0,
			expectedResultF:       HaveOccurred,
		},
	}

	for _, tc := range tcs {
		pc := newPodFailurePodChaos("pc-test")
		pc.Spec.Value = tc.fixedValue
		pc.Spec.Duration = tc.Duration
		objects, pods := generateNRunningPods("pc-test", tc.podsCount)
		job := newPodFailureJob(pc, objects...)
		job.cli = fake.NewSimpleClientset(pc)
		g.Expect(job.failFixedPercentagePods(pods, tc.record)).Should(tc.expectedResultF(), tc.name)
		tpc, err := job.cli.PingcapV1alpha1().PodChaoses(pc.Namespace).Get(pc.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(len(tpc.Finalizers)).To(Equal(tc.expectedFinalizersLen), tc.name)
		g.Expect(len(tc.record.Pods)).To(Equal(tc.expectedFinalizersLen), tc.name)
	}
}

func TestPodFailureJobFailMaxPercentagePods(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name                  string
		fixedValue            string
		Duration              string
		record                *v1alpha1.PodChaosExperimentStatus
		podsCount             int
		expectedFinalizersLen int
		expectedResultF       resultF
	}

	tcs := []TestCase{
		{
			name:                  "fixed max 0%%, pods 5",
			fixedValue:            "0",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 0,
			expectedResultF:       Succeed,
		},
		{
			name:                  "fixed max 100%%, pods 5",
			fixedValue:            "100",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 5,
			expectedResultF:       Succeed,
		},
		{
			name:                  "fixed max 100%%, pods 0",
			fixedValue:            "100",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             0,
			expectedFinalizersLen: 0,
			expectedResultF:       Succeed,
		},
		{
			name:                  "fixed max 50%%, pods 5",
			fixedValue:            "50",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 2,
			expectedResultF:       Succeed,
		},
		{
			name:                  "fixed max 200%%, pods 5",
			fixedValue:            "200",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 0,
			expectedResultF:       HaveOccurred,
		},
		{
			name:                  "fixed max -100%%, pods 5",
			fixedValue:            "-100",
			Duration:              "1ms",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             5,
			expectedFinalizersLen: 0,
			expectedResultF:       HaveOccurred,
		},
		{
			name:                  "invalid duration",
			fixedValue:            "3",
			Duration:              "1",
			record:                &v1alpha1.PodChaosExperimentStatus{},
			podsCount:             2,
			expectedFinalizersLen: 0,
			expectedResultF:       HaveOccurred,
		},
	}

	for _, tc := range tcs {
		pc := newPodFailurePodChaos("pc-test")
		pc.Spec.Value = tc.fixedValue
		pc.Spec.Duration = tc.Duration
		objects, pods := generateNRunningPods("pc-test", tc.podsCount)
		job := newPodFailureJob(pc, objects...)
		job.cli = fake.NewSimpleClientset(pc)
		g.Expect(job.failMaxPercentagePods(pods, tc.record)).Should(tc.expectedResultF(), tc.name)
		tpc, err := job.cli.PingcapV1alpha1().PodChaoses(pc.Namespace).Get(pc.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(len(tpc.Finalizers)).Should(BeNumerically("<=", tc.expectedFinalizersLen), tc.name)
		g.Expect(len(tc.record.Pods)).Should(BeNumerically("<=", tc.expectedFinalizersLen), tc.name)
	}
}

func TestPodFailureJobFailAllPods(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name                  string
		podsCount             int
		expectedFinalizersLen int
		record                *v1alpha1.PodChaosExperimentStatus
		expectedResultF       resultF
		Duration              string
	}

	tcs := []TestCase{
		{
			name:                  "5 pods",
			podsCount:             5,
			record:                &v1alpha1.PodChaosExperimentStatus{},
			expectedFinalizersLen: 5,
			Duration:              "1ms",
			expectedResultF:       Succeed,
		},
		{
			name:                  "0 pods",
			podsCount:             0,
			record:                &v1alpha1.PodChaosExperimentStatus{},
			expectedFinalizersLen: 0,
			Duration:              "1ms",
			expectedResultF:       Succeed,
		},
		{
			name:                  "invalid duration",
			podsCount:             5,
			record:                &v1alpha1.PodChaosExperimentStatus{},
			expectedFinalizersLen: 0,
			Duration:              "1",
			expectedResultF:       HaveOccurred,
		},
	}

	for _, tc := range tcs {
		pc := newPodFailurePodChaos("pc-test")
		pc.Spec.Duration = tc.Duration
		objects, pods := generateNRunningPods("pc-test", tc.podsCount)
		job := newPodFailureJob(pc, objects...)
		job.cli = fake.NewSimpleClientset(pc)
		g.Expect(job.failAllPod(pods, tc.record)).Should(tc.expectedResultF(), tc.name)
		tpc, err := job.cli.PingcapV1alpha1().PodChaoses(pc.Namespace).Get(pc.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(len(tpc.Finalizers)).To(Equal(tc.expectedFinalizersLen), tc.name)
		g.Expect(len(tc.record.Pods)).To(Equal(tc.expectedFinalizersLen), tc.name)
	}
}

func TestPodFailureJobCleanFinalizersAndRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name              string
		pods              []v1.Pod
		podChaos          *v1alpha1.PodChaos
		expectedPodsCount int
	}

	tcs := []TestCase{
		{
			name:              "zero finalizer",
			pods:              newPodsWithFakImageAnnotations("t1", 1),
			podChaos:          newPodChaosWithFinalizers("test", []string{}),
			expectedPodsCount: 1,
		},
		{
			name:              "one finalizer",
			pods:              newPodsWithFakImageAnnotations("t2", 1),
			podChaos:          newPodChaosWithFinalizers("test", []string{"default/t2-0"}),
			expectedPodsCount: 0,
		},
		{
			name:              "two finalizers, 2 pods",
			pods:              newPodsWithFakImageAnnotations("t3", 2),
			podChaos:          newPodChaosWithFinalizers("test", []string{"default/t3-0", "default/t3-1"}),
			expectedPodsCount: 0,
		},
		{
			name:              "two finalizers, 4 pods",
			pods:              newPodsWithFakImageAnnotations("t4", 4),
			podChaos:          newPodChaosWithFinalizers("test", []string{"default/t4-0", "default/t4-1"}),
			expectedPodsCount: 2,
		},
	}

	for _, tc := range tcs {
		var objects []runtime.Object
		for _, pod := range tc.pods {
			pod := pod
			objects = append(objects, &pod)
		}

		job := newPodFailureJob(tc.podChaos, objects...)
		job.cli = fake.NewSimpleClientset(tc.podChaos)
		g.Expect(job.cleanFinalizersAndRecover()).ShouldNot(HaveOccurred(), tc.name)

		tpc, err := job.cli.PingcapV1alpha1().PodChaoses(tc.podChaos.Namespace).Get(tc.podChaos.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(len(tpc.Finalizers)).To(Equal(0), tc.name)
		g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.expectedPodsCount), tc.name)
	}
}

func newPodsWithFakImageAnnotations(prefix string, num int) []v1.Pod {
	var pods []v1.Pod

	for i := 0; i < num; i++ {
		pods = append(pods, v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        fmt.Sprintf("%s-%d", prefix, i),
				Namespace:   NAMESPACE,
				Annotations: map[string]string{GenAnnotationKeyForImage(newPodFailurePodChaos("test"), "t1"): "pingcap.com/image1"},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{{Image: fakeImage, Name: "t1"}},
			},
		})
	}

	return pods
}

func TestPodFailureJobConcurrentFailPods(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name                  string
		podsCount             int
		expectedFinalizersLen int
		failNum               int
		record                *v1alpha1.PodChaosExperimentStatus
		expectedResultF       resultF
		Duration              string
	}

	tcs := []TestCase{
		{
			name:                  "5 pods, fail 5",
			podsCount:             5,
			failNum:               5,
			record:                &v1alpha1.PodChaosExperimentStatus{},
			expectedFinalizersLen: 5,
			Duration:              "1ms",
			expectedResultF:       Succeed,
		},
		{
			name:                  "5 pods, fail 2",
			podsCount:             5,
			failNum:               2,
			record:                &v1alpha1.PodChaosExperimentStatus{},
			expectedFinalizersLen: 2,
			Duration:              "1ms",
			expectedResultF:       Succeed,
		},
		{
			name:                  "5 pods, fail -1",
			podsCount:             5,
			failNum:               -1,
			record:                &v1alpha1.PodChaosExperimentStatus{},
			expectedFinalizersLen: 0,
			Duration:              "1ms",
			expectedResultF:       Succeed,
		},
		{
			name:                  "0 pods, fail 1",
			podsCount:             0,
			failNum:               1,
			record:                &v1alpha1.PodChaosExperimentStatus{},
			expectedFinalizersLen: 0,
			Duration:              "1ms",
			expectedResultF:       Succeed,
		},
	}

	for _, tc := range tcs {
		pc := newPodFailurePodChaos("pc-test")
		pc.Spec.Duration = tc.Duration
		objects, pods := generateNRunningPods("pc-test", tc.podsCount)
		job := newPodFailureJob(pc, objects...)
		job.cli = fake.NewSimpleClientset(pc)
		g.Expect(job.concurrentFailPods(pods, tc.failNum, tc.record)).Should(tc.expectedResultF(), tc.name)
		tpc, err := job.cli.PingcapV1alpha1().PodChaoses(pc.Namespace).Get(pc.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(len(tpc.Finalizers)).To(Equal(tc.expectedFinalizersLen), tc.name)
		g.Expect(len(tc.record.Pods)).To(Equal(tc.expectedFinalizersLen), tc.name)
	}
}

func TestPodFailureJobSetRecordPods(t *testing.T) {
	g := NewGomegaWithT(t)

	pc := newPodFailurePodChaos("pc-test")
	objects, pods := generateNRunningPods("pc-test", 5)
	job := newPodFailureJob(pc, objects...)
	job.cli = fake.NewSimpleClientset(pc)
	record := &v1alpha1.PodChaosExperimentStatus{}

	job.setRecordPods(record, pods...)
	g.Expect(string(record.Phase)).To(Equal(string(v1alpha1.ExperimentPhaseRunning)))
	g.Expect(len(record.Pods)).To(Equal(5))
}
