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
	"sync"
	"testing"

	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tproxyconfig"
)

func TestApplyHTTPChaosDoesNotLogTLSContents(t *testing.T) {
	ctx := context.Background()
	core, logs := observer.New(zap.InfoLevel)
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
	response, err := server.applyHttpChaos(ctx, &pb.ApplyHttpChaosRequest{
		Instance:    int64(process.Pair.Pid),
		InstanceUid: process.Uid,
		StartTime:   process.Pair.CreateTime,
		ProxyPorts:  []uint32{443},
		Rules:       `[]`,
		Tls:         string(tls),
	})
	require.NoError(t, err)
	require.Equal(t, int32(http.StatusOK), response.StatusCode)
	require.Equal(t, int64(process.Pair.Pid), response.Instance)
	require.Equal(t, process.Uid, response.InstanceUid)
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
