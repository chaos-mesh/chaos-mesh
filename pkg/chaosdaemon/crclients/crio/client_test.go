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

package crio

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	runtimev1 "k8s.io/cri-api/pkg/apis/runtime/v1"

	crerrors "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/errors"
)

func TestCRIOClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CRI-O client")
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type runtimeStub struct {
	runtimev1.RuntimeServiceClient
	list   func(context.Context, *runtimev1.ListContainersRequest) (*runtimev1.ListContainersResponse, error)
	status func(context.Context, *runtimev1.ContainerStatusRequest) (*runtimev1.ContainerStatusResponse, error)
}

func (r runtimeStub) ListContainers(ctx context.Context, req *runtimev1.ListContainersRequest, _ ...grpc.CallOption) (*runtimev1.ListContainersResponse, error) {
	return r.list(ctx, req)
}

func (r runtimeStub) ContainerStatus(ctx context.Context, req *runtimev1.ContainerStatusRequest, _ ...grpc.CallOption) (*runtimev1.ContainerStatusResponse, error) {
	return r.status(ctx, req)
}

var _ = Describe("CRI-O container identity and inspection", func() {
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "experiment")
	const target = "cri-o://abc123"
	unavailable := errors.New("runtime unavailable")

	It("looks up the container without the runtime prefix and retains the request context", func() {
		client := Client{socketPath: "runtime.fixture", client: &http.Client{Transport: transportFunc(func(req *http.Request) (*http.Response, error) {
			Expect(req.Method).To(Equal(http.MethodGet))
			Expect(req.Host).To(Equal("crio"))
			Expect(req.URL.Path).To(Equal("/containers/abc123"))
			Expect(req.Context().Value(contextKey{})).To(Equal("experiment"))
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"pid":302}`))}, nil
		})}}
		Expect(client.GetPidFromContainerID(ctx, target)).To(Equal(uint32(302)))
	})

	DescribeTable("rejects unparseable inspections",
		func(body, message string, transportErr error) {
			client := Client{socketPath: "runtime.fixture", client: &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
				if transportErr != nil {
					return nil, transportErr
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			pid, err := client.GetPidFromContainerID(ctx, target)
			Expect(pid).To(BeZero())
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("transport error", "", unavailable.Error(), unavailable),
		Entry("malformed response", "x", "invalid character", nil),
		Entry("missing PID", "{}", "fail to get pid", nil),
		Entry("string PID", `{"pid":"302"}`, "fail to get pid", nil),
	)

	It("keeps runtime errors and does not inspect containers with a different runtime", func() {
		client := Client{runtimeClient: runtimeStub{
			list: func(context.Context, *runtimev1.ListContainersRequest) (*runtimev1.ListContainersResponse, error) {
				return nil, unavailable
			},
			status: func(ctx context.Context, req *runtimev1.ContainerStatusRequest) (*runtimev1.ContainerStatusResponse, error) {
				Expect(ctx.Value(contextKey{})).To(Equal("experiment"))
				Expect(req.ContainerId).To(Equal("abc123"))
				return nil, unavailable
			},
		}}
		Expect(client.ListContainerIDs(ctx)).Error().To(MatchError(unavailable))
		Expect(client.GetLabelsFromContainerID(ctx, target)).Error().To(MatchError(unavailable))
		Expect(client.GetLabelsFromContainerID(ctx, "containerd://abc123")).Error().To(MatchError(ContainSubstring("expected cri-o://")))
		Expect(client.GetPidFromContainerID(ctx, "short")).Error().To(MatchError(ContainSubstring("not a crio container id")))
		Expect(client.ContainerKillByContainerID(ctx, "short")).To(MatchError(ContainSubstring("not a crio container id")))
		Expect(client.GetSandboxPidFromPodUID(ctx, "pod")).Error().To(MatchError(crerrors.ErrNotSupported))
	})

	It("qualifies listed IDs and returns labels for the requested container", func() {
		labels := map[string]string{"io.kubernetes.pod.name": "fixture", "io.kubernetes.pod.namespace": "workload"}
		client := Client{runtimeClient: runtimeStub{
			list: func(ctx context.Context, req *runtimev1.ListContainersRequest) (*runtimev1.ListContainersResponse, error) {
				Expect(ctx.Value(contextKey{})).To(Equal("experiment"))
				Expect(req.Filter).To(BeNil())
				return &runtimev1.ListContainersResponse{Containers: []*runtimev1.Container{{Id: "abc123"}, {Id: "def456"}}}, nil
			},
			status: func(ctx context.Context, req *runtimev1.ContainerStatusRequest) (*runtimev1.ContainerStatusResponse, error) {
				Expect(ctx.Value(contextKey{})).To(Equal("experiment"))
				Expect(req.ContainerId).To(Equal("abc123"))
				return &runtimev1.ContainerStatusResponse{Status: &runtimev1.ContainerStatus{Labels: labels}}, nil
			},
		}}
		Expect(client.ListContainerIDs(ctx)).To(Equal([]string{target, "cri-o://def456"}))
		Expect(client.GetLabelsFromContainerID(ctx, target)).To(Equal(labels))
	})

	It("rejects an unrepresentable socket path before connecting to the runtime", func() {
		path := "/" + strings.Repeat("long", maxUnixSocketPathSize)
		Expect(New(path)).Error().To(MatchError(ContainSubstring("is too long")))
	})
})
