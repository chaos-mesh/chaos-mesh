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
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	containersapi "github.com/containerd/containerd/api/services/containers/v1"
	namespacesapi "github.com/containerd/containerd/api/services/namespaces/v1"
	tasksapi "github.com/containerd/containerd/api/services/tasks/v1"
	versionapi "github.com/containerd/containerd/api/services/version/v1"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/containerd/errdefs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type operationContainersServer struct {
	containersapi.UnimplementedContainersServer
	t          *testing.T
	containers []*containersapi.Container
	filter     string
	getErrors  map[string]codes.Code
}

func (s operationContainersServer) Get(_ context.Context, req *containersapi.GetContainerRequest) (*containersapi.GetContainerResponse, error) {
	if code := s.getErrors[req.ID]; code != codes.OK {
		return nil, status.Error(code, "container unavailable")
	}
	for _, container := range s.containers {
		if container.ID == req.ID {
			return &containersapi.GetContainerResponse{Container: container}, nil
		}
	}
	s.t.Errorf("unexpected container ID %q", req.ID)
	return nil, status.Error(codes.NotFound, "container not found")
}

func (s operationContainersServer) List(_ context.Context, req *containersapi.ListContainersRequest) (*containersapi.ListContainersResponse, error) {
	if !reflect.DeepEqual(req.Filters, []string{s.filter}) {
		s.t.Errorf("container filters = %v, want %q", req.Filters, s.filter)
	}
	return &containersapi.ListContainersResponse{Containers: s.containers}, nil
}

type operationTasksServer struct {
	tasksapi.UnimplementedTasksServer
	t           *testing.T
	containerID string
	pid         uint32
	getError    codes.Code
	killError   codes.Code
	kills       chan *tasksapi.KillRequest
}

func (s operationTasksServer) Get(_ context.Context, req *tasksapi.GetRequest) (*tasksapi.GetResponse, error) {
	if req.ContainerID != s.containerID || req.ExecID != "" {
		s.t.Errorf("task lookup = %v, want container %q and no exec ID", req, s.containerID)
	}
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

func newOperationClient(t *testing.T, containers operationContainersServer, tasks operationTasksServer) (*Client, context.Context) {
	t.Helper()
	address := testSocketPath(t)
	listener, err := net.Listen("unix", address)
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if strings.HasPrefix(info.FullMethod, "/containerd.services.containers.") || strings.HasPrefix(info.FullMethod, "/containerd.services.tasks.") {
			md, _ := metadata.FromIncomingContext(ctx)
			if got := md.Get("containerd-namespace"); !reflect.DeepEqual(got, []string{"test-namespace"}) {
				t.Errorf("%s namespace = %v, want test-namespace", info.FullMethod, got)
			}
		}
		return handler(ctx, req)
	}))
	containers.t, tasks.t = t, t
	containersapi.RegisterContainersServer(server, containers)
	tasksapi.RegisterTasksServer(server, tasks)
	namespacesapi.RegisterNamespacesServer(server, testNamespacesServer{})
	versionapi.RegisterVersionServer(server, testVersionServer{})
	go server.Serve(listener)
	t.Cleanup(server.Stop)
	client, err := New(address, WithDefaultNamespace("test-namespace"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.client.(io.Closer).Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return client, ctx
}

func TestContainerdRPCContainerPID(t *testing.T) {
	client, ctx := newOperationClient(t, operationContainersServer{
		containers: []*containersapi.Container{{ID: "app-a"}},
	}, operationTasksServer{containerID: "app-a", pid: 4321})
	pid, err := client.GetPidFromContainerID(ctx, "containerd://app-a")
	if err != nil || pid != 4321 {
		t.Fatalf("container PID = %d, error = %v, want 4321 and no error", pid, err)
	}
}

func TestContainerdRPCKill(t *testing.T) {
	for _, code := range []codes.Code{codes.OK, codes.PermissionDenied} {
		t.Run(code.String(), func(t *testing.T) {
			kills := make(chan *tasksapi.KillRequest, 1)
			client, ctx := newOperationClient(t, operationContainersServer{
				containers: []*containersapi.Container{{ID: "app-a"}},
			}, operationTasksServer{containerID: "app-a", pid: 4321, killError: code, kills: kills})
			err := client.ContainerKillByContainerID(ctx, "containerd://app-a")
			if code == codes.OK && err != nil {
				t.Fatalf("kill failed: %v", err)
			}
			if code == codes.PermissionDenied && !errdefs.IsPermissionDenied(err) {
				t.Fatalf("kill error = %v, want PermissionDenied", err)
			}
			select {
			case req := <-kills:
				if req.ContainerID != "app-a" || req.Signal != uint32(syscall.SIGKILL) || req.ExecID != "" || req.All {
					t.Fatalf("kill request = %v, want SIGKILL for container app-a", req)
				}
			default:
				t.Fatal("no kill request reached containerd")
			}
		})
	}
}

func TestContainerdRPCListContainerIDs(t *testing.T) {
	client, ctx := newOperationClient(t, operationContainersServer{
		containers: []*containersapi.Container{{ID: "app-a"}, {ID: "app-b"}},
		filter:     `labels."io.cri-containerd.kind"=="container"`,
	}, operationTasksServer{})
	ids, err := client.ListContainerIDs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"containerd://app-a", "containerd://app-b"}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("container IDs = %v, want %v", ids, want)
	}
}

func TestContainerdRPCSandboxSelection(t *testing.T) {
	for _, code := range []codes.Code{codes.OK, codes.NotFound} {
		t.Run(code.String(), func(t *testing.T) {
			client, ctx := newOperationClient(t, operationContainersServer{
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
				if err != nil || pid != 9876 {
					t.Fatalf("sandbox PID = %d, error = %v, want 9876 and no error", pid, err)
				}
			} else if pid != 0 || !errdefs.IsNotFound(err) {
				t.Fatalf("sandbox PID = %d, error = %v, want 0 and task NotFound", pid, err)
			}
		})
	}
}
