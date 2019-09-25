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
	"math/rand"
	"testing"

	. "github.com/onsi/gomega"
	gtype "github.com/onsi/gomega/types"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

const (
	NAMESPACE  = metav1.NamespaceDefault
	IDENTIFIER = "chaos-operator-id"
)

func newPod(name string, status v1.PodPhase) v1.Pod {
	return v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: NAMESPACE,
			Labels: map[string]string{
				"chaos-operator/identifier": IDENTIFIER,
			},
		},
		Status: v1.PodStatus{
			Phase: status,
		},
	}
}

func generateNPods(namePrefix string, n int, status v1.PodPhase) ([]runtime.Object, []v1.Pod) {
	var podObjects []runtime.Object
	var pods []v1.Pod
	for i := 0; i < n; i++ {
		pod := newPod(fmt.Sprintf("%s%d", namePrefix, i), status)
		podObjects = append(podObjects, &pod)
		pods = append(pods, pod)
	}

	return podObjects, pods
}

func generateNRunningPods(namePrefix string, n int) ([]runtime.Object, []v1.Pod) {
	return generateNPods(namePrefix, n, v1.PodRunning)
}

func newPodKillJob(pc *v1alpha1.PodChaos, objects ...runtime.Object) PodKillJob {
	kubeCli := kubefake.NewSimpleClientset(objects...)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeCli, 0)

	podLister := kubeInformerFactory.Core().V1().Pods().Lister()

	return PodKillJob{
		podChaos:  pc,
		kubeCli:   kubeCli,
		podLister: podLister,
	}
}

func newPodChaos(name string) *v1alpha1.PodChaos {
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
			Action: v1alpha1.PodKillAction,
		},
	}
}

func getPodList(client kubernetes.Interface) *v1.PodList {
	podList, _ := client.CoreV1().Pods(NAMESPACE).List(metav1.ListOptions{})
	return podList
}

func TestPodKillJobEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		job1Name      string
		job1Version   string
		job2Name      string
		job2Version   string
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name:          "same name and version",
			job1Name:      "test-job",
			job1Version:   "1",
			job2Name:      "test-job",
			job2Version:   "1",
			expectedValue: true,
		},
		{
			name:          "same name, different version ",
			job1Name:      "test-job",
			job1Version:   "1",
			job2Name:      "test-job",
			job2Version:   "2",
			expectedValue: false,
		},
		{
			name:          "different name, different version ",
			job1Name:      "test-job-1",
			job1Version:   "1",
			job2Name:      "test-job-2",
			job2Version:   "2",
			expectedValue: false,
		},
		{
			name:          "different name, same version ",
			job1Name:      "test-job-1",
			job1Version:   "1",
			job2Name:      "test-job-2",
			job2Version:   "1",
			expectedValue: false,
		},
	}

	for _, tc := range tcs {
		job1 := newPodKillJob(newPodChaos(tc.job1Name))
		job1.podChaos.ResourceVersion = tc.job1Version
		job2 := newPodKillJob(newPodChaos(tc.job2Name))
		job2.podChaos.ResourceVersion = tc.job2Version

		g.Expect(job1.Equal(job2)).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestPodKillJobDeletePod(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name              string
		podName           string
		podsCount         int
		expectedPodsCount int
	}

	tcs := []TestCase{
		{
			name:              "one pod",
			podsCount:         1,
			expectedPodsCount: 0,
		},
		{
			name:              "one pod",
			podsCount:         3,
			expectedPodsCount: 2,
		},
	}

	for _, tc := range tcs {
		objects, pods := generateNRunningPods("test", tc.podsCount)
		job := newPodKillJob(newPodChaos("job"), objects...)

		g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.podsCount), tc.name)
		g.Expect(job.deletePod(pods[rand.Intn(len(pods))])).Should(Succeed())
		g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.expectedPodsCount), tc.name)
	}
}

func TestGetDeleteOptsForPod(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name                   string
		terminationGracePeriod *int64
		expectedGracePeriod    *int64
	}

	// helper method to create *int64 from int64 since Go does not allow
	// use of address operator (&) on numeric constants
	newInt64Pointer := func(val int64) *int64 {
		return &val
	}

	defaultGracePeriod := newInt64Pointer(0)
	tcs := []TestCase{
		{
			name:                   "nil pod TerminationGracePeriod",
			terminationGracePeriod: nil,
			expectedGracePeriod:    defaultGracePeriod,
		},
		{
			name:                   "pod TerminateGracePeriod lower than configured grace period",
			terminationGracePeriod: newInt64Pointer(*defaultGracePeriod - int64(1)),
			expectedGracePeriod:    defaultGracePeriod,
		},
		{
			name:                   "pod TerminationGracePeriod higher than configured grace period",
			terminationGracePeriod: newInt64Pointer(*defaultGracePeriod + int64(1)),
			expectedGracePeriod:    newInt64Pointer(*defaultGracePeriod + int64(1)),
		},
	}

	for _, tc := range tcs {
		pod := newPod("app", v1.PodRunning)
		pod.Spec.TerminationGracePeriodSeconds = tc.terminationGracePeriod

		job := newPodKillJob(newPodChaos("job"), &pod)
		deleteOpts := job.getDeleteOptsForPod(pod)

		g.Expect(deleteOpts.GracePeriodSeconds).To(Equal(tc.expectedGracePeriod), tc.name)
	}
}

func TestPodKillJobDeleteRandomPod(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name         string
		lenPods      int
		expectedPods int
	}

	tcs := []TestCase{
		{
			name:         "3 pods",
			lenPods:      3,
			expectedPods: 2,
		},
		{
			name:         "5 pods",
			lenPods:      5,
			expectedPods: 4,
		},
		{
			name:         "0 pods",
			lenPods:      0,
			expectedPods: 0,
		},
	}

	for _, tc := range tcs {
		objects, pods := generateNRunningPods("pod-kill-", tc.lenPods)
		job := newPodKillJob(newPodChaos(tc.name), objects...)
		g.Expect(job.deleteRandomPod(pods)).Should(Succeed(), tc.name)
		g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.expectedPods), tc.name)
	}
}

func TestPodKillJobDeleteAllPods(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name              string
		podsCount         int
		expectedPodsCount int
	}

	tcs := []TestCase{
		{
			name:              "5 pods",
			podsCount:         5,
			expectedPodsCount: 0,
		},
		{
			name:              "1 pods",
			podsCount:         1,
			expectedPodsCount: 0,
		},
		{
			name:              "0 pods",
			podsCount:         0,
			expectedPodsCount: 0,
		},
	}

	for _, tc := range tcs {
		objects, pods := generateNRunningPods("pod-kill-", tc.podsCount)

		job := newPodKillJob(newPodChaos("job"), objects...)

		g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.podsCount), tc.name)
		g.Expect(job.deleteAllPods(pods)).Should(Succeed())
		g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.expectedPodsCount), tc.name)
	}
}

func TestPodKillJobDeleteFixedPods(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name              string
		fixedValue        string
		podsCount         int
		expectedPodsCount int
	}

	tcs := []TestCase{
		{
			name:              "fixed 0 pod",
			fixedValue:        "0",
			podsCount:         5,
			expectedPodsCount: 5,
		},
		{
			name:              "fixed 5 pod",
			fixedValue:        "5",
			podsCount:         5,
			expectedPodsCount: 0,
		},
		{
			name:              "fixed 5 pod, create 0 pod",
			fixedValue:        "5",
			podsCount:         0,
			expectedPodsCount: 0,
		},
		{
			name:              "fixed 2 pod",
			fixedValue:        "2",
			podsCount:         5,
			expectedPodsCount: 3,
		},
	}

	for _, tc := range tcs {
		pc := newPodChaos("pc-test")
		pc.Spec.Value = tc.fixedValue

		objects, pods := generateNRunningPods("pod-kill-", tc.podsCount)

		job := newPodKillJob(pc, objects...)

		g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.podsCount), tc.name)
		g.Expect(job.deleteFixedPods(pods)).Should(Succeed())
		g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.expectedPodsCount), tc.name)
	}
}

type resultF func() gtype.GomegaMatcher

func TestPodKillJobFixedPercentagePods(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name              string
		fixedValue        string
		podsCount         int
		expectedPodsCount int
		expectedResult    resultF
	}

	tcs := []TestCase{
		{
			name:              "fixed 0%% pod",
			fixedValue:        "0",
			podsCount:         10,
			expectedPodsCount: 10,
			expectedResult:    Succeed,
		},
		{
			name:              "fixed 100%% pod",
			fixedValue:        "100",
			podsCount:         10,
			expectedPodsCount: 0,
			expectedResult:    Succeed,
		},
		{
			name:              "fixed 100%% pod, create 0 pod",
			fixedValue:        "5",
			podsCount:         0,
			expectedPodsCount: 0,
			expectedResult:    Succeed,
		},
		{
			name:              "fixed 50%% pod",
			fixedValue:        "50",
			podsCount:         10,
			expectedPodsCount: 5,
			expectedResult:    Succeed,
		},
		{
			name:              "fixed 28%% pod",
			fixedValue:        "28",
			podsCount:         10,
			expectedPodsCount: 8,
			expectedResult:    Succeed,
		},
		{
			name:              "fixed 200%% pod",
			fixedValue:        "200",
			podsCount:         10,
			expectedPodsCount: 10,
			expectedResult:    HaveOccurred,
		},
		{
			name:              "fixed -10%% pod",
			fixedValue:        "-10",
			podsCount:         10,
			expectedPodsCount: 10,
			expectedResult:    HaveOccurred,
		},
	}

	for _, tc := range tcs {
		pc := newPodChaos("pc-test")
		pc.Spec.Value = tc.fixedValue

		objects, pods := generateNRunningPods("pod-kill-", tc.podsCount)

		job := newPodKillJob(pc, objects...)

		g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.podsCount), tc.name)
		g.Expect(job.deleteFixedPercentagePods(pods)).Should(tc.expectedResult())
		g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.expectedPodsCount), tc.name)
	}
}

func TestPodKillJobMaxPercentagePods(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name              string
		fixedValue        string
		podsCount         int
		expectedPodsCount int
		expectedResult    resultF
	}

	tcs := []TestCase{
		{
			name:              "fixed max 0%% pod",
			fixedValue:        "0",
			podsCount:         10,
			expectedPodsCount: 10,
			expectedResult:    Succeed,
		},
		{
			name:              "fixed max 100%% pod",
			fixedValue:        "100",
			podsCount:         10,
			expectedPodsCount: 0,
			expectedResult:    Succeed,
		},
		{
			name:              "fixed max 100%% pod, create 0 pod",
			fixedValue:        "5",
			podsCount:         0,
			expectedPodsCount: 0,
			expectedResult:    Succeed,
		},
		{
			name:              "fixed max 50%% pod",
			fixedValue:        "50",
			podsCount:         10,
			expectedPodsCount: 5,
			expectedResult:    Succeed,
		},
		{
			name:              "fixed max 28%% pod",
			fixedValue:        "28",
			podsCount:         10,
			expectedPodsCount: 8,
			expectedResult:    Succeed,
		},
		{
			name:              "fixed max 200%% pod",
			fixedValue:        "200",
			podsCount:         10,
			expectedPodsCount: 10,
			expectedResult:    HaveOccurred,
		},
		{
			name:              "fixed max -10%% pod",
			fixedValue:        "-10",
			podsCount:         10,
			expectedPodsCount: 10,
			expectedResult:    HaveOccurred,
		},
	}

	for i := 0; i < 5; i++ {
		for _, tc := range tcs {
			pc := newPodChaos("pc-test")
			pc.Spec.Value = tc.fixedValue

			objects, pods := generateNRunningPods("pod-kill-", tc.podsCount)

			job := newPodKillJob(pc, objects...)

			g.Expect(len(getPodList(job.kubeCli).Items)).To(Equal(tc.podsCount), tc.name)
			g.Expect(job.deleteMaxPercentagePods(pods)).Should(tc.expectedResult())
			g.Expect(len(getPodList(job.kubeCli).Items)).Should(BeNumerically(">=", tc.expectedPodsCount), tc.name)
		}
	}
}
