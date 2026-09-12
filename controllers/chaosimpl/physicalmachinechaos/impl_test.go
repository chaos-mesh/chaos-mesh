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

package physicalmachinechaos

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-logr/logr"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
)

func TestRequestContext(t *testing.T) {
	secure := config.ControllerCfg.ChaosdSecurityMode
	config.ControllerCfg.ChaosdSecurityMode = false
	t.Cleanup(func() { config.ControllerCfg.ChaosdSecurityMode = secure })

	impl := &Impl{Log: logr.Discard()}
	for _, operation := range []struct {
		name   string
		method string
		call   func(context.Context, int, []*v1alpha1.Record, v1alpha1.InnerObject) (v1alpha1.Phase, error)
		failed v1alpha1.Phase
		done   v1alpha1.Phase
	}{
		{"apply", http.MethodPost, impl.Apply, v1alpha1.NotInjected, v1alpha1.Injected},
		{"recover", http.MethodDelete, impl.Recover, v1alpha1.Injected, v1alpha1.NotInjected},
	} {
		t.Run(operation.name, func(t *testing.T) {
			for _, cancellation := range []string{"before request", "during request", "none"} {
				t.Run(cancellation, func(t *testing.T) {
					var requests atomic.Int32
					started := make(chan struct{})
					release := make(chan struct{})
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requests.Add(1)
						if r.Method != operation.method {
							t.Errorf("method = %s, want %s", r.Method, operation.method)
						}
						if cancellation == "during request" {
							close(started)
							select {
							case <-r.Context().Done():
							case <-release:
							}
						}
						w.WriteHeader(http.StatusOK)
					}))
					defer server.Close()
					defer close(release)
					chaos := &v1alpha1.PhysicalMachineChaos{
						Spec: v1alpha1.PhysicalMachineChaosSpec{
							Action: "stress-mem",
							PhysicalMachineSelector: v1alpha1.PhysicalMachineSelector{
								Address: []string{server.URL},
							},
							ExpInfo: v1alpha1.ExpInfo{
								UID:          "test-experiment",
								StressMemory: &v1alpha1.StressMemorySpec{Size: "250MB"},
							},
						},
					}
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					if cancellation == "before request" {
						cancel()
					}
					type result struct {
						phase v1alpha1.Phase
						err   error
					}
					finished := make(chan result, 1)
					go func() {
						phase, err := operation.call(ctx, 0, []*v1alpha1.Record{{Id: server.URL}}, chaos)
						finished <- result{phase, err}
					}()
					if cancellation == "during request" {
						select {
						case <-started:
							cancel()
						case <-time.After(2 * time.Second):
							t.Fatal("request did not reach the server")
						}
					}
					select {
					case got := <-finished:
						if cancellation == "none" {
							if got.err != nil || got.phase != operation.done || requests.Load() != 1 {
								t.Fatalf("normal request: phase=%s err=%v requests=%d", got.phase, got.err, requests.Load())
							}
						} else if !errors.Is(got.err, context.Canceled) || got.phase != operation.failed {
							t.Errorf("canceled request: phase=%s err=%v", got.phase, got.err)
						}
					case <-time.After(2 * time.Second):
						t.Fatal("operation did not return after cancellation")
					}
					if cancellation == "before request" && requests.Load() != 0 {
						t.Error("sent a request after the reconciliation was canceled")
					}
				})
			}
		})
	}
}
