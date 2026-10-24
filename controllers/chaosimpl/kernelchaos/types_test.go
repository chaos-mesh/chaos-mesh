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

package kernelchaos_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/kernelchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	controllerconfig "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	pbkernel "github.com/chaos-mesh/chaos-mesh/pkg/chaoskernel/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
	containerselector "github.com/chaos-mesh/chaos-mesh/pkg/selector/container"
	podselector "github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

func TestKernelChaosContainerTargets(t *testing.T) {
	first := v1.ContainerStatus{Name: "first", ContainerID: "containerd://first-id"}
	selected := v1.ContainerStatus{Name: "selected", ContainerID: "containerd://selected-id"}
	forbidden := apierrors.NewForbidden(schema.GroupResource{Resource: "pods"}, "target", errors.New("denied"))

	for _, tc := range []struct {
		name           string
		containerNames []string
		statuses       []v1.ContainerStatus
		wantName       string
		wantID         string
		recordID       string
		deletePod      bool
		getErr         error
		daemonErr      error
		bpfErr         error
		wantErr        bool
		wantCause      error
		wantCode       codes.Code
		wantDaemon     int
		wantBPF        int
	}{
		{
			name: "selected container", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected",
			wantID: "containerd://selected-id", wantDaemon: 2, wantBPF: 2,
		},
		{
			name: "reversed status order", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{selected, first}, wantName: "selected",
			wantID: "containerd://selected-id", wantDaemon: 2, wantBPF: 2,
		},
		{
			name: "default first spec container", statuses: []v1.ContainerStatus{selected, first},
			wantName: "first", wantID: "containerd://first-id", wantDaemon: 2, wantBPF: 2,
		},
		{
			name: "empty statuses", containerNames: []string{"selected"}, wantName: "selected",
			wantErr: true, wantCause: utils.ErrContainerNotFound,
		},
		{
			name: "selected status missing", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first}, wantName: "selected",
			wantErr: true, wantCause: utils.ErrContainerNotFound,
		},
		{
			name: "selected runtime ID empty", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, {Name: "selected"}}, wantName: "selected",
			wantErr: true, wantCause: utils.ErrContainerNotFound,
		},
		{
			name: "malformed record", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected",
			recordID: "kernel-ci/target", wantErr: true,
		},
		{
			name: "pod not found", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected", deletePod: true,
		},
		{
			name: "pod forbidden", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected",
			getErr: forbidden, wantErr: true, wantCause: forbidden,
		},
		{
			name: "daemon RPC error", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected", wantID: "containerd://selected-id",
			daemonErr: status.Error(codes.Unavailable, "PID lookup failed"),
			wantErr:   true, wantCode: codes.Unavailable, wantDaemon: 2,
		},
		{
			name: "BPFKI RPC error", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected", wantID: "containerd://selected-id",
			bpfErr:  status.Error(codes.Internal, "kernel request failed"),
			wantErr: true, wantCode: codes.Internal, wantDaemon: 2, wantBPF: 2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := &kernelRPCCapture{daemonErr: tc.daemonErr, bpfErr: tc.bpfErr}
			daemonPort := serveKernelTestRPC(t, func(server *grpc.Server) { pb.RegisterChaosDaemonServer(server, calls) })
			bpfPort := serveKernelTestRPC(t, func(server *grpc.Server) { pbkernel.RegisterBPFKIServiceServer(server, calls) })
			previousConfig := controllerconfig.ControllerCfg
			controllerconfig.ControllerCfg = &config.ChaosControllerConfig{
				Namespace: "chaos-mesh", ClusterScoped: true,
				ChaosDaemonPort: daemonPort, BPFKIPort: bpfPort,
				TLSConfig: config.TLSConfig{ChaosMeshCACert: ""},
			}
			t.Cleanup(func() { controllerconfig.ControllerCfg = previousConfig })

			scheme := runtime.NewScheme()
			require.NoError(t, v1.AddToScheme(scheme))
			require.NoError(t, discoveryv1.AddToScheme(scheme))
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Namespace: "kernel-ci", Name: "target", UID: "target-uid"},
				Spec: v1.PodSpec{
					NodeName: "node-a", Containers: []v1.Container{{Name: "first"}, {Name: "selected"}},
				},
				Status: v1.PodStatus{Phase: v1.PodRunning, ContainerStatuses: tc.statuses},
			}
			endpoints := &discoveryv1.EndpointSlice{
				ObjectMeta:  metav1.ObjectMeta{Namespace: "chaos-mesh", Name: "chaos-daemon-test"},
				AddressType: discoveryv1.AddressTypeIPv4,
				Endpoints:   []discoveryv1.Endpoint{{NodeName: &pod.Spec.NodeName, Addresses: []string{"127.0.0.1"}}},
			}
			kubeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod, endpoints).Build()
			chaos := &v1alpha1.KernelChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: "kernel-ci", Name: "kernel-test"},
				Spec: v1alpha1.KernelChaosSpec{
					ContainerSelector: v1alpha1.ContainerSelector{
						PodSelector: v1alpha1.PodSelector{
							Mode:     v1alpha1.AllMode,
							Selector: v1alpha1.PodSelectorSpec{Pods: map[string][]string{"kernel-ci": {"target"}}},
						},
						ContainerNames: tc.containerNames,
					},
					FailKernRequest: v1alpha1.FailKernRequest{
						FailType: 1, Headers: []string{"linux/mmzone.h"}, Probability: 25, Times: 7,
						Callchain: []v1alpha1.Frame{{Funcname: "alloc_pages", Parameters: "int order", Predicate: "order > 0"}},
					},
				},
			}
			dispatcher := selector.New(selector.Params{
				PodSelector:       podselector.New(podselector.Params{Client: kubeClient, Reader: kubeClient}),
				ContainerSelector: containerselector.New(containerselector.Params{Client: kubeClient, Reader: kubeClient}),
			})
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			targets, err := dispatcher.Select(ctx, chaos.GetSelectorSpecs()["."])
			require.NoError(t, err)
			require.Len(t, targets, 1)
			require.Equal(t, "kernel-ci/target/"+tc.wantName, targets[0].Id())
			records := []*v1alpha1.Record{{Id: targets[0].Id()}}
			if tc.recordID != "" {
				records[0].Id = tc.recordID
			}
			if tc.deletePod {
				require.NoError(t, kubeClient.Delete(ctx, pod))
			}
			var implClient client.Client = kubeClient
			if tc.getErr != nil {
				implClient = &podGetErrorClient{Client: kubeClient, err: tc.getErr}
			}
			impl := kernelchaos.NewImpl(implClient, logr.Discard(), &chaosdaemon.ClientBuilder{Reader: kubeClient}).Impl
			for _, operation := range []struct {
				name string
				run  func(context.Context, int, []*v1alpha1.Record, v1alpha1.InnerObject) (v1alpha1.Phase, error)
			}{
				{name: "Apply", run: impl.Apply},
				{name: "Recover", run: impl.Recover},
			} {
				phase, err := operation.run(ctx, 0, records, chaos)
				wantPhase := v1alpha1.NotInjected
				if (operation.name == "Apply" && !tc.wantErr && !tc.deletePod) || (operation.name == "Recover" && tc.wantErr) {
					wantPhase = v1alpha1.Injected
				}
				require.Equal(t, wantPhase, phase, operation.name)
				if tc.wantErr {
					require.Error(t, err, operation.name)
					if tc.wantCause != nil {
						require.ErrorIs(t, err, tc.wantCause, operation.name)
					}
					if tc.wantCode != codes.OK {
						require.Equal(t, tc.wantCode, status.Code(err), operation.name)
					}
				} else {
					require.NoError(t, err, operation.name)
				}
			}

			calls.mu.Lock()
			defer calls.mu.Unlock()
			require.Len(t, calls.pidRequests, tc.wantDaemon)
			require.Len(t, calls.bpfRequests, tc.wantBPF)
			for _, request := range calls.pidRequests {
				require.Equal(t, tc.wantID, request.ContainerId)
				require.NotNil(t, request.Action)
				require.Equal(t, pb.ContainerAction_GETPID, request.Action.Action)
			}
			if tc.wantBPF > 0 {
				require.Equal(t, []string{"Apply", "Recover"}, calls.bpfMethods)
				callchain := []*pbkernel.FailKernRequestFrame{{Funcname: "alloc_pages", Parameters: "int order", Predicate: "order > 0"}}
				wantApply := &pbkernel.FailKernRequest{
					Pid: 4242, Ftype: pbkernel.FailKernRequest_FAILTYPE(1), Headers: []string{"linux/mmzone.h"},
					Callchain: callchain, Probability: 0.25, Times: 7,
				}
				wantRecover := &pbkernel.FailKernRequest{Pid: 4242, Callchain: callchain}
				require.True(t, proto.Equal(wantApply, calls.bpfRequests[0]), "Apply request: %v", calls.bpfRequests[0])
				require.True(t, proto.Equal(wantRecover, calls.bpfRequests[1]), "Recover request: %v", calls.bpfRequests[1])
			}
		})
	}
}

type podGetErrorClient struct {
	client.Client
	err error
}

func (c *podGetErrorClient) Get(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	return c.err
}

type kernelRPCCapture struct {
	pb.UnimplementedChaosDaemonServer
	pbkernel.UnimplementedBPFKIServiceServer
	mu          sync.Mutex
	pidRequests []*pb.ContainerRequest
	bpfRequests []*pbkernel.FailKernRequest
	bpfMethods  []string
	daemonErr   error
	bpfErr      error
}

func (c *kernelRPCCapture) ContainerGetPid(_ context.Context, request *pb.ContainerRequest) (*pb.ContainerResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pidRequests = append(c.pidRequests, request)
	return &pb.ContainerResponse{Pid: 4242}, c.daemonErr
}

func (c *kernelRPCCapture) FailMMOrBIO(_ context.Context, request *pbkernel.FailKernRequest) (*pbkernel.StatusResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bpfRequests = append(c.bpfRequests, request)
	c.bpfMethods = append(c.bpfMethods, "Apply")
	return &pbkernel.StatusResponse{}, c.bpfErr
}

func (c *kernelRPCCapture) RecoverMMOrBIO(_ context.Context, request *pbkernel.FailKernRequest) (*pbkernel.StatusResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bpfRequests = append(c.bpfRequests, request)
	c.bpfMethods = append(c.bpfMethods, "Recover")
	return &pbkernel.StatusResponse{}, c.bpfErr
}

func serveKernelTestRPC(t *testing.T, register func(*grpc.Server)) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer()
	register(server)
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Error("gRPC server did not stop")
		}
	})
	return listener.Addr().(*net.TCPAddr).Port
}
