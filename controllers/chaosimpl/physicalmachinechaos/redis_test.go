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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
)

func TestApplyRedisSentinelPath(t *testing.T) {
	security := config.ControllerCfg.ChaosdSecurityMode
	config.ControllerCfg.ChaosdSecurityMode = false
	t.Cleanup(func() { config.ControllerCfg.ChaosdSecurityMode = security })
	for _, action := range []string{"redis-stop", "redis-restart"} {
		for _, path := range []struct{ name, field, want string }{
			{"directory", `,"redisPath":"/opt/redis/bin"`, "/opt/redis/bin"},
			{"empty", `,"redisPath":""`, ""},
			{"omitted", "", ""},
		} {
			t.Run(action+"/"+path.name, func(t *testing.T) {
				var obj v1alpha1.PhysicalMachineChaos
				input := fmt.Sprintf(`{"spec":{"action":%q,%q:{"addr":"localhost:26379","conf":"/etc/redis/sentinel.conf"%s}}}`, action, action, path.field)
				require.NoError(t, json.Unmarshal([]byte(input), &obj))
				var body map[string]any
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPost || r.URL.Path != "/api/attack/redis" {
						t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("decode request: %v", err)
					}
					w.WriteHeader(http.StatusOK)
				}))
				defer server.Close()
				obj.Spec.Address = []string{server.URL}
				obj.Spec.UID = action + "-" + path.name
				impl := Impl{Log: logr.Discard()}
				phase, err := impl.Apply(context.Background(), 0, []*v1alpha1.Record{{Id: server.URL}}, &obj)
				require.NoError(t, err)
				require.Equal(t, v1alpha1.Injected, phase)
				require.Equal(t, action[len("redis-"):], body["action"])
				require.Equal(t, "/etc/redis/sentinel.conf", body["conf"])
				if path.want == "" {
					require.NotContains(t, body, "redisPath")
				} else {
					require.Equal(t, path.want, body["redisPath"])
				}
			})
		}
	}
}
