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

package kernelchaos

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	daemonpb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	kernelpb "github.com/chaos-mesh/chaos-mesh/pkg/chaoskernel/pb"
)

func TestKernelChaos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KernelChaos controller")
}

type pidService struct {
	daemonpb.UnimplementedChaosDaemonServer
	requests chan *daemonpb.ContainerRequest
}

func (s *pidService) ContainerGetPid(_ context.Context, in *daemonpb.ContainerRequest) (*daemonpb.ContainerResponse, error) {
	s.requests <- in
	if in.GetContainerId() != "fixture" {
		return nil, status.Error(codes.NotFound, "unknown container")
	}
	return &daemonpb.ContainerResponse{Pid: 73}, nil
}

type kernelService struct {
	kernelpb.UnimplementedBPFKIServiceServer
	applied   chan *kernelpb.FailKernRequest
	recovered chan *kernelpb.FailKernRequest
	attempts  atomic.Int32
}

func (s *kernelService) FailMMOrBIO(_ context.Context, req *kernelpb.FailKernRequest) (*kernelpb.StatusResponse, error) {
	s.applied <- req
	return &kernelpb.StatusResponse{}, nil
}

func (s *kernelService) RecoverMMOrBIO(_ context.Context, req *kernelpb.FailKernRequest) (*kernelpb.StatusResponse, error) {
	s.recovered <- req
	if s.attempts.Add(1) == 1 {
		return nil, status.Error(codes.Unavailable, "kernel service temporarily unavailable")
	}
	return &kernelpb.StatusResponse{}, nil
}

func startKernelTestGRPC(server *grpc.Server) int {
	GinkgoHelper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	DeferCleanup(func() {
		server.Stop()
		Eventually(done).Should(Receive())
	})
	return listener.Addr().(*net.TCPAddr).Port
}

var _ = Describe("KernelChaos records", func() {
	const recordID = "workload/app/fixture"
	var impl *Impl
	var chaos *v1alpha1.KernelChaos
	var daemon *pidService
	var kernel *kernelService

	BeforeEach(func() {
		daemon = &pidService{requests: make(chan *daemonpb.ContainerRequest, 8)}
		daemonServer := grpc.NewServer()
		daemonpb.RegisterChaosDaemonServer(daemonServer, daemon)
		kernel = &kernelService{applied: make(chan *kernelpb.FailKernRequest, 2), recovered: make(chan *kernelpb.FailKernRequest, 2)}
		kernelServer := grpc.NewServer()
		kernelpb.RegisterBPFKIServiceServer(kernelServer, kernel)

		old := config.ControllerCfg
		cfg := *old
		cfg.ChaosDaemonPort = startKernelTestGRPC(daemonServer)
		cfg.BPFKIPort = startKernelTestGRPC(kernelServer)
		cfg.ChaosMeshCACert = ""
		config.ControllerCfg = &cfg
		DeferCleanup(func() { config.ControllerCfg = old })

		node := "fixture-node"
		endpoints := &discoveryv1.EndpointSlice{
			ObjectMeta:  metav1.ObjectMeta{Namespace: cfg.Namespace, Name: "chaos-daemon-fixture"},
			AddressType: discoveryv1.AddressTypeIPv4,
			Endpoints:   []discoveryv1.Endpoint{{NodeName: &node, Addresses: []string{"127.0.0.1"}}},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: "workload", Name: "app"},
			Spec:       corev1.PodSpec{NodeName: node},
			Status:     corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "app", ContainerID: "containerd://fixture"}}},
		}
		pending := pod.DeepCopy()
		pending.Name = "pending"
		pending.Status.ContainerStatuses = nil

		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(discoveryv1.AddToScheme(scheme)).To(Succeed())
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod, pending, endpoints).Build()
		impl = NewImpl(kube, logr.Discard(), &chaosdaemon.ClientBuilder{Reader: kube}).Impl.(*Impl)
		chaos = &v1alpha1.KernelChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: "workload", Name: "experiment"},
			Spec: v1alpha1.KernelChaosSpec{FailKernRequest: v1alpha1.FailKernRequest{
				FailType: 1, Headers: []string{"fixture.h"}, Probability: 25, Times: 2,
				Callchain: []v1alpha1.Frame{{Funcname: "fixture", Parameters: "value", Predicate: "fixture-condition"}},
			}},
		}
	})

	It("uses the container PID for apply and retries recover without changing the record phase", func(ctx SpecContext) {
		records := []*v1alpha1.Record{{Id: recordID}}
		Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))
		frame := &kernelpb.FailKernRequestFrame{Funcname: "fixture", Parameters: "value", Predicate: "fixture-condition"}
		var applied *kernelpb.FailKernRequest
		Expect(kernel.applied).To(Receive(&applied))
		Expect(proto.Equal(applied, &kernelpb.FailKernRequest{
			Pid: 73, Ftype: kernelpb.FailKernRequest_PAGE, Headers: []string{"fixture.h"},
			Probability: 0.25, Times: 2, Callchain: []*kernelpb.FailKernRequestFrame{frame},
		})).To(BeTrue())

		phase, err := impl.Recover(ctx, 0, records, chaos)
		Expect(phase).To(Equal(v1alpha1.Injected))
		Expect(status.Code(err)).To(Equal(codes.Unavailable))
		Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
		Expect(kernel.recovered).To(HaveLen(2))
		for range 2 {
			Expect(proto.Equal(<-kernel.recovered, &kernelpb.FailKernRequest{
				Pid: 73, Callchain: []*kernelpb.FailKernRequestFrame{frame},
			})).To(BeTrue())
		}
		Expect(daemon.requests).To(HaveLen(3))
		for range 3 {
			request := <-daemon.requests
			Expect(request.GetContainerId()).To(Equal("fixture"))
			Expect(request.GetAction().GetAction()).To(Equal(daemonpb.ContainerAction_GETPID))
		}
	}, SpecTimeout(10*time.Second))

	It("handles invalid, missing and not-yet-running targets without calling the kernel service", func(ctx SpecContext) {
		for _, record := range []string{"invalid", "workload/pending/fixture", "workload/app/missing"} {
			records := []*v1alpha1.Record{{Id: record}}
			phase, err := impl.Apply(ctx, 0, records, chaos)
			Expect(err).To(HaveOccurred())
			Expect(phase).To(Equal(v1alpha1.NotInjected))
			phase, err = impl.Recover(ctx, 0, records, chaos)
			Expect(err).To(HaveOccurred())
			Expect(phase).To(Equal(v1alpha1.Injected))
		}
		vanished := []*v1alpha1.Record{{Id: "workload/vanished/fixture"}}
		Expect(impl.Apply(ctx, 0, vanished, chaos)).To(Equal(v1alpha1.NotInjected))
		Expect(impl.Recover(ctx, 0, vanished, chaos)).To(Equal(v1alpha1.NotInjected))
		Expect(kernel.applied).To(BeEmpty())
		Expect(kernel.recovered).To(BeEmpty())
	}, SpecTimeout(10*time.Second))
})
