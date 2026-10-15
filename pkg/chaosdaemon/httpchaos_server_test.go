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

package chaosdaemon

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	containerderrdefs "github.com/containerd/containerd/errdefs"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tproxyconfig"
)

type httpChaosRuntimeClientStub struct {
	containerPID   uint32
	containerErr   error
	sandboxPID     uint32
	sandboxErr     error
	containerCalls int
	sandboxCalls   int
	receivedPodUID string
}

func (s *httpChaosRuntimeClientStub) GetPidFromContainerID(context.Context, string) (uint32, error) {
	s.containerCalls++
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

		pid, identifier, err := server.getPIDForHTTPChaos(context.Background(), "containerd://container-id", "pod-uid")
		if err != nil {
			t.Fatalf("get PID: %v", err)
		}
		if pid != 9527 {
			t.Fatalf("expected PID 9527, got %d", pid)
		}
		if identifier != "containerd://container-id" {
			t.Fatalf("expected container identifier, got %q", identifier)
		}
		if runtimeClient.sandboxCalls != 0 {
			t.Fatalf("expected no sandbox lookup, got %d", runtimeClient.sandboxCalls)
		}
	})

	t.Run("falls back to the sandbox PID when the application container is not running", func(t *testing.T) {
		runtimeClient := &httpChaosRuntimeClientStub{
			containerErr: errors.Wrap(containerderrdefs.ErrNotFound, "container not found"),
			sandboxPID:   9527,
		}
		server := &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}

		pid, identifier, err := server.getPIDForHTTPChaos(context.Background(), "containerd://container-id", "pod-uid")
		if err != nil {
			t.Fatalf("get PID: %v", err)
		}
		if pid != 9527 {
			t.Fatalf("expected PID 9527, got %d", pid)
		}
		if identifier != "pod-pod-uid" {
			t.Fatalf("expected pod identifier, got %q", identifier)
		}
		if runtimeClient.receivedPodUID != "pod-uid" {
			t.Fatalf("expected pod UID pod-uid, got %q", runtimeClient.receivedPodUID)
		}
	})

	t.Run("returns an error when neither application nor sandbox PID is available", func(t *testing.T) {
		runtimeClient := &httpChaosRuntimeClientStub{
			containerErr: errors.Wrap(containerderrdefs.ErrNotFound, "container not found"),
			sandboxErr:   errors.New("sandbox container not found"),
		}
		server := &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}

		_, _, err := server.getPIDForHTTPChaos(context.Background(), "containerd://container-id", "pod-uid")
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

	t.Run("rejects an empty container ID without looking up the sandbox", func(t *testing.T) {
		runtimeClient := &httpChaosRuntimeClientStub{sandboxPID: 9527}
		server := &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}

		_, _, err := server.getPIDForHTTPChaos(context.Background(), "", "pod-uid")
		if err == nil || !strings.Contains(err.Error(), "container ID is empty") {
			t.Fatalf("expected empty container ID error, got %v", err)
		}
		if runtimeClient.containerCalls != 0 {
			t.Fatalf("expected no container lookup, got %d", runtimeClient.containerCalls)
		}
		if runtimeClient.sandboxCalls != 0 {
			t.Fatalf("expected no sandbox lookup, got %d", runtimeClient.sandboxCalls)
		}
	})

	t.Run("does not fall back for an invalid container ID", func(t *testing.T) {
		runtimeClient := &httpChaosRuntimeClientStub{
			containerErr: errors.New("container id is not a containerd container id"),
			sandboxPID:   9527,
		}
		server := &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}

		_, _, err := server.getPIDForHTTPChaos(context.Background(), "invalid", "pod-uid")
		if err == nil || !strings.Contains(err.Error(), "not a containerd container id") {
			t.Fatalf("expected invalid container ID error, got %v", err)
		}
		if runtimeClient.sandboxCalls != 0 {
			t.Fatalf("expected no sandbox lookup, got %d", runtimeClient.sandboxCalls)
		}
	})

	t.Run("does not fall back for unrelated runtime errors", func(t *testing.T) {
		runtimeClient := &httpChaosRuntimeClientStub{
			containerErr: errors.New("containerd unavailable"),
			sandboxPID:   9527,
		}
		server := &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}

		_, _, err := server.getPIDForHTTPChaos(context.Background(), "containerd://container-id", "pod-uid")
		if err == nil || !strings.Contains(err.Error(), "containerd unavailable") {
			t.Fatalf("expected runtime error, got %v", err)
		}
		if runtimeClient.sandboxCalls != 0 {
			t.Fatalf("expected no sandbox lookup, got %d", runtimeClient.sandboxCalls)
		}
	})

	t.Run("does not fall back without a pod UID", func(t *testing.T) {
		runtimeClient := &httpChaosRuntimeClientStub{
			containerErr: errors.Wrap(containerderrdefs.ErrNotFound, "container not found"),
			sandboxPID:   9527,
		}
		server := &DaemonServer{crClient: runtimeClient, rootLogger: logr.Discard()}

		_, _, err := server.getPIDForHTTPChaos(context.Background(), "containerd://container-id", "")
		if err == nil || !strings.Contains(err.Error(), "pod UID is empty") {
			t.Fatalf("expected empty pod UID error, got %v", err)
		}
		if runtimeClient.sandboxCalls != 0 {
			t.Fatalf("expected no sandbox lookup, got %d", runtimeClient.sandboxCalls)
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

func TestApplyHTTPChaosDoesNotLogTLSContents(t *testing.T) {
	ctx := context.Background()
	core, logs := observer.New(zap.DebugLevel)
	logger := zapr.NewLogger(zap.New(core))
	manager := bpm.StartBackgroundProcessManager(nil, logger)
	t.Cleanup(func() { manager.Shutdown(ctx) })
	executable, err := os.Executable()
	require.NoError(t, err)
	cmd := bpm.DefaultProcessBuilder(executable, "-test.run=^TestHTTPChaosStdioHelper$").
		SetEnv("CHAOS_HTTP_STDIO_HELPER", "1").
		Build(ctx)
	cmd.Stderr = os.Stderr
	process, err := manager.StartProcess(ctx, cmd)
	require.NoError(t, err)

	tls, err := json.Marshal(tproxyconfig.TLSConfig{
		CertFile: tproxyconfig.TLSConfigItem{Type: "Contents", Value: []byte("CERT")},
		KeyFile:  tproxyconfig.TLSConfigItem{Type: "Contents", Value: []byte("KEY")},
		CAFile:   &tproxyconfig.TLSConfigItem{Type: "Contents", Value: []byte("CA")},
	})
	require.NoError(t, err)
	server := &DaemonServer{
		backgroundProcessManager: manager,
		tproxyLocker:             &sync.Map{},
		rootLogger:               logger,
	}
	response, err := server.applyHTTPChaos(ctx, &pb.ApplyHttpChaosRequest{
		Instance:    int64(process.Pair.Pid),
		InstanceUid: process.UID,
		StartTime:   process.Pair.CreateTime,
		ProxyPorts:  []uint32{443},
		Rules:       `[]`,
		Tls:         string(tls),
	})
	require.NoError(t, err)
	require.Equal(t, int32(http.StatusOK), response.StatusCode)
	require.Equal(t, int64(process.Pair.Pid), response.Instance)
	require.Equal(t, process.UID, response.InstanceUid)
	require.Equal(t, process.Pair.CreateTime, response.StartTime)
	require.JSONEq(t, `{"proxy_ports":[443],"rules":[],"tls":{"cert_file":{"type":"Contents","value":[67,69,82,84]},"key_file":{"type":"Contents","value":[75,69,89]},"ca_file":{"type":"Contents","value":[67,65]}}}`, response.Error)
	require.Equal(t, 1, logs.FilterMessage("http chaos applied").Len())
	for _, entry := range logs.All() {
		fields, err := json.Marshal(entry.ContextMap())
		require.NoError(t, err)
		logged := entry.Message + string(fields)
		for _, forbidden := range []string{
			"CERT", "KEY", "CA",
			"[67,69,82,84]", "[75,69,89]", "[67,65]",
			`"config"`, `"tls"`, "cert_file", "key_file", "ca_file",
		} {
			require.NotContains(t, logged, forbidden)
		}
	}
}

func TestHTTPChaosStdioHelper(t *testing.T) {
	if os.Getenv("CHAOS_HTTP_STDIO_HELPER") != "1" {
		return
	}
	request, err := http.ReadRequest(bufio.NewReader(os.Stdin))
	require.NoError(t, err)
	require.Equal(t, http.MethodPut, request.Method)
	require.Equal(t, "/", request.URL.Path)
	body, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.NoError(t, request.Body.Close())
	response := &http.Response{
		StatusCode:    http.StatusOK,
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(len(body)),
		Body:          io.NopCloser(bytes.NewReader(body)),
	}
	require.NoError(t, response.Write(os.Stdout))
	_, err = io.Copy(io.Discard, os.Stdin)
	require.NoError(t, err)
	os.Exit(0)
}
