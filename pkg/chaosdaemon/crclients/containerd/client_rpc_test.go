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

package containerd

import (
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	containersapi "github.com/containerd/containerd/api/services/containers/v1"
	namespacesapi "github.com/containerd/containerd/api/services/namespaces/v1"
	tasksapi "github.com/containerd/containerd/api/services/tasks/v1"
	versionapi "github.com/containerd/containerd/api/services/version/v1"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/containerd/errdefs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type failingContainersServer struct {
	containersapi.UnimplementedContainersServer
	code codes.Code
}

type testNamespacesServer struct {
	namespacesapi.UnimplementedNamespacesServer
	code codes.Code
}

func (s testNamespacesServer) Get(_ context.Context, req *namespacesapi.GetNamespaceRequest) (*namespacesapi.GetNamespaceResponse, error) {
	if s.code != codes.OK {
		return nil, status.Error(s.code, "namespace unavailable")
	}
	return &namespacesapi.GetNamespaceResponse{Namespace: &namespacesapi.Namespace{Name: req.Name}}, nil
}

type testVersionServer struct {
	versionapi.UnimplementedVersionServer
}

func (testVersionServer) Version(context.Context, *emptypb.Empty) (*versionapi.VersionResponse, error) {
	return &versionapi.VersionResponse{Version: "v1.7.36"}, nil
}

func testSocketPath() string {
	GinkgoHelper()
	dir, err := os.MkdirTemp("", "ctrd-")
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(os.RemoveAll(dir)).To(Succeed()) })
	return filepath.Join(dir, "c.sock")
}

func (s failingContainersServer) Get(ctx context.Context, req *containersapi.GetContainerRequest) (*containersapi.GetContainerResponse, error) {
	defer GinkgoRecover()
	md, _ := metadata.FromIncomingContext(ctx)
	Expect(md.Get("containerd-namespace")).To(Equal([]string{"test-namespace"}))
	Expect(req.ID).To(Equal("missing"))
	return nil, status.Error(s.code, "container unavailable")
}

type operationContainersServer struct {
	containersapi.UnimplementedContainersServer
	containers []*containersapi.Container
	filter     string
	getErrors  map[string]codes.Code
}

func (s operationContainersServer) Get(_ context.Context, req *containersapi.GetContainerRequest) (*containersapi.GetContainerResponse, error) {
	defer GinkgoRecover()
	if code := s.getErrors[req.ID]; code != codes.OK {
		return nil, status.Error(code, "container unavailable")
	}
	for _, container := range s.containers {
		if container.ID == req.ID {
			return &containersapi.GetContainerResponse{Container: container}, nil
		}
	}
	Expect(s.containers).To(ContainElement(HaveField("ID", req.ID)), "unexpected container ID %q", req.ID)
	return nil, status.Error(codes.NotFound, "container not found")
}

func (s operationContainersServer) List(_ context.Context, req *containersapi.ListContainersRequest) (*containersapi.ListContainersResponse, error) {
	defer GinkgoRecover()
	Expect(req.Filters).To(Equal([]string{s.filter}))
	return &containersapi.ListContainersResponse{Containers: s.containers}, nil
}

type operationTasksServer struct {
	tasksapi.UnimplementedTasksServer
	containerID string
	pid         uint32
	getError    codes.Code
	killError   codes.Code
	kills       chan *tasksapi.KillRequest
}

func (s operationTasksServer) Get(_ context.Context, req *tasksapi.GetRequest) (*tasksapi.GetResponse, error) {
	defer GinkgoRecover()
	Expect(req.ContainerID).To(Equal(s.containerID))
	Expect(req.ExecID).To(BeEmpty())
	if s.getError != codes.OK {
		return nil, status.Error(s.getError, "task unavailable")
	}
	return &tasksapi.GetResponse{Process: &tasktypes.Process{ContainerID: s.containerID, ID: s.containerID, Pid: s.pid}}, nil
}

func (s operationTasksServer) Kill(_ context.Context, req *tasksapi.KillRequest) (*emptypb.Empty, error) {
	s.kills <- req
	if s.killError != codes.OK {
		return nil, status.Error(s.killError, "kill denied")
	}
	return &emptypb.Empty{}, nil
}

func newOperationClient(containers operationContainersServer, tasks operationTasksServer) (*Client, context.Context) {
	GinkgoHelper()
	address := testSocketPath()
	listener, err := net.Listen("unix", address)
	Expect(err).NotTo(HaveOccurred())
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		defer GinkgoRecover()
		if strings.HasPrefix(info.FullMethod, "/containerd.services.containers.") || strings.HasPrefix(info.FullMethod, "/containerd.services.tasks.") {
			md, _ := metadata.FromIncomingContext(ctx)
			Expect(md.Get("containerd-namespace")).To(Equal([]string{"test-namespace"}), "%s namespace", info.FullMethod)
		}
		return handler(ctx, req)
	}))
	containersapi.RegisterContainersServer(server, containers)
	tasksapi.RegisterTasksServer(server, tasks)
	namespacesapi.RegisterNamespacesServer(server, testNamespacesServer{})
	versionapi.RegisterVersionServer(server, testVersionServer{})
	go server.Serve(listener)
	DeferCleanup(server.Stop)
	client, err := New(address, WithDefaultNamespace("test-namespace"))
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(client.client.(io.Closer).Close()).To(Succeed()) })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	DeferCleanup(cancel)
	return client, ctx
}

var _ = Describe("containerd RPC", func() {
	DescribeTable("classifies container lookup errors",
		func(code codes.Code) {
			address := testSocketPath()
			listener, err := net.Listen("unix", address)
			Expect(err).NotTo(HaveOccurred())
			server := grpc.NewServer()
			containersapi.RegisterContainersServer(server, failingContainersServer{code: code})
			namespacesapi.RegisterNamespacesServer(server, testNamespacesServer{})
			versionapi.RegisterVersionServer(server, testVersionServer{})
			go server.Serve(listener)
			DeferCleanup(server.Stop)

			client, err := New(address, WithDefaultNamespace("test-namespace"))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { Expect(client.client.(io.Closer).Close()).To(Succeed()) })
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			DeferCleanup(cancel)
			_, err = client.GetPidFromContainerID(ctx, "containerd://missing")
			Expect(err).To(HaveOccurred())
			Expect(errdefs.IsNotFound(err)).To(Equal(code == codes.NotFound))
		},
		Entry("NotFound", codes.NotFound),
		Entry("PermissionDenied", codes.PermissionDenied),
	)

	It("rejects an unavailable socket", func() {
		client, err := New(testSocketPath())
		if err == nil {
			DeferCleanup(func() { Expect(client.client.(io.Closer).Close()).To(Succeed()) })
		}
		Expect(err).To(HaveOccurred())
	})

	It("rejects an unavailable namespace", func() {
		address := testSocketPath()
		listener, err := net.Listen("unix", address)
		Expect(err).NotTo(HaveOccurred())
		server := grpc.NewServer()
		versionapi.RegisterVersionServer(server, testVersionServer{})
		namespacesapi.RegisterNamespacesServer(server, testNamespacesServer{code: codes.NotFound})
		go server.Serve(listener)
		DeferCleanup(server.Stop)
		client, err := New(address, WithDefaultNamespace("missing-namespace"))
		if err == nil {
			DeferCleanup(func() { Expect(client.client.(io.Closer).Close()).To(Succeed()) })
		}
		Expect(err).To(HaveOccurred())
		Expect(errdefs.IsNotFound(err)).To(BeTrue())
	})

	It("waits for the socket to become available", func() {
		address := testSocketPath()
		server := grpc.NewServer()
		versionapi.RegisterVersionServer(server, testVersionServer{})
		DeferCleanup(server.Stop)
		started := make(chan error, 1)
		timer := time.AfterFunc(100*time.Millisecond, func() {
			listener, err := net.Listen("unix", address)
			started <- err
			if err == nil {
				go server.Serve(listener)
			}
		})
		DeferCleanup(func() { timer.Stop() })
		client, err := New(address)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(client.client.(io.Closer).Close()).To(Succeed()) })
		var startErr error
		Expect(started).To(Receive(&startErr), "constructor returned before the socket became available")
		Expect(startErr).NotTo(HaveOccurred())
	})

	It("returns the container task PID", func() {
		client, ctx := newOperationClient(operationContainersServer{
			containers: []*containersapi.Container{{ID: "app-a"}},
		}, operationTasksServer{containerID: "app-a", pid: 4321})
		pid, err := client.GetPidFromContainerID(ctx, "containerd://app-a")
		Expect(err).NotTo(HaveOccurred())
		Expect(pid).To(Equal(uint32(4321)))
	})

	DescribeTable("sends SIGKILL to the container task",
		func(code codes.Code) {
			kills := make(chan *tasksapi.KillRequest, 1)
			client, ctx := newOperationClient(operationContainersServer{
				containers: []*containersapi.Container{{ID: "app-a"}},
			}, operationTasksServer{containerID: "app-a", pid: 4321, killError: code, kills: kills})
			err := client.ContainerKillByContainerID(ctx, "containerd://app-a")
			if code == codes.OK {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(errdefs.IsPermissionDenied(err)).To(BeTrue())
			}
			var req *tasksapi.KillRequest
			Expect(kills).To(Receive(&req), "no kill request reached containerd")
			Expect(req.ContainerID).To(Equal("app-a"))
			Expect(req.Signal).To(Equal(uint32(syscall.SIGKILL)))
			Expect(req.ExecID).To(BeEmpty())
			Expect(req.All).To(BeFalse())
		},
		Entry("succeeds", codes.OK),
		Entry("preserves PermissionDenied", codes.PermissionDenied),
	)

	It("lists workload containers with prefixed IDs", func() {
		client, ctx := newOperationClient(operationContainersServer{
			containers: []*containersapi.Container{{ID: "app-a"}, {ID: "app-b"}},
			filter:     `labels."io.cri-containerd.kind"=="container"`,
		}, operationTasksServer{})
		ids, err := client.ListContainerIDs(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(ids).To(Equal([]string{"containerd://app-a", "containerd://app-b"}))
	})

	DescribeTable("selects the first readable sandbox for the pod",
		func(code codes.Code) {
			client, ctx := newOperationClient(operationContainersServer{
				containers: []*containersapi.Container{
					{ID: "unreadable", Labels: map[string]string{"io.kubernetes.pod.uid": "wanted-pod"}},
					{ID: "other-pod", Labels: map[string]string{"io.kubernetes.pod.uid": "other-pod"}},
					{ID: "no-pod-label"},
					{ID: "wanted-sandbox", Labels: map[string]string{"io.kubernetes.pod.uid": "wanted-pod"}},
					{ID: "later-sandbox", Labels: map[string]string{"io.kubernetes.pod.uid": "wanted-pod"}},
				},
				filter:    `labels."io.cri-containerd.kind"=="sandbox"`,
				getErrors: map[string]codes.Code{"unreadable": codes.PermissionDenied},
			}, operationTasksServer{containerID: "wanted-sandbox", pid: 9876, getError: code})
			pid, err := client.GetSandboxPidFromPodUID(ctx, "wanted-pod")
			if code == codes.OK {
				Expect(err).NotTo(HaveOccurred())
				Expect(pid).To(Equal(uint32(9876)))
			} else {
				Expect(pid).To(BeZero())
				Expect(errdefs.IsNotFound(err)).To(BeTrue())
			}
		},
		Entry("returns the task PID", codes.OK),
		Entry("preserves task NotFound", codes.NotFound),
	)
})
