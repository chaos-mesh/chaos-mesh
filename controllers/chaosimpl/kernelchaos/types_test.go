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
		name            string
		containerNames  []string
		statuses        []v1.ContainerStatus
		wantName        string
		wantID          string
		recordID        string
		deletePod       bool
		getErr          error
		daemonErr       error
		bpfErr          error
		wantErr         bool
		wantCause       error
		wantCode        codes.Code
		wantDaemon      int
		wantBPF         int
		probability     uint32
		times           uint32
		wantProbability float32
		wantTimes       uint32
	}{
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "selected container", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected",
			wantID: "containerd://selected-id", wantDaemon: 1, wantBPF: 1,
		},
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "reversed status order", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{selected, first}, wantName: "selected",
			wantID: "containerd://selected-id", wantDaemon: 1, wantBPF: 1,
		},
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "default first spec container", statuses: []v1.ContainerStatus{selected, first},
			wantName: "first", wantID: "containerd://first-id", wantDaemon: 1, wantBPF: 1,
		},
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "empty statuses", containerNames: []string{"selected"}, wantName: "selected",
			wantErr: true, wantCause: utils.ErrContainerNotFound,
		},
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "selected status missing", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first}, wantName: "selected",
			wantErr: true, wantCause: utils.ErrContainerNotFound,
		},
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "selected runtime ID empty", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, {Name: "selected"}}, wantName: "selected",
			wantErr: true, wantCause: utils.ErrContainerNotFound,
		},
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "malformed record", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected",
			recordID: "kernel-ci/target", wantErr: true,
		},
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "pod not found", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected", deletePod: true,
		},
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "pod forbidden", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected",
			getErr: forbidden, wantErr: true, wantCause: forbidden,
		},
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "daemon RPC error", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected", wantID: "containerd://selected-id",
			daemonErr: status.Error(codes.Unavailable, "PID lookup failed"),
			wantErr:   true, wantCode: codes.Unavailable, wantDaemon: 1,
		},
		{
			probability: 25, times: 7, wantProbability: 0.25, wantTimes: 7,
			name: "BPFKI RPC error", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected", wantID: "containerd://selected-id",
			bpfErr:  status.Error(codes.Internal, "kernel request failed"),
			wantErr: true, wantCode: codes.Internal, wantDaemon: 1, wantBPF: 1,
		},
		{
			name: "zero probability and times", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected",
			wantID: "containerd://selected-id", wantDaemon: 1, wantBPF: 1,
			probability: 0, times: 0, wantProbability: 0, wantTimes: 0,
		},
		{
			name: "100 percent probability and zero times", containerNames: []string{"selected"},
			statuses: []v1.ContainerStatus{first, selected}, wantName: "selected",
			wantID: "containerd://selected-id", wantDaemon: 1, wantBPF: 1,
			probability: 100, times: 0, wantProbability: 1, wantTimes: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := &kernelRPCCapture{daemonErr: tc.daemonErr, bpfErr: tc.bpfErr}
			daemonPort := serveKernelTestRPC(t, func(server *grpc.Server) { pb.RegisterChaosDaemonServer(server, calls) })
			bpfPort := serveKernelTestRPC(t, func(server *grpc.Server) { pbkernel.RegisterBPFKIServiceServer(server, calls) })
			previousConfig := controllerconfig.ControllerCfg
			controllerconfig.ControllerCfg = &config.ChaosControllerConfig{
				Namespace: "chaos-mesh", ClusterScoped: true,
				ChaosDaemonPort: daemonPort, BPFKIPort: bpfPort, MaxEvents: 20,
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
			node := &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node-a", UID: "node-a-uid"},
				Status:     v1.NodeStatus{NodeInfo: v1.NodeSystemInfo{BootID: kernelTestBootID}},
			}
			kubeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod, node, endpoints).Build()
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
						FailType: 1, Headers: []string{"linux/mmzone.h"}, Probability: tc.probability, Times: tc.times,
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
			records := []*v1alpha1.Record{{Id: targets[0].Id(), Phase: v1alpha1.NotInjected}}
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
			phase, err := impl.Apply(ctx, 0, records, chaos)
			if tc.wantErr {
				require.Equal(t, v1alpha1.NotInjected, phase)
				require.Error(t, err)
				if tc.wantCause != nil {
					require.ErrorIs(t, err, tc.wantCause)
				}
				if tc.wantCode != codes.OK {
					require.Equal(t, tc.wantCode, status.Code(err))
				}
			} else if tc.deletePod {
				require.NoError(t, err)
				require.Equal(t, v1alpha1.NotInjected, phase)
			} else {
				require.NoError(t, err)
				require.Equal(t, v1alpha1.KernelReserved, phase)
				require.Contains(t, chaos.Status.KernelFaults, records[0].Id)
				records[0].Phase = phase
				phase, err = impl.Apply(ctx, 0, records, chaos)
				require.NoError(t, err)
				require.Equal(t, v1alpha1.Injected, phase)
				records[0].Phase = phase
				phase, err = impl.Recover(ctx, 0, records, chaos)
				require.NoError(t, err)
				require.Equal(t, v1alpha1.NotInjected, phase)
				require.Empty(t, chaos.Status.KernelFaults)
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
				if tc.wantErr {
					require.Equal(t, []string{"Reserve"}, calls.bpfMethods)
				} else {
					require.Equal(t, []string{"Reserve", "Arm", "Release"}, calls.bpfMethods)
					require.True(t, proto.Equal(calls.references[0], calls.references[1]))
					require.Equal(t, kernelTestBootID, calls.references[1].NodeBootId)
				}
				callchain := []*pbkernel.FailKernRequestFrame{{Funcname: "alloc_pages", Parameters: "int order", Predicate: "order > 0"}}
				wantApply := &pbkernel.FailKernRequest{
					Pid: 4242, Ftype: pbkernel.FailKernRequest_FAILTYPE(1), Headers: []string{"linux/mmzone.h"},
					Callchain: callchain, Probability: tc.wantProbability, Times: tc.wantTimes,
				}
				require.True(t, proto.Equal(wantApply, calls.bpfRequests[0]), "Apply request: %v", calls.bpfRequests[0])
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
	references  []*pbkernel.KernelFaultReference
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

const kernelTestBootID = "11111111-2222-3333-4444-555555555555"

func kernelTestHandle() *pbkernel.KernelFaultHandle {
	return &pbkernel.KernelFaultHandle{
		Id: "0123456789abcdef0123456789abcdef",
		Session: &pbkernel.KernelFaultSession{
			Instance: "fedcba9876543210fedcba9876543210", BootId: kernelTestBootID,
			Pid: 300, StartTime: 10000, PidNamespace: 4026531836,
		},
	}
}

func (c *kernelRPCCapture) ReserveKernelFault(_ context.Context, request *pbkernel.ReserveKernelFaultRequest) (*pbkernel.KernelFaultHandle, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bpfRequests = append(c.bpfRequests, request.Fault)
	c.bpfMethods = append(c.bpfMethods, "Reserve")
	return kernelTestHandle(), c.bpfErr
}

func (c *kernelRPCCapture) ArmKernelFault(_ context.Context, request *pbkernel.KernelFaultReference) (*pbkernel.StatusResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.references = append(c.references, request)
	c.bpfMethods = append(c.bpfMethods, "Arm")
	return &pbkernel.StatusResponse{}, c.bpfErr
}

func (c *kernelRPCCapture) ReleaseKernelFault(_ context.Context, request *pbkernel.KernelFaultReference) (*pbkernel.ReleaseKernelFaultResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.references = append(c.references, request)
	c.bpfMethods = append(c.bpfMethods, "Release")
	return &pbkernel.ReleaseKernelFaultResponse{Outcome: pbkernel.ReleaseKernelFaultResponse_RELEASED}, c.bpfErr
}

func serveKernelTestRPC(t *testing.T, register func(*grpc.Server)) int {
	t.Helper()
	return serveKernelTestRPCAt(t, "127.0.0.1:0", register)
}

func serveKernelTestRPCAt(t *testing.T, address string, register func(*grpc.Server)) int {
	t.Helper()
	listener, err := net.Listen("tcp", address)
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
