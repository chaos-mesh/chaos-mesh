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

package physicalmachinechaos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
)

func TestDiskFillRequest(t *testing.T) {
	originalSecurity := config.ControllerCfg.ChaosdSecurityMode
	config.ControllerCfg.ChaosdSecurityMode = false
	t.Cleanup(func() { config.ControllerCfg.ChaosdSecurityMode = originalSecurity })
	for _, test := range []struct {
		name, option string
		want         bool
	}{
		{"enabled", `,"fill-by-fallocate":true`, true},
		{"disabled", `,"fill-by-fallocate":false`, false},
		{"omitted", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var request map[string]any
			var path string
			done := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(done)
				path = r.URL.Path
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Errorf("decode request: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			var chaos v1alpha1.PhysicalMachineChaos
			require.NoError(t, json.Unmarshal([]byte(`{"spec":{"action":"disk-fill","uid":"disk-`+test.name+`","disk-fill":{"path":"/mnt/fill","size":"1M"`+test.option+`}}}`), &chaos))
			chaos.Spec.Address = []string{server.URL}
			impl := &Impl{Log: logr.Discard()}
			phase, err := impl.Apply(context.Background(), 0, []*v1alpha1.Record{{Id: server.URL}}, &chaos)
			require.NoError(t, err)
			<-done
			require.Equal(t, v1alpha1.Injected, phase)
			require.Equal(t, "/api/attack/disk", path)
			require.Equal(t, test.want, request["fallocate"])
			require.NotContains(t, request, "fill-by-fallocate")
			require.Equal(t, "fill", request["action"])
			require.Equal(t, "disk-"+test.name, request["uid"])
			require.Equal(t, "/mnt/fill", request["path"])
			require.Equal(t, "1M", request["size"])
		})
	}
}
