// Copyright 2021 Chaos Mesh Authors.
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

package dnschaos

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"

	dnspb "github.com/chaos-mesh/k8s_dns_chaos/pb"
	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	daemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

type dnsDaemonClient struct {
	daemonclient.ChaosDaemonClientInterface
	requests []*pb.SetDNSServerRequest
	err      error
	closed   int
}

func (d *dnsDaemonClient) SetDNSServer(_ context.Context, req *pb.SetDNSServerRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	d.requests = append(d.requests, req)
	return &emptypb.Empty{}, d.err
}

func (d *dnsDaemonClient) Close() error {
	d.closed++
	return nil
}

type dnsLifecycleServer struct {
	dnspb.UnimplementedDNSServer
	applied  chan *dnspb.SetDNSChaosRequest
	canceled chan *dnspb.CancelDNSChaosRequest
	behavior atomic.Int32
}

func (d *dnsLifecycleServer) response() (*dnspb.DNSChaosResponse, error) {
	switch d.behavior.Load() {
	case 1:
		return &dnspb.DNSChaosResponse{Msg: "fixture rejected rule"}, nil
	case 2:
		return nil, status.Error(codes.Unavailable, "fixture offline")
	default:
		return &dnspb.DNSChaosResponse{Result: true}, nil
	}
}

func (d *dnsLifecycleServer) SetDNSChaos(_ context.Context, req *dnspb.SetDNSChaosRequest) (*dnspb.DNSChaosResponse, error) {
	d.applied <- req
	return d.response()
}

func (d *dnsLifecycleServer) CancelDNSChaos(_ context.Context, req *dnspb.CancelDNSChaosRequest) (*dnspb.DNSChaosResponse, error) {
	d.canceled <- req
	return d.response()
}

func TestDNSChaosRecordLifecycle(t *testing.T) {
	g := NewWithT(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	g.Expect(err).NotTo(HaveOccurred())
	dns := &dnsLifecycleServer{applied: make(chan *dnspb.SetDNSChaosRequest, 8), canceled: make(chan *dnspb.CancelDNSChaosRequest, 8)}
	server := grpc.NewServer()
	dnspb.RegisterDNSServer(server, dns)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	oldConfig := config.ControllerCfg
	cfg := *oldConfig
	cfg.Namespace = "chaos-mesh"
	cfg.DNSServiceName = "dns"
	cfg.DNSServicePort = listener.Addr().(*net.TCPAddr).Port
	config.ControllerCfg = &cfg
	t.Cleanup(func() { config.ControllerCfg = oldConfig })

	daemon := &dnsDaemonClient{}
	cleanup := mock.With("MockChaosDaemonClient", daemon)
	t.Cleanup(func() { g.Expect(cleanup()).To(Succeed()) })
	target := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "workload", Name: "target"},
		Status:     corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "app", ContainerID: "containerd://fixture"}}},
	}
	matching := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: cfg.Namespace, Name: "dns-pod", Labels: map[string]string{"app": "fixture-dns"}},
		Status:     corev1.PodStatus{PodIP: "127.0.0.1"},
	}
	foreign := matching.DeepCopy()
	foreign.Namespace = "workload"
	foreign.Status.PodIP = "127.0.0.2"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: cfg.Namespace, Name: cfg.DNSServiceName},
		Spec:       corev1.ServiceSpec{ClusterIP: "192.0.2.53", Selector: matching.Labels},
	}
	chaos := &v1alpha1.DNSChaos{
		ObjectMeta: metav1.ObjectMeta{Namespace: "workload", Name: "experiment"},
		Spec:       v1alpha1.DNSChaosSpec{Action: v1alpha1.RandomAction, DomainNamePatterns: []string{"example.org"}},
	}
	ctx := t.Context()
	var listErr error
	kube := fake.NewClientBuilder().WithObjects(target, matching, foreign).WithInterceptorFuncs(interceptor.Funcs{
		List: func(ctx context.Context, c client.WithWatch, out client.ObjectList, opts ...client.ListOption) error {
			if listErr != nil {
				return listErr
			}
			return c.List(ctx, out, opts...)
		},
	}).Build()
	impl := NewImpl(kube, logr.Discard(), utils.NewContainerRecordDecoder(kube, &chaosdaemon.ClientBuilder{Reader: kube})).Impl
	records := []*v1alpha1.Record{{Id: "workload/target/app"}}
	phase, err := impl.Apply(ctx, 0, records, chaos)
	g.Expect(err).To(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	phase, err = impl.Recover(ctx, 0, records, chaos)
	g.Expect(err).To(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))
	g.Expect(daemon.requests).To(BeEmpty())

	g.Expect(kube.Create(ctx, service)).To(Succeed())
	listErr = errors.New("cannot list DNS pods")
	phase, err = impl.Apply(ctx, 0, records, chaos)
	g.Expect(err).To(MatchError(listErr))
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	phase, err = impl.Recover(ctx, 0, records, chaos)
	g.Expect(err).To(MatchError(listErr))
	g.Expect(phase).To(Equal(v1alpha1.Injected))
	listErr = nil

	g.Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))
	g.Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
	var applied *dnspb.SetDNSChaosRequest
	var canceled *dnspb.CancelDNSChaosRequest
	g.Expect(dns.applied).To(Receive(&applied))
	g.Expect(proto.Equal(applied, &dnspb.SetDNSChaosRequest{
		Name: "experiment", Action: "random", Pods: []*dnspb.Pod{{Name: "target", Namespace: "workload"}}, Patterns: []string{"example.org"},
	})).To(BeTrue())
	g.Expect(dns.canceled).To(Receive(&canceled))
	g.Expect(canceled.Name).To(Equal("experiment"))
	g.Expect(daemon.requests).To(HaveLen(2))
	g.Expect(proto.Equal(daemon.requests[0], &pb.SetDNSServerRequest{ContainerId: "containerd://fixture", DnsServer: "192.0.2.53", Enable: true, EnterNS: true})).To(BeTrue())
	g.Expect(proto.Equal(daemon.requests[1], &pb.SetDNSServerRequest{ContainerId: "containerd://fixture", EnterNS: true})).To(BeTrue())

	for _, ruleFailure := range []int32{1, 2} {
		dns.behavior.Store(ruleFailure)
		phase, err = impl.Apply(ctx, 0, records, chaos)
		g.Expect(err).To(HaveOccurred())
		g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		phase, err = impl.Recover(ctx, 0, records, chaos)
		g.Expect(err).To(HaveOccurred())
		g.Expect(phase).To(Equal(v1alpha1.Injected))
	}
	g.Expect(daemon.requests).To(HaveLen(2))
	dns.behavior.Store(0)
	daemon.err = errors.New("fixture daemon failed")
	phase, err = impl.Apply(ctx, 0, records, chaos)
	g.Expect(err).To(MatchError(daemon.err))
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	phase, err = impl.Recover(ctx, 0, records, chaos)
	g.Expect(err).To(MatchError(daemon.err))
	g.Expect(phase).To(Equal(v1alpha1.Injected))
	g.Expect(daemon.closed).To(Equal(12))

	vanished := []*v1alpha1.Record{{Id: "workload/deleted/app"}}
	phase, err = impl.Apply(ctx, 0, vanished, chaos)
	g.Expect(err).To(MatchError(utils.ErrContainerNotFound))
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	g.Expect(impl.Recover(ctx, 0, vanished, chaos)).To(Equal(v1alpha1.NotInjected))
}
