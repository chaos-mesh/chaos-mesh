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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
)

func TestApplyStoresChaosdUIDAndRecoverUsesIt(t *testing.T) {
	originalSecurityMode := config.ControllerCfg.ChaosdSecurityMode
	config.ControllerCfg.ChaosdSecurityMode = false
	t.Cleanup(func() {
		config.ControllerCfg.ChaosdSecurityMode = originalSecurityMode
	})

	const uid = "chaosd-uid"
	chaosd := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			require.Equal(t, "/api/attack/stress", request.URL.Path)
			require.NoError(t, json.NewEncoder(responseWriter).Encode(map[string]string{"uid": uid}))
		case http.MethodDelete:
			require.Equal(t, "/api/attack/"+uid, request.URL.Path)
			responseWriter.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected request method %s", request.Method)
		}
	}))
	defer chaosd.Close()

	chaos := &v1alpha1.PhysicalMachineChaos{
		Spec: v1alpha1.PhysicalMachineChaosSpec{
			Action: v1alpha1.PMStressMemAction,
			PhysicalMachineSelector: v1alpha1.PhysicalMachineSelector{
				Address: []string{chaosd.URL},
			},
			ExpInfo: v1alpha1.ExpInfo{
				StressMemory: &v1alpha1.StressMemorySpec{Size: "250MB"},
			},
		},
	}
	records := []*v1alpha1.Record{{Id: chaosd.URL, Phase: v1alpha1.NotInjected}}
	impl := &Impl{Log: logr.Discard()}

	phase, err := impl.Apply(context.Background(), 0, records, chaos)
	require.NoError(t, err)
	require.Equal(t, v1alpha1.Injected, phase)
	require.Equal(t, uid, chaos.Status.ChaosdUIDs[chaosd.URL])

	phase, err = impl.Recover(context.Background(), 0, records, chaos)
	require.NoError(t, err)
	require.Equal(t, v1alpha1.NotInjected, phase)
	require.NotContains(t, chaos.Status.ChaosdUIDs, chaosd.URL)
}

func TestRecoverFailsWhenChaosdUIDIsMissing(t *testing.T) {
	t.Parallel()

	chaos := &v1alpha1.PhysicalMachineChaos{
		Spec: v1alpha1.PhysicalMachineChaosSpec{
			PhysicalMachineSelector: v1alpha1.PhysicalMachineSelector{
				Address: []string{"http://chaosd.example"},
			},
		},
	}
	records := []*v1alpha1.Record{{Id: "http://chaosd.example"}}

	phase, err := (&Impl{Log: logr.Discard()}).Recover(context.Background(), 0, records, chaos)
	require.Equal(t, v1alpha1.Injected, phase)
	require.EqualError(t, err, "chaosd uid not found for target http://chaosd.example")
}
