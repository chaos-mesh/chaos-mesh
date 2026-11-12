// Copyright 2026 Chaos Mesh Authors.
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

package steps

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	restfake "k8s.io/client-go/rest/fake"
	clienttesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func runningContainerPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx", Namespace: "test", UID: "pod-1", Labels: map[string]string{"app": "nginx"},
		},
		Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
			Name: "nginx", ContainerID: "container://original", RestartCount: 1, Ready: true,
			State:                corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
			LastTerminationState: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{}},
		}}},
	}
}

var _ = Describe("Container restart evidence", func() {
	It("does not accept termination from before recording", func() {
		pod := runningContainerPod()
		tc := &TestContext{KubeCli: fake.NewSimpleClientset(pod), Namespace: "test"}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		Expect(tc.containerTerminated(ctx, "nginx", "app=nginx")).NotTo(Succeed())
	})
})

var _ = Describe("Recorded container targets", func() {
	It("does not attribute a natural restart of a terminated baseline to chaos", func() {
		pod := runningContainerPod()
		pod.Status.ContainerStatuses[0].Ready = false
		pod.Status.ContainerStatuses[0].State = corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{}}
		cli := fake.NewSimpleClientset(pod)
		tc := &TestContext{Namespace: "test", KubeCli: cli}
		recordErr := tc.recordContainerID(context.Background(), "nginx", "app=nginx")
		pod.Status.ContainerStatuses[0].State = corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}
		pod.Status.ContainerStatuses[0].Ready = true
		pod.Status.ContainerStatuses[0].ContainerID = "container://natural-restart"
		pod.Status.ContainerStatuses[0].RestartCount++
		_, err := cli.CoreV1().Pods("test").UpdateStatus(context.Background(), pod, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		ctx, cancel := shortContext()
		defer cancel()
		Expect(recordErr).To(HaveOccurred())
		Expect(tc.containerIDChanged(ctx)).To(HaveOccurred())
		Expect(tc.containerRunningAndReady(ctx, "nginx", "app=nginx")).To(HaveOccurred())
	})

	It("rejects loss of the recorded target during stability", func() {
		pod := runningContainerPod()
		cli := fake.NewSimpleClientset(pod)
		tc := &TestContext{Namespace: "test", KubeCli: cli}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		pod.Labels = nil
		_, err := cli.CoreV1().Pods("test").Update(context.Background(), pod, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(tc.containerIDNotChanged(context.Background(), time.Millisecond, 10*time.Millisecond)).To(HaveOccurred())
	})

	It("preserves the recorded selector and container for every observation", func() {
		pod := runningContainerPod()
		pod.Labels = map[string]string{"role": "worker"}
		pod.Status.ContainerStatuses[0].Name = "worker"
		cli := fake.NewSimpleClientset(pod)
		tc := &TestContext{Namespace: "test", KubeCli: cli}
		Expect(tc.recordContainerID(context.Background(), "worker", "role=worker")).To(Succeed())
		Expect(tc.containerIDNotChanged(context.Background(), time.Millisecond, 10*time.Millisecond)).To(Succeed())
		pod.Status.ContainerStatuses[0].ContainerID = "container://new"
		pod.Status.ContainerStatuses[0].RestartCount++
		pod.Status.ContainerStatuses[0].LastTerminationState = corev1.ContainerState{}
		_, err := cli.CoreV1().Pods("test").UpdateStatus(context.Background(), pod, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(tc.containerIDChanged(context.Background())).To(Succeed())
		Expect(tc.containerTerminated(context.Background(), "worker", "role=worker")).To(Succeed())
		Expect(tc.containerRunningAndReady(context.Background(), "worker", "role=worker")).To(Succeed())
	})

	It("does not hide API errors during stability", func() {
		apiError := fmt.Errorf("rate limit: %w", context.DeadlineExceeded)
		cli := fake.NewSimpleClientset(runningContainerPod())
		tc := &TestContext{Namespace: "test", KubeCli: cli}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		cli.PrependReactor("list", "pods", func(clienttesting.Action) (bool, runtime.Object, error) { return true, nil, apiError })
		Expect(tc.containerIDNotChanged(context.Background(), time.Millisecond, time.Millisecond)).To(MatchError(apiError))
	})
})

var _ = Describe("Pod stability evidence", func() {
	It("does not treat failed or empty pod observations as a stable pause", func() {
		tc := &TestContext{Namespace: "test", KubeCli: fake.NewSimpleClientset()}
		Expect(tc.noFurtherPodsShouldBeKilledWithinMinutes(context.Background(), 1)).To(HaveOccurred())
		cli := fake.NewSimpleClientset(runningContainerPod())
		tc.KubeCli = cli
		calls := 0
		cli.PrependReactor("list", "pods", func(clienttesting.Action) (bool, runtime.Object, error) {
			calls++
			if calls > 1 {
				return true, nil, fmt.Errorf("list: %w", context.DeadlineExceeded)
			}
			return false, nil, nil
		})
		Expect(tc.noFurtherPodsShouldBeKilledWithinMinutes(context.Background(), 1)).To(HaveOccurred())
	})
})

var _ = Describe("Observation deadline boundaries", func() {
	DescribeTable("keeps one-shot pause observations within their deadline", func(mode string, succeeds bool) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{v1alpha1.SchemeGroupVersion})
		mapper.Add(v1alpha1.SchemeGroupVersion.WithKind("PodChaos"), meta.RESTScopeNamespace)
		chaos := &v1alpha1.PodChaos{Name: "kill", Namespace: "test", Spec: v1alpha1.PodChaosSpec{Action: v1alpha1.PodKillAction}}
		chaos.APIVersion = v1alpha1.SchemeGroupVersion.String()
		chaos.Kind = "PodChaos"
		observations := 0
		cli, err := client.New(&rest.Config{Host: "http://steps.invalid"}, client.Options{Log: GinkgoLogr, Scheme: scheme, Mapper: mapper, HTTPClient: restfake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			if _, bounded := req.Context().Deadline(); req.Method != http.MethodGet || !bounded {
				return jsonResponse(chaos), nil
			}
			observations++
			switch mode {
			case "blocked last query":
				if observations == 5 {
					<-req.Context().Done()
					return nil, req.Context().Err()
				}
			case "API deadline":
				return nil, fmt.Errorf("upstream timeout: %w", context.DeadlineExceeded)
			case "parent cancellation":
				cancel()
				return nil, req.Context().Err()
			}
			select {
			case <-time.After(2 * time.Millisecond):
				return jsonResponse(chaos), nil
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		})})
		Expect(err).NotTo(HaveOccurred())
		tc := &TestContext{Namespace: "test", Client: cli}
		started := time.Now()
		err = tc.theChaosExperimentIsPaused(ctx, "kill")
		if succeeds {
			Expect(err).NotTo(HaveOccurred())
			Expect(ctx.Err()).NotTo(HaveOccurred())
			Expect(time.Since(started)).To(BeNumerically(">=", 5*time.Second))
		} else {
			Expect(err).To(HaveOccurred())
		}
	},
		Entry("healthy requests complete the five second window", "healthy", true),
		Entry("a request started before the deadline must still finish", "blocked last query", false),
		Entry("an active-window API deadline fails", "API deadline", false),
		Entry("parent cancellation fails", "parent cancellation", false))

	It("accepts stable container observations through the complete window", func() {
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			if _, bounded := req.Context().Deadline(); bounded {
				select {
				case <-time.After(2 * time.Millisecond):
				case <-req.Context().Done():
					return nil, req.Context().Err()
				}
			}
			return jsonResponse(&corev1.PodList{Items: []corev1.Pod{*runningContainerPod()}}), nil
		})}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		started := time.Now()
		Expect(tc.containerIDNotChanged(context.Background(), 10*time.Millisecond, 50*time.Millisecond)).To(Succeed())
		Expect(time.Since(started)).To(BeNumerically(">=", 50*time.Millisecond))
	})

	It("accepts stable Pod observations through the complete minute", func() {
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			select {
			case <-time.After(2 * time.Millisecond):
				return jsonResponse(&corev1.PodList{Items: []corev1.Pod{*runningContainerPod()}}), nil
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		})}
		started := time.Now()
		Expect(tc.noFurtherPodsShouldBeKilledWithinMinutes(context.Background(), 1)).To(Succeed())
		Expect(time.Since(started)).To(BeNumerically(">=", time.Minute))
	})

	It("does not accept an elapsed parent deadline as successful container observation", func() {
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			select {
			case <-time.After(2 * time.Millisecond):
				return jsonResponse(&corev1.PodList{Items: []corev1.Pod{*runningContainerPod()}}), nil
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		})}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		err := tc.containerIDNotChanged(ctx, 10*time.Millisecond, 100*time.Millisecond)
		cancel()
		Expect(err).To(MatchError(context.DeadlineExceeded))
	})
})
