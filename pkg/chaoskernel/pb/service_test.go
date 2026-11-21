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

package bpfki

import (
	"context"
	"net"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
)

func TestBPFKIService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BPFKI RPC contract")
}

type rpcCase struct {
	method  string
	request proto.Message
	call    func(context.Context, BPFKIServiceClient) (*StatusResponse, error)
}

func bpfkiRPC[T proto.Message](method string, request T, invoke func(BPFKIServiceClient, context.Context, T, ...grpc.CallOption) (*StatusResponse, error)) rpcCase {
	return rpcCase{method: method, request: request, call: func(ctx context.Context, client BPFKIServiceClient) (*StatusResponse, error) {
		return invoke(client, ctx, request)
	}}
}

var _ = Describe("BPFKI service", func() {
	DescribeTable("transmits the typed request and the gRPC status",
		func(ctx SpecContext, example rpcCase) {
			listener := bufconn.Listen(64 * 1024)
			type receivedCall struct {
				method  string
				id      []string
				request proto.Message
			}
			received := make(chan receivedCall, 1)
			server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handle grpc.UnaryHandler) (any, error) {
				received <- receivedCall{info.FullMethod, metadata.ValueFromIncomingContext(ctx, "request-id"), req.(proto.Message)}
				if len(metadata.ValueFromIncomingContext(ctx, "fallback")) != 0 {
					return handle(ctx, req)
				}
				return &StatusResponse{Ret: 7, Msg: "fixture response"}, nil
			}))
			RegisterBPFKIServiceServer(server, &UnimplementedBPFKIServiceServer{})
			serving := make(chan error, 1)
			go func() { serving <- server.Serve(listener) }()
			DeferCleanup(func() {
				server.Stop()
				Eventually(serving).Should(Receive())
			})

			conn, err := grpc.NewClient("passthrough:///bpfki", grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
					return listener.DialContext(ctx)
				}))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(conn.Close)
			client := NewBPFKIServiceClient(conn)

			outgoing := metadata.AppendToOutgoingContext(ctx, "request-id", example.method)
			reply, err := example.call(outgoing, client)
			Expect(err).NotTo(HaveOccurred())
			Expect(reply.GetRet()).To(Equal(int32(7)))
			Expect(reply.GetMsg()).To(Equal("fixture response"))
			var call receivedCall
			Expect(received).To(Receive(&call))
			Expect(call.method).To(Equal("/bpfki.BPFKIService/" + example.method))
			Expect(call.id).To(Equal([]string{example.method}))
			Expect(proto.Equal(call.request, example.request)).To(BeTrue())

			reply, err = example.call(metadata.AppendToOutgoingContext(outgoing, "fallback", "true"), client)
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
			Expect(status.Convert(err).Message()).To(ContainSubstring(example.method))
			Expect(reply).To(BeNil())

			for _, method := range _BPFKIService_serviceDesc.Methods {
				if method.MethodName != example.method {
					continue
				}
				_, err = method.Handler(&UnimplementedBPFKIServiceServer{}, ctx, func(any) error {
					return status.Error(codes.InvalidArgument, "invalid request")
				}, nil)
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
				_, err = method.Handler(&UnimplementedBPFKIServiceServer{}, ctx, func(value any) error {
					proto.Merge(value.(proto.Message), example.request)
					return nil
				}, nil)
				Expect(status.Code(err)).To(Equal(codes.Unimplemented))
			}
		},
		Entry("SetTimeVal", bpfkiRPC("SetTimeVal", &BumpTimeRequest{Pid: 11, Second: -2, Subsecond: 5}, BPFKIServiceClient.SetTimeVal), SpecTimeout(5*time.Second)),
		Entry("RecoverTimeVal", bpfkiRPC("RecoverTimeVal", &BumpTimeRequest{Pid: 11, Tid: 12}, BPFKIServiceClient.RecoverTimeVal), SpecTimeout(5*time.Second)),
		Entry("SetTimeSpec", bpfkiRPC("SetTimeSpec", &BumpTimeRequest{Pid: 11, Subsecond: 10, Probability: 0.25}, BPFKIServiceClient.SetTimeSpec), SpecTimeout(5*time.Second)),
		Entry("RecoverTimeSpec", bpfkiRPC("RecoverTimeSpec", &BumpTimeRequest{Pid: 13}, BPFKIServiceClient.RecoverTimeSpec), SpecTimeout(5*time.Second)),
		Entry("FailMMOrBIO", bpfkiRPC("FailMMOrBIO", &FailKernRequest{Pid: 13, Ftype: FailKernRequest_PAGE, Times: 3, Callchain: []*FailKernRequestFrame{{Funcname: "fixture"}}}, BPFKIServiceClient.FailMMOrBIO), SpecTimeout(5*time.Second)),
		Entry("RecoverMMOrBIO", bpfkiRPC("RecoverMMOrBIO", &FailKernRequest{Pid: 13, Headers: []string{"fixture.h"}}, BPFKIServiceClient.RecoverMMOrBIO), SpecTimeout(5*time.Second)),
		Entry("FailSyscall", bpfkiRPC("FailSyscall", &FailSyscallRequest{Pid: 11, Methods: []string{"fixture"}, Err: 5, Probability: 0.5}, BPFKIServiceClient.FailSyscall), SpecTimeout(5*time.Second)),
		Entry("RecoverSyscall", bpfkiRPC("RecoverSyscall", &FailSyscallRequest{Pid: 11, Tid: 12, Methods: []string{"fixture"}}, BPFKIServiceClient.RecoverSyscall), SpecTimeout(5*time.Second)),
	)
})
