// Copyright Chaos Mesh Authors.
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

package podiochaos

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	daemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

func TestPodIOChaos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pod IO Chaos")
}

type ioDaemon struct {
	daemonclient.ChaosDaemonClientInterface
	requests []*pb.ApplyIOChaosRequest
	err      error
	closes   int
}

func (d *ioDaemon) ApplyIOChaos(_ context.Context, req *pb.ApplyIOChaosRequest, _ ...grpc.CallOption) (*pb.ApplyIOChaosResponse, error) {
	d.requests = append(d.requests, req)
	if d.err != nil {
		return nil, d.err
	}
	return &pb.ApplyIOChaosResponse{Instance: 29, StartTime: 301}, nil
}

func (d *ioDaemon) Close() error {
	d.closes++
	return nil
}

var _ = Describe("Reconciling PodIOChaos", func() {
	var (
		reconciler *Reconciler
		kube       client.Client
		pod        *corev1.Pod
		chaos      *v1alpha1.PodIOChaos
		daemon     *ioDaemon
		events     *record.FakeRecorder
	)
	ctx := context.Background()
	key := client.ObjectKey{Namespace: "apps", Name: "example"}
	request := ctrl.Request{NamespacedName: key}
	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
				{Name: "sidecar", ContainerID: "containerd://sidecar"},
				{Name: "main", ContainerID: "containerd://main"},
			}},
		}
		container := "main"
		chaos = &v1alpha1.PodIOChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name, Generation: 5},
			Spec: v1alpha1.PodIOChaosSpec{
				Container: &container, VolumeMountPath: "/data",
				Actions: []v1alpha1.IOChaosAction{{Type: v1alpha1.IoLatency, Latency: "10ms", Source: "ops/delay"}},
			},
			Status: v1alpha1.PodIOChaosStatus{ObservedGeneration: 4, Pid: 19, StartTime: 101},
		}
		kube = fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod, chaos).WithStatusSubresource(pod, chaos).Build()
		daemon = &ioDaemon{}
		DeferCleanup(mock.With("MockChaosDaemonClient", daemon))
		Expect(mock.On("MockChaosDaemonClient")).To(BeIdenticalTo(daemon), "run with the repository's failpoint instrumentation")
		events = record.NewFakeRecorder(4)
		reconciler = &Reconciler{Client: kube, Recorder: events, Log: logr.Discard(), ChaosDaemonClientBuilder: &chaosdaemon.ClientBuilder{Reader: kube}}
	})

	It("sends the action and prior process identity to the selected container and reconciles only once", func() {
		Expect(reconciler.Reconcile(ctx, request)).To(Equal(ctrl.Result{}))
		Expect(daemon.requests).To(HaveLen(1))
		rpc := daemon.requests[0]
		Expect(rpc.ContainerId).To(Equal("containerd://main"))
		Expect(rpc.Volume).To(Equal("/data"))
		Expect(rpc.Instance).To(BeNumerically("==", 19))
		Expect(rpc.StartTime).To(BeNumerically("==", 101))
		Expect(rpc.EnterNS).To(BeTrue())
		var actions []v1alpha1.IOChaosAction
		Expect(json.Unmarshal([]byte(rpc.Actions), &actions)).To(Succeed())
		Expect(actions).To(Equal(chaos.Spec.Actions))
		Expect(kube.Get(ctx, key, chaos)).To(Succeed())
		Expect(chaos.Status).To(Equal(v1alpha1.PodIOChaosStatus{ObservedGeneration: 5, Pid: 29, StartTime: 301}))
		Expect(reconciler.Reconcile(ctx, request)).To(Equal(ctrl.Result{}))
		Expect(daemon.requests).To(HaveLen(1))
		Expect(daemon.closes).To(Equal(1))
	})

	It("retains the old process identity after a daemon error and retries the same generation", func() {
		daemon.err = errors.New("daemon unavailable")
		Expect(reconciler.Reconcile(ctx, request)).To(Equal(ctrl.Result{Requeue: true}))
		Expect(kube.Get(ctx, key, chaos)).To(Succeed())
		Expect(chaos.Status.Pid).To(BeNumerically("==", 19))
		Expect(chaos.Status.StartTime).To(BeNumerically("==", 101))
		Expect(chaos.Status.ObservedGeneration).To(BeNumerically("==", 5))
		Expect(chaos.Status.FailedMessage).To(ContainSubstring("daemon unavailable"))
		Eventually(events.Events).Should(Receive(ContainSubstring("daemon unavailable")))
		daemon.err = nil
		Expect(reconciler.Reconcile(ctx, request)).To(Equal(ctrl.Result{}))
		Expect(daemon.requests).To(HaveLen(2))
		Expect(daemon.requests[1].Instance).To(BeNumerically("==", 19))
		Expect(kube.Get(ctx, key, chaos)).To(Succeed())
		Expect(chaos.Status).To(Equal(v1alpha1.PodIOChaosStatus{ObservedGeneration: 5, Pid: 29, StartTime: 301}))
	})

	It("reports an unknown target container without contacting the daemon and recovers when it appears", func() {
		pod.Status.ContainerStatuses = pod.Status.ContainerStatuses[:1]
		Expect(kube.Status().Update(ctx, pod)).To(Succeed())
		Expect(reconciler.Reconcile(ctx, request)).To(Equal(ctrl.Result{}))
		Expect(daemon.requests).To(BeEmpty())
		Expect(kube.Get(ctx, key, chaos)).To(Succeed())
		Expect(chaos.Status.FailedMessage).To(ContainSubstring("cannot find container with name main"))
		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, corev1.ContainerStatus{Name: "main", ContainerID: "containerd://replacement"})
		Expect(kube.Status().Update(ctx, pod)).To(Succeed())
		Expect(reconciler.Reconcile(ctx, request)).To(Equal(ctrl.Result{}))
		Expect(daemon.requests).To(HaveLen(1))
		Expect(daemon.requests[0].ContainerId).To(Equal("containerd://replacement"))
	})
})
