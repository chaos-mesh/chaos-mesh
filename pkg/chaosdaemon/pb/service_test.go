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

package pb

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestChaosDaemonRPC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaos Daemon wire contract")
}

type daemonRPCCase struct {
	name              string
	request, response proto.Message
	call              func(context.Context, ChaosDaemonClient, ...grpc.CallOption) (proto.Message, error)
}

func daemonRPC[Request, Response proto.Message](
	name string,
	request Request,
	response Response,
	invoke func(ChaosDaemonClient, context.Context, Request, ...grpc.CallOption) (Response, error),
) daemonRPCCase {
	return daemonRPCCase{name: name, request: request, response: response,
		call: func(ctx context.Context, client ChaosDaemonClient, opts ...grpc.CallOption) (proto.Message, error) {
			return invoke(client, ctx, request, opts...)
		},
	}
}

func wireExamples() []daemonRPCCase {
	empty := &emptypb.Empty{}
	const target = "containerd://fixture"
	const pod = "fixture-pod"
	return []daemonRPCCase{
		daemonRPC("SetTcs", &TcsRequest{ContainerId: target, PodUid: pod, EnterNS: true, Tcs: []*Tc{
			{Device: "eth0", Type: Tc_NETEM, Netem: &Netem{Time: "15ms", Handle: &TcHandle{Major: 1}}},
			{Device: "eth1", Type: Tc_BANDWIDTH, Tbf: &Tbf{Rate: "1mbit", Limit: 128, Buffer: 64}},
		}}, empty, ChaosDaemonClient.SetTcs),
		daemonRPC("FlushIPSets", &IPSetsRequest{ContainerId: target, PodUid: pod, EnterNS: true, Ipsets: []*IPSet{
			{Name: "example", CidrAndPorts: []*CidrAndPort{{Cidr: "192.0.2.0/24", Port: 8080}}},
		}}, empty, ChaosDaemonClient.FlushIPSets),
		daemonRPC("SetIptablesChains", &IptablesChainsRequest{ContainerId: target, PodUid: pod, EnterNS: true, Chains: []*Chain{
			{Name: "fixture", Direction: Chain_OUTPUT, Ipsets: []string{"example"}, Target: "DROP", FallbackTarget: "RETURN"},
		}}, empty, ChaosDaemonClient.SetIptablesChains),
		daemonRPC("SetTimeOffset", &TimeRequest{ContainerId: target, Sec: -2, Nsec: 300, ClkIdsMask: 2, Uid: "timer", PodContainerName: "app"}, empty, ChaosDaemonClient.SetTimeOffset),
		daemonRPC("RecoverTimeOffset", &TimeRequest{ContainerId: target, Uid: "timer", PodContainerName: "app"}, empty, ChaosDaemonClient.RecoverTimeOffset),
		daemonRPC("ContainerKill", &ContainerRequest{ContainerId: target, Action: &ContainerAction{Action: ContainerAction_KILL}}, empty, ChaosDaemonClient.ContainerKill),
		daemonRPC("ContainerGetPid", &ContainerRequest{ContainerId: target, Action: &ContainerAction{Action: ContainerAction_GETPID}}, &ContainerResponse{Pid: 42}, ChaosDaemonClient.ContainerGetPid),
		daemonRPC("ExecStressors", &ExecStressRequest{Target: target, Scope: ExecStressRequest_POD, CpuStressors: "fixture cpu", MemoryStressors: "fixture memory", OomScoreAdj: 10, EnterNS: true},
			&ExecStressResponse{CpuInstance: "20", CpuStartTime: 25, CpuInstanceUid: "cpu", MemoryInstance: "81", MemoryStartTime: 27, MemoryInstanceUid: "memory"}, ChaosDaemonClient.ExecStressors),
		daemonRPC("CancelStressors", &CancelStressRequest{CpuInstance: "20", CpuStartTime: 25, CpuInstanceUid: "cpu", MemoryInstance: "81", MemoryStartTime: 27, MemoryInstanceUid: "memory"}, empty, ChaosDaemonClient.CancelStressors),
		daemonRPC("ApplyIOChaos", &ApplyIOChaosRequest{ContainerId: target, Actions: "[]", Volume: "/fixture", Instance: 10, StartTime: 20, InstanceUid: "io", EnterNS: true},
			&ApplyIOChaosResponse{Instance: 11, StartTime: 40, InstanceUid: "new-io"}, ChaosDaemonClient.ApplyIOChaos),
		daemonRPC("ApplyHttpChaos", &ApplyHttpChaosRequest{ContainerId: target, PodUid: pod, Rules: "[]", Tls: "{}", ProxyPorts: []uint32{8080, 9443}, InstanceUid: "proxy", EnterNS: true},
			&ApplyHttpChaosResponse{Instance: 11, StartTime: 41, StatusCode: 200, InstanceUid: "new-proxy"}, ChaosDaemonClient.ApplyHttpChaos),
		daemonRPC("ApplyBlockChaos", &ApplyBlockChaosRequest{ContainerId: target, VolumePath: "/fixture", Delay: &BlockDelaySpec{Delay: 15, Correlation: 0.2, Jitter: 3}, EnterNS: true},
			&ApplyBlockChaosResponse{InjectionId: 17}, ChaosDaemonClient.ApplyBlockChaos),
		daemonRPC("RecoverBlockChaos", &RecoverBlockChaosRequest{InjectionId: 17}, empty, ChaosDaemonClient.RecoverBlockChaos),
		daemonRPC("SetDNSServer", &SetDNSServerRequest{ContainerId: target, DnsServer: "192.0.2.12", Enable: true, EnterNS: true}, empty, ChaosDaemonClient.SetDNSServer),
		daemonRPC("InstallJVMRules", &InstallJVMRulesRequest{ContainerId: target, Rule: "fixture", Port: 9091, EnterNS: true}, empty, ChaosDaemonClient.InstallJVMRules),
		daemonRPC("UninstallJVMRules", &UninstallJVMRulesRequest{ContainerId: target, Rule: "fixture", Port: 9091, EnterNS: true}, empty, ChaosDaemonClient.UninstallJVMRules),
	}
}

type rpcConnection func(context.Context, string, any, any, ...grpc.CallOption) error

func (c rpcConnection) Invoke(ctx context.Context, method string, req, reply any, opts ...grpc.CallOption) error {
	return c(ctx, method, req, reply, opts...)
}

func (rpcConnection) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, status.Error(codes.Unimplemented, "unexpected stream")
}

var _ = Describe("Chaos Daemon RPC client", func() {
	It("passes each request, its context and call options to the mocked connection", func() {
		for _, example := range wireExamples() {
			By(example.name)
			ctx := metadata.AppendToOutgoingContext(context.Background(), "experiment", example.name)
			option := grpc.WaitForReady(true)
			payload, err := proto.Marshal(example.response)
			Expect(err).NotTo(HaveOccurred())

			var connectionErr error
			calls := 0
			conn := rpcConnection(func(gotCtx context.Context, method string, req, reply any, opts ...grpc.CallOption) error {
				calls++
				Expect(gotCtx).To(Equal(ctx))
				Expect(method).To(Equal("/pb.ChaosDaemon/" + example.name))
				Expect(req).To(BeIdenticalTo(example.request))
				Expect(opts).To(ConsistOf(option))
				if connectionErr != nil {
					return connectionErr
				}
				Expect(reply).To(BeAssignableToTypeOf(example.response))
				return proto.Unmarshal(payload, reply.(proto.Message))
			})
			client := NewChaosDaemonClient(conn)
			response, err := example.call(ctx, client, option)
			Expect(err).NotTo(HaveOccurred())
			Expect(proto.Equal(response, example.response)).To(BeTrue())

			connectionErr = status.Error(codes.Unavailable, "connection unavailable")
			response, err = example.call(ctx, client, option)
			Expect(response).To(BeNil())
			Expect(err).To(MatchError(connectionErr))
			Expect(calls).To(Equal(2))
		}
	})
})

var _ = Describe("Chaos Daemon RPC handlers", func() {
	It("registers the expected unary methods and dispatches each through the interceptor", func() {
		examples := wireExamples()
		service := &UnimplementedChaosDaemonServer{}
		registry := grpc.NewServer()
		RegisterChaosDaemonServer(registry, service)
		info, registered := registry.GetServiceInfo()["pb.ChaosDaemon"]
		Expect(registered).To(BeTrue())
		Expect(info.Methods).To(HaveLen(len(examples)))
		handlers := make(map[string]grpc.MethodDesc, len(info.Methods))
		for _, method := range _ChaosDaemon_serviceDesc.Methods {
			handlers[method.MethodName] = method
		}

		for _, example := range examples {
			By(example.name)
			method, found := handlers[example.name]
			Expect(found).To(BeTrue())
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("experiment", example.name))
			payload, err := proto.Marshal(example.request)
			Expect(err).NotTo(HaveOccurred())
			decode := func(value any) error {
				Expect(value).To(BeAssignableToTypeOf(example.request))
				return proto.Unmarshal(payload, value.(proto.Message))
			}

			reply, err := method.Handler(service, ctx, decode, func(gotCtx context.Context, req any, info *grpc.UnaryServerInfo, handle grpc.UnaryHandler) (any, error) {
				Expect(gotCtx).To(Equal(ctx))
				Expect(info.FullMethod).To(Equal("/pb.ChaosDaemon/" + example.name))
				Expect(info.Server).To(BeIdenticalTo(service))
				Expect(proto.Equal(req.(proto.Message), example.request)).To(BeTrue())
				return handle(gotCtx, req)
			})
			Expect(reply).To(BeNil())
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
			Expect(status.Convert(err).Message()).To(ContainSubstring(example.name))

			reply, err = method.Handler(service, ctx, decode, nil)
			Expect(reply).To(BeNil())
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
			Expect(status.Convert(err).Message()).To(ContainSubstring(example.name))

			decodeErr := status.Error(codes.InvalidArgument, "invalid request")
			reply, err = method.Handler(service, ctx, func(any) error { return decodeErr },
				func(context.Context, any, *grpc.UnaryServerInfo, grpc.UnaryHandler) (any, error) {
					Fail("interceptor received an undecodable request")
					return nil, nil
				})
			Expect(reply).To(BeNil())
			Expect(err).To(MatchError(decodeErr))
		}
	})
})
