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
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
)

func TestApplyUID(t *testing.T) {
	securityMode := config.ControllerCfg.ChaosdSecurityMode
	config.ControllerCfg.ChaosdSecurityMode = false
	t.Cleanup(func() { config.ControllerCfg.ChaosdSecurityMode = securityMode })

	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))

	for _, tc := range []struct {
		name          string
		action        v1alpha1.PhysicalMachineChaosAction
		signal        int
		injectedCount int
		target        string
		uid           string
	}{
		{"empty UID signal 9 initial", v1alpha1.PMProcessAction, 9, 0, "address", ""},
		{"empty UID signal 9 reapply", v1alpha1.PMProcessAction, 9, 1, "address", ""},
		{"empty UID signal 19 initial", v1alpha1.PMProcessAction, 19, 0, "address", ""},
		{"empty UID signal 19 reapply", v1alpha1.PMProcessAction, 19, 1, "address", ""},
		{"empty UID selector", v1alpha1.PMProcessAction, 19, 0, "selector", ""},
		{"empty UID missing machine", v1alpha1.PMProcessAction, 19, 0, "missing machine", ""},
		{"empty UID missing machine reapply", v1alpha1.PMProcessAction, 9, 1, "missing machine", ""},
		{"empty UID stress", v1alpha1.PMStressCPUAction, 0, 0, "address", ""},
		{"valid UID signal 9 initial", v1alpha1.PMProcessAction, 9, 0, "address", "original-uid"},
		{"valid UID signal 9 reapply", v1alpha1.PMProcessAction, 9, 1, "address", "original-uid"},
		{"valid UID signal 19 initial", v1alpha1.PMProcessAction, 19, 0, "address", "original-uid"},
		{"valid UID signal 19 reapply", v1alpha1.PMProcessAction, 19, 1, "address", "original-uid"},
		{"valid UID selector", v1alpha1.PMProcessAction, 19, 1, "selector", "original-uid"},
		{"valid UID stress", v1alpha1.PMStressCPUAction, 0, 0, "address", "original-uid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				wantPath := "/api/attack/" + strings.SplitN(string(tc.action), "-", 2)[0]
				if r.Method != http.MethodPost || r.URL.Path != wantPath {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				var body struct {
					UID    string `json:"uid"`
					Signal int    `json:"signal"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode request: %v", err)
				}
				if body.UID != tc.uid || body.Signal != tc.signal {
					t.Errorf("unexpected request body %+v", body)
				}
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(server.Close)

			obj := &v1alpha1.PhysicalMachineChaos{
				Spec: v1alpha1.PhysicalMachineChaosSpec{
					Action: tc.action,
					ExpInfo: v1alpha1.ExpInfo{
						UID: tc.uid,
					},
				},
			}
			if tc.action == v1alpha1.PMProcessAction {
				obj.Spec.Process = &v1alpha1.ProcessSpec{Process: "1234", Signal: tc.signal}
			} else {
				obj.Spec.StressCPU = &v1alpha1.StressCPUSpec{Load: 1, Workers: 1}
			}
			record := &v1alpha1.Record{Id: "default/machine", Phase: v1alpha1.NotInjected, InjectedCount: tc.injectedCount}
			builder := fake.NewClientBuilder().WithScheme(scheme)
			switch tc.target {
			case "address":
				obj.Spec.Address = []string{server.URL}
				record.Id = server.URL
			case "selector":
				builder.WithObjects(&v1alpha1.PhysicalMachine{
					ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "machine"},
					Spec:       v1alpha1.PhysicalMachineSpec{Address: server.URL},
				})
			}
			impl := &Impl{Client: builder.Build(), Log: logr.Discard()}

			phase, err := impl.Apply(context.Background(), 0, []*v1alpha1.Record{record}, obj)
			if tc.uid == "" {
				require.Zero(t, requests.Load())
				require.Equal(t, v1alpha1.NotInjected, phase)
				require.ErrorContains(t, err, "spec.uid is empty")
				require.ErrorContains(t, err, "mutating webhook")
			} else {
				require.NoError(t, err)
				require.Equal(t, v1alpha1.Injected, phase)
				require.EqualValues(t, 1, requests.Load())
			}
			require.Equal(t, tc.uid, obj.Spec.UID)
			require.Equal(t, tc.injectedCount, record.InjectedCount)
		})
	}
}
