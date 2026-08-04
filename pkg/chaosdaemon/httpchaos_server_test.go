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

package chaosdaemon

import (
	"context"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

type httpChaosRuntimeClientStub struct {
	containerPID   uint32
	containerErr   error
	sandboxPID     uint32
	sandboxErr     error
	sandboxCalls   int
	receivedPodUID string
}

func (s *httpChaosRuntimeClientStub) GetPidFromContainerID(context.Context, string) (uint32, error) {
	return s.containerPID, s.containerErr
}

func (s *httpChaosRuntimeClientStub) GetSandboxPidFromPodUID(_ context.Context, podUID string) (uint32, error) {
	s.sandboxCalls++
	s.receivedPodUID = podUID
	return s.sandboxPID, s.sandboxErr
}

func (s *httpChaosRuntimeClientStub) ContainerKillByContainerID(context.Context, string) error {
	return nil
}

func (s *httpChaosRuntimeClientStub) FormatContainerID(context.Context, string) (string, error) {
	return "", nil
}

func (s *httpChaosRuntimeClientStub) ListContainerIDs(context.Context) ([]string, error) {
	return nil, nil
}

func (s *httpChaosRuntimeClientStub) GetLabelsFromContainerID(context.Context, string) (map[string]string, error) {
	return nil, nil
}

func TestGetPIDForHTTPChaos(t *testing.T) {
	t.Run("uses the application container PID when it is running", func(t *testing.T) {
		runtimeClient := &httpChaosRuntimeClientStub{containerPID: 9527}
		server := &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}

		pid, err := server.getPIDForHTTPChaos(context.Background(), "containerd://container-id", "pod-uid")
		if err != nil {
			t.Fatalf("get PID: %v", err)
		}
		if pid != 9527 {
			t.Fatalf("expected PID 9527, got %d", pid)
		}
		if runtimeClient.sandboxCalls != 0 {
			t.Fatalf("expected no sandbox lookup, got %d", runtimeClient.sandboxCalls)
		}
	})

	t.Run("falls back to the sandbox PID when the application container is not running", func(t *testing.T) {
		runtimeClient := &httpChaosRuntimeClientStub{
			containerErr: errors.New("container not found"),
			sandboxPID:   9527,
		}
		server := &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}

		pid, err := server.getPIDForHTTPChaos(context.Background(), "containerd://container-id", "pod-uid")
		if err != nil {
			t.Fatalf("get PID: %v", err)
		}
		if pid != 9527 {
			t.Fatalf("expected PID 9527, got %d", pid)
		}
		if runtimeClient.receivedPodUID != "pod-uid" {
			t.Fatalf("expected pod UID pod-uid, got %q", runtimeClient.receivedPodUID)
		}
	})

	t.Run("returns an error when neither application nor sandbox PID is available", func(t *testing.T) {
		runtimeClient := &httpChaosRuntimeClientStub{
			containerErr: errors.New("container not found"),
			sandboxErr:   errors.New("sandbox container not found"),
		}
		server := &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}

		_, err := server.getPIDForHTTPChaos(context.Background(), "containerd://container-id", "pod-uid")
		if err == nil {
			t.Fatal("expected PID lookup to fail")
		}
		if !errors.Is(err, runtimeClient.sandboxErr) {
			t.Fatalf("expected sandbox lookup error, got %v", err)
		}
		if !containsAll(err.Error(), "sandbox container not found", "container not found") {
			t.Fatalf("expected both lookup failures in error, got %v", err)
		}
	})
}

func containsAll(value string, substrings ...string) bool {
	for _, substring := range substrings {
		if !strings.Contains(value, substring) {
			return false
		}
	}
	return true
}
