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
	"reflect"
	"testing"
	"time"

	containersapi "github.com/containerd/containerd/api/services/containers/v1"
	namespacesapi "github.com/containerd/containerd/api/services/namespaces/v1"
	versionapi "github.com/containerd/containerd/api/services/version/v1"
	"github.com/containerd/errdefs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type failingContainersServer struct {
	containersapi.UnimplementedContainersServer
	t    *testing.T
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

func testSocketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "ctrd-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, "c.sock")
}

func (s failingContainersServer) Get(ctx context.Context, req *containersapi.GetContainerRequest) (*containersapi.GetContainerResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if got := md.Get("containerd-namespace"); !reflect.DeepEqual(got, []string{"test-namespace"}) {
		s.t.Errorf("namespace = %v, want test-namespace", got)
	}
	if req.ID != "missing" {
		s.t.Errorf("container ID = %q, want missing", req.ID)
	}
	return nil, status.Error(s.code, "container unavailable")
}

func TestContainerdRPCErrorClassification(t *testing.T) {
	for _, code := range []codes.Code{codes.NotFound, codes.PermissionDenied} {
		t.Run(code.String(), func(t *testing.T) {
			address := testSocketPath(t)
			listener, err := net.Listen("unix", address)
			if err != nil {
				t.Fatal(err)
			}
			server := grpc.NewServer()
			containersapi.RegisterContainersServer(server, failingContainersServer{t: t, code: code})
			namespacesapi.RegisterNamespacesServer(server, testNamespacesServer{})
			versionapi.RegisterVersionServer(server, testVersionServer{})
			go server.Serve(listener)
			defer server.Stop()

			client, err := New(address, WithDefaultNamespace("test-namespace"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.client.(io.Closer).Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err = client.GetPidFromContainerID(ctx, "containerd://missing")
			if err == nil {
				t.Fatal("expected the container lookup to fail")
			}
			if got, want := errdefs.IsNotFound(err), code == codes.NotFound; got != want {
				t.Fatalf("IsNotFound(%v) = %v, want %v", err, got, want)
			}
		})
	}
}

func TestContainerdRejectsUnavailableSocket(t *testing.T) {
	client, err := New(testSocketPath(t))
	if err == nil {
		client.client.(io.Closer).Close()
		t.Fatal("constructor accepted an unavailable containerd socket")
	}
}

func TestContainerdRejectsUnavailableNamespace(t *testing.T) {
	address := testSocketPath(t)
	listener, err := net.Listen("unix", address)
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	versionapi.RegisterVersionServer(server, testVersionServer{})
	namespacesapi.RegisterNamespacesServer(server, testNamespacesServer{code: codes.NotFound})
	go server.Serve(listener)
	defer server.Stop()
	client, err := New(address, WithDefaultNamespace("missing-namespace"))
	if err == nil {
		client.client.(io.Closer).Close()
		t.Fatal("constructor accepted an unavailable containerd namespace")
	}
	if !errdefs.IsNotFound(err) {
		t.Fatalf("namespace error = %v, want NotFound", err)
	}
}

func TestContainerdWaitsForSocket(t *testing.T) {
	address := testSocketPath(t)
	server := grpc.NewServer()
	versionapi.RegisterVersionServer(server, testVersionServer{})
	defer server.Stop()
	started := make(chan error, 1)
	timer := time.AfterFunc(100*time.Millisecond, func() {
		listener, err := net.Listen("unix", address)
		started <- err
		if err == nil {
			go server.Serve(listener)
		}
	})
	defer timer.Stop()
	client, err := New(address)
	if err != nil {
		t.Fatal(err)
	}
	defer client.client.(io.Closer).Close()
	select {
	case err := <-started:
		if err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatal("constructor returned before the socket became available")
	}
}
