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

package clusterregistry

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-controller-manager/provider"
)

type recordedRequest struct {
	path        string
	timeout     time.Duration
	hasDeadline bool
}

type recordingTransport struct {
	http.RoundTripper
	mu       sync.Mutex
	requests []recordedRequest
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	deadline, ok := req.Context().Deadline()
	r.mu.Lock()
	r.requests = append(r.requests, recordedRequest{path: req.URL.Path, timeout: time.Until(deadline), hasDeadline: ok})
	r.mu.Unlock()
	return r.RoundTripper.RoundTrip(req)
}

func (r *recordingTransport) take() []recordedRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	requests := r.requests
	r.requests = nil
	return requests
}

func serveForwardingDiscovery(w http.ResponseWriter, r *http.Request) bool {
	var body string
	switch r.URL.Path {
	case "/api":
		body = `{"kind":"APIVersions","apiVersion":"v1","versions":[]}`
	case "/apis":
		body = `{"kind":"APIGroupList","apiVersion":"v1","groups":[{"name":"chaos-mesh.org","versions":[{"groupVersion":"chaos-mesh.org/v1alpha1","version":"v1alpha1"}],"preferredVersion":{"groupVersion":"chaos-mesh.org/v1alpha1","version":"v1alpha1"}}]}`
	case "/apis/chaos-mesh.org/v1alpha1":
		body = `{"kind":"APIResourceList","apiVersion":"v1","groupVersion":"chaos-mesh.org/v1alpha1","resources":[{"name":"podchaos","singularName":"podchaos","namespaced":true,"kind":"PodChaos","verbs":["get","list","watch"]}]}`
	default:
		return false
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, body)
	return true
}

func TestForwardingClientIsolatesManagerTransport(t *testing.T) {
	for _, timeout := range []time.Duration{0, 2 * time.Second, time.Minute} {
		t.Run(timeout.String(), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if serveForwardingDiscovery(w, r) {
					return
				}
				if r.URL.Path != "/apis/chaos-mesh.org/v1alpha1/namespaces/test/podchaos/experiment" {
					t.Errorf("unexpected API path %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"PodChaos","metadata":{"name":"experiment","namespace":"test"}}`)
			}))
			defer server.Close()
			transport := &recordingTransport{}
			cfg := &rest.Config{Host: server.URL, Timeout: timeout, WrapTransport: func(rt http.RoundTripper) http.RoundTripper {
				transport.RoundTripper = rt
				return transport
			}}
			mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: provider.NewScheme(), Metrics: metricsserver.Options{BindAddress: "0"}})
			require.NoError(t, err)
			managerClient, managerMapper, managerHTTP := mgr.GetClient(), mgr.GetRESTMapper(), mgr.GetHTTPClient()
			forwarder, err := forwardingClient(mgr)
			require.NoError(t, err)
			key := client.ObjectKey{Namespace: "test", Name: "experiment"}
			object := &v1alpha1.PodChaos{}
			require.NoError(t, forwarder.Get(context.Background(), key, object))
			require.Equal(t, "experiment", object.Name)
			requests := transport.take()
			paths := make([]string, 0, len(requests))
			expected := 10 * time.Second
			if timeout == 2*time.Second {
				expected = 2 * time.Second
			}
			for _, request := range requests {
				paths = append(paths, request.path)
				require.True(t, request.hasDeadline, "forwarded request %s must be bounded", request.path)
				require.InDelta(t, expected.Seconds(), request.timeout.Seconds(), 0.5, "forwarded request %s", request.path)
			}
			require.Contains(t, paths, "/apis/chaos-mesh.org/v1alpha1")
			require.Contains(t, paths, "/apis/chaos-mesh.org/v1alpha1/namespaces/test/podchaos/experiment")
			require.Same(t, managerClient, mgr.GetClient())
			require.Same(t, managerMapper, mgr.GetRESTMapper())
			require.Same(t, managerHTTP, mgr.GetHTTPClient())
			require.Equal(t, timeout, mgr.GetHTTPClient().Timeout)
			require.Equal(t, timeout, mgr.GetConfig().Timeout)
			require.Equal(t, timeout, cfg.Timeout)

			object = &v1alpha1.PodChaos{}
			require.NoError(t, mgr.GetAPIReader().Get(context.Background(), key, object))
			require.Equal(t, "experiment", object.Name)
			requests = transport.take()
			paths = nil
			for _, request := range requests {
				paths = append(paths, request.path)
				require.Equal(t, timeout != 0, request.hasDeadline, "manager request %s", request.path)
				if timeout != 0 {
					require.InDelta(t, timeout.Seconds(), request.timeout.Seconds(), 0.5, "manager request %s", request.path)
				}
			}
			require.Contains(t, paths, "/apis/chaos-mesh.org/v1alpha1", "forwarding must not populate the manager discovery cache")
			require.Contains(t, paths, "/apis/chaos-mesh.org/v1alpha1/namespaces/test/podchaos/experiment")
		})
	}
}
