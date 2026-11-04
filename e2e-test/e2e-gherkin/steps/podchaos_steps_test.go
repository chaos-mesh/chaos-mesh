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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

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
	DescribeTable("retries transient Pod queries while waiting for a restart", func(step func(*TestContext, context.Context) error) {
		pod := runningContainerPod()
		cli := fake.NewSimpleClientset(pod)
		tc := &TestContext{Namespace: "test", KubeCli: cli}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		pod.Status.ContainerStatuses[0].ContainerID = "container://new"
		pod.Status.ContainerStatuses[0].RestartCount++
		_, err := cli.CoreV1().Pods("test").UpdateStatus(context.Background(), pod, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		queries := 0
		cli.PrependReactor("list", "pods", func(clienttesting.Action) (bool, runtime.Object, error) {
			queries++
			if queries == 1 {
				return true, nil, fmt.Errorf("temporary API failure")
			}
			return false, nil, nil
		})
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		Expect(step(tc, ctx)).To(Succeed())
	},
		Entry("changed ID", (*TestContext).containerIDChanged),
		Entry("terminated", func(tc *TestContext, ctx context.Context) error {
			return tc.containerTerminated(ctx, "nginx", "app=nginx")
		}),
		Entry("running and ready", func(tc *TestContext, ctx context.Context) error {
			return tc.containerRunningAndReady(ctx, "nginx", "app=nginx")
		}))

	DescribeTable("rejects missing baselines and replaced Pods without waiting", func(step func(*TestContext, context.Context) error) {
		pod := runningContainerPod()
		cli := fake.NewSimpleClientset(pod)
		tc := &TestContext{Namespace: "test", KubeCli: cli}
		ctx, cancel := shortContext()
		defer cancel()
		Expect(step(tc, ctx)).To(HaveOccurred())
		Expect(ctx.Err()).NotTo(HaveOccurred())
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		pod.UID = "replacement"
		_, err := cli.CoreV1().Pods("test").Update(context.Background(), pod, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(step(tc, ctx)).To(HaveOccurred())
		Expect(ctx.Err()).NotTo(HaveOccurred())
	},
		Entry("changed ID", (*TestContext).containerIDChanged),
		Entry("terminated", func(tc *TestContext, ctx context.Context) error {
			return tc.containerTerminated(ctx, "nginx", "app=nginx")
		}),
		Entry("running and ready", func(tc *TestContext, ctx context.Context) error {
			return tc.containerRunningAndReady(ctx, "nginx", "app=nginx")
		}))

	It("does not accept termination from before recording", func() {
		pod := runningContainerPod()
		tc := &TestContext{KubeCli: fake.NewSimpleClientset(pod), Namespace: "test"}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		Expect(tc.containerTerminated(ctx, "nginx", "app=nginx")).NotTo(Succeed())
	})
	It("does not accept an empty current container ID as a restart", func() {
		pod := runningContainerPod()
		tc := &TestContext{KubeCli: fake.NewSimpleClientset(pod), Namespace: "test"}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		pod.Status.ContainerStatuses[0].ContainerID = ""
		_, err := tc.KubeCli.CoreV1().Pods("test").UpdateStatus(context.Background(), pod, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		Expect(tc.containerIDChanged(ctx)).NotTo(Succeed())
	})
})

var _ = Describe("Recorded container targets", func() {
	DescribeTable("requires an observable recording baseline", func(change func(*corev1.Pod), valid bool) {
		pod := runningContainerPod()
		change(pod)
		tc := &TestContext{Namespace: "test", KubeCli: fake.NewSimpleClientset(pod)}
		err := tc.recordContainerID(context.Background(), "nginx", "app=nginx")
		if valid {
			Expect(err).NotTo(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
		}
	},
		Entry("valid running baseline", func(*corev1.Pod) {}, true),
		Entry("assigned ID before readiness", func(p *corev1.Pod) { p.Status.ContainerStatuses[0].Ready = false }, true),
		Entry("missing Pod UID", func(p *corev1.Pod) { p.UID = "" }, false),
		Entry("missing statuses", func(p *corev1.Pod) { p.Status.ContainerStatuses = nil }, false),
		Entry("different container", func(p *corev1.Pod) { p.Status.ContainerStatuses[0].Name = "sidecar" }, false),
		Entry("empty container ID", func(p *corev1.Pod) { p.Status.ContainerStatuses[0].ContainerID = "" }, false),
		Entry("missing selected Pod", func(p *corev1.Pod) { p.Labels = nil }, false))

	It("rejects ambiguous Pods", func() {
		pod := runningContainerPod()
		extra := pod.DeepCopy()
		extra.Name = "extra"
		extra.UID = "extra"
		tc := &TestContext{Namespace: "test", KubeCli: fake.NewSimpleClientset(pod, extra)}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(HaveOccurred())
	})

	DescribeTable("rejects loss of the recorded target during stability", func(change func(*corev1.Pod)) {
		pod := runningContainerPod()
		cli := fake.NewSimpleClientset(pod)
		tc := &TestContext{Namespace: "test", KubeCli: cli}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		change(pod)
		_, err := cli.CoreV1().Pods("test").Update(context.Background(), pod, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(tc.containerIDNotChanged(context.Background(), time.Millisecond, 10*time.Millisecond)).To(HaveOccurred())
	},
		Entry("no matching pods", func(p *corev1.Pod) { p.Labels = nil }),
		Entry("empty status list", func(p *corev1.Pod) { p.Status.ContainerStatuses = nil }),
		Entry("named container missing", func(p *corev1.Pod) { p.Status.ContainerStatuses[0].Name = "other" }),
		Entry("empty container ID", func(p *corev1.Pod) { p.Status.ContainerStatuses[0].ContainerID = "" }),
		Entry("Pod replaced", func(p *corev1.Pod) { p.UID = "replacement" }),
		Entry("container ID changed", func(p *corev1.Pod) { p.Status.ContainerStatuses[0].ContainerID = "container://new" }),
		Entry("restart count changed", func(p *corev1.Pod) { p.Status.ContainerStatuses[0].RestartCount++ }))

	It("rejects an additional selected Pod after recording", func() {
		pod := runningContainerPod()
		cli := fake.NewSimpleClientset(pod)
		tc := &TestContext{Namespace: "test", KubeCli: cli}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		pod.Name = "extra"
		_, err := cli.CoreV1().Pods("test").Create(context.Background(), pod, metav1.CreateOptions{})
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

	DescribeTable("requires both a changed ID and increased restart count", func(id string, count int32, ready bool) {
		pod := runningContainerPod()
		cli := fake.NewSimpleClientset(pod)
		tc := &TestContext{Namespace: "test", KubeCli: cli}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		pod.Status.ContainerStatuses[0].ContainerID = id
		pod.Status.ContainerStatuses[0].RestartCount = count
		pod.Status.ContainerStatuses[0].Ready = ready
		_, err := cli.CoreV1().Pods("test").UpdateStatus(context.Background(), pod, metav1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		ctx, cancel := shortContext()
		defer cancel()
		if !ready {
			Expect(tc.containerRunningAndReady(ctx, "nginx", "app=nginx")).To(HaveOccurred())
		} else {
			Expect(tc.containerIDChanged(ctx)).To(HaveOccurred())
		}
	}, Entry("ID changed without restart", "container://new", int32(1), true), Entry("count changed without ID", "container://original", int32(2), true), Entry("restart not ready", "container://new", int32(2), false))

	It("rejects missing baselines and mismatched explicit targets", func() {
		tc := &TestContext{Namespace: "test", KubeCli: fake.NewSimpleClientset(runningContainerPod())}
		Expect(tc.containerIDChanged(context.Background())).To(HaveOccurred())
		Expect(tc.containerIDNotChanged(context.Background(), time.Millisecond, time.Millisecond)).To(HaveOccurred())
		Expect(tc.containerTerminated(context.Background(), "nginx", "app=nginx")).To(HaveOccurred())
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		Expect(tc.containerTerminated(context.Background(), "other", "app=nginx")).To(HaveOccurred())
		Expect(tc.containerRunningAndReady(context.Background(), "nginx", "app=other")).To(HaveOccurred())
	})

	DescribeTable("does not hide API errors during stability", func(apiError error) {
		cli := fake.NewSimpleClientset(runningContainerPod())
		tc := &TestContext{Namespace: "test", KubeCli: cli}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		cli.PrependReactor("list", "pods", func(clienttesting.Action) (bool, runtime.Object, error) { return true, nil, apiError })
		Expect(tc.containerIDNotChanged(context.Background(), time.Millisecond, time.Millisecond)).To(MatchError(apiError))
	}, Entry("ordinary error", fmt.Errorf("list failed")), Entry("wrapped deadline", fmt.Errorf("rate limit: %w", context.DeadlineExceeded)), Entry("wrapped cancellation", fmt.Errorf("transport: %w", context.Canceled)))

	It("fails on parent cancellation before its observation completes", func() {
		tc := &TestContext{Namespace: "test", KubeCli: fake.NewSimpleClientset(runningContainerPod())}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		ctx, cancel := shortContext()
		defer cancel()
		Expect(tc.containerIDNotChanged(ctx, time.Millisecond, time.Second)).To(MatchError(context.DeadlineExceeded))
		canceled, stop := context.WithCancel(context.Background())
		stop()
		Expect(tc.containerIDNotChanged(canceled, time.Millisecond, time.Second)).To(MatchError(context.Canceled))
	})

	It("bounds a blocked API call and reports its deadline as a failure", func() {
		calls := 0
		tc := &TestContext{Namespace: "test", KubeCli: requestClient(func(req *http.Request) (*http.Response, error) {
			calls++
			Expect(req.URL.Query().Get("labelSelector")).To(Equal("app=nginx"))
			if calls == 1 {
				return jsonResponse(&corev1.PodList{Items: []corev1.Pod{*runningContainerPod()}}), nil
			}
			_, bounded := req.Context().Deadline()
			Expect(bounded).To(BeTrue())
			<-req.Context().Done()
			return nil, fmt.Errorf("request interrupted: %w", req.Context().Err())
		})}
		Expect(tc.recordContainerID(context.Background(), "nginx", "app=nginx")).To(Succeed())
		Expect(tc.containerIDNotChanged(context.Background(), time.Millisecond, 10*time.Millisecond)).To(HaveOccurred())
	})
})

var _ = Describe("Pod chaos adapters", func() {
	It("retries transient queries while waiting for a non-one-shot pause", func() {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		chaos := &v1alpha1.PodChaos{Name: "failure", Namespace: "test", Spec: v1alpha1.PodChaosSpec{Action: v1alpha1.PodFailureAction}}
		queries := 0
		cli := clientfake.NewClientBuilder().WithScheme(scheme).WithObjects(chaos).WithInterceptorFuncs(interceptor.Funcs{Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			queries++
			if queries == 2 {
				return fmt.Errorf("temporary API failure")
			}
			if err := c.Get(ctx, key, obj, opts...); err != nil {
				return err
			}
			if queries > 2 {
				obj.(*v1alpha1.PodChaos).Status.Experiment.DesiredPhase = v1alpha1.StoppedPhase
			}
			return nil
		}}).Build()
		tc := &TestContext{Namespace: "test", Client: cli}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		Expect(tc.theChaosExperimentIsPaused(ctx, "failure")).To(Succeed())
	})

	It("constructs PodChaos actions consistently", func() {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		cli := clientfake.NewClientBuilder().WithScheme(scheme).Build()
		tc := &TestContext{Namespace: "test", Client: cli}
		Expect(tc.aChaosNamedWithModeIsAppliedToPodsWithLabel("PodKill", "kill", "all", "role=worker")).To(Succeed())
		Expect(tc.applyContainerKill("container-kill", "one", "worker", "role=worker")).To(Succeed())
		for _, name := range []string{"kill", "container-kill"} {
			chaos := &v1alpha1.PodChaos{}
			Expect(cli.Get(context.Background(), client.ObjectKey{Namespace: "test", Name: name}, chaos)).To(Succeed())
			Expect(chaos.Spec.Selector.Namespaces).To(Equal([]string{"test"}))
			Expect(chaos.Spec.Selector.LabelSelectors).To(Equal(map[string]string{"role": "worker"}))
			if name == "kill" {
				Expect(chaos.Spec.Action).To(Equal(v1alpha1.PodKillAction))
				Expect(chaos.Spec.Mode).To(Equal(v1alpha1.AllMode))
				Expect(chaos.Spec.ContainerNames).To(BeEmpty())
			} else {
				Expect(chaos.Spec.Action).To(Equal(v1alpha1.ContainerKillAction))
				Expect(chaos.Spec.Mode).To(Equal(v1alpha1.OneMode))
				Expect(chaos.Spec.ContainerNames).To(Equal([]string{"worker"}))
			}
		}
	})

	It("propagates one-shot pause API failures and parent cancellation", func() {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		chaos := &v1alpha1.PodChaos{Name: "kill", Namespace: "test", Spec: v1alpha1.PodChaosSpec{Action: v1alpha1.ContainerKillAction}}
		gets := 0
		cli := clientfake.NewClientBuilder().WithScheme(scheme).WithObjects(chaos).WithInterceptorFuncs(interceptor.Funcs{Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			gets++
			if gets > 1 {
				return fmt.Errorf("API: %w", context.DeadlineExceeded)
			}
			return c.Get(ctx, key, obj, opts...)
		}}).Build()
		tc := &TestContext{Namespace: "test", Client: cli}
		Expect(tc.theChaosExperimentIsPaused(context.Background(), "kill")).To(HaveOccurred())
		tc.Client = clientfake.NewClientBuilder().WithScheme(scheme).WithObjects(chaos).Build()
		ctx, cancel := shortContext()
		defer cancel()
		Expect(tc.theChaosExperimentIsPaused(ctx, "kill")).To(MatchError(context.DeadlineExceeded))
	})

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
