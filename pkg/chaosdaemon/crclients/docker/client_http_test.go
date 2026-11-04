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

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDockerHTTP(t *testing.T) {
	for _, tc := range []struct {
		name       string
		version    string
		envVersion string
		apiVersion string
		emptyList  bool
	}{
		{name: "negotiated", apiVersion: "1.41"},
		{name: "minimum API", apiVersion: "1.40"},
		{name: "empty list", apiVersion: "1.41", emptyList: true},
		{name: "environment", envVersion: "1.41", apiVersion: "1.41"},
		{name: "explicit", version: "1.41", envVersion: "1.42", apiVersion: "1.41"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DOCKER_API_VERSION", tc.envVersion)
			t.Setenv("DOCKER_CERT_PATH", "")
			t.Setenv("DOCKER_TLS_VERIFY", "")
			var requests []string
			var requestCount atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount.Add(1)
				if r.URL.Path == "/_ping" {
					w.Header().Set("API-Version", tc.apiVersion)
					return
				}
				requests = append(requests, r.Method+" "+r.URL.Path)
				if got := r.Header.Get("X-Chaos-Mesh-Test"); got != "docker" {
					t.Errorf("custom header = %q, want docker", got)
				}
				if got := r.Header.Get("User-Agent"); got != "chaos-mesh-test" {
					t.Errorf("custom User-Agent = %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				switch r.Method + " " + strings.TrimPrefix(r.URL.Path, "/v"+tc.apiVersion) {
				case "GET /containers/running/json":
					fmt.Fprint(w, `{"Id":"running","State":{"Pid":9527,"Status":"running"},"Config":{"Labels":{"app":"test"}}}`)
				case "GET /containers/stopped/json":
					fmt.Fprint(w, `{"Id":"stopped","State":{"Pid":0,"Status":"exited"}}`)
				case "GET /containers/missing/json":
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, `{"message":"No such container: missing"}`)
				case "GET /containers/json":
					var filters map[string]map[string]bool
					if err := json.Unmarshal([]byte(r.URL.Query().Get("filters")), &filters); err != nil {
						t.Errorf("decode filters: %v", err)
					}
					if !reflect.DeepEqual(filters, map[string]map[string]bool{"label": {"io.kubernetes.docker.type=container": true}}) {
						t.Errorf("container filters = %v", filters)
					}
					if r.URL.Query().Get("all") == "1" {
						t.Error("list includes stopped containers")
					}
					if tc.emptyList {
						fmt.Fprint(w, `[]`)
					} else {
						fmt.Fprint(w, `[{"Id":"running"}]`)
					}
				case "POST /containers/running/kill":
					if got := r.URL.Query().Get("signal"); got != "SIGKILL" {
						t.Errorf("kill signal = %q, want SIGKILL", got)
					}
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					http.Error(w, "unexpected request", http.StatusBadRequest)
				}
			}))
			defer server.Close()
			c, err := New(server.URL, tc.version, server.Client(), map[string]string{"X-Chaos-Mesh-Test": "docker", "User-Agent": "chaos-mesh-test"})
			if err != nil {
				t.Fatal(err)
			}
			if count := requestCount.Load(); count != 0 {
				t.Fatalf("constructor made %d requests", count)
			}
			ctx := context.Background()
			if pid, err := c.GetPidFromContainerID(ctx, "docker://running"); err != nil || pid != 9527 {
				t.Fatalf("inspect PID = %d, %v", pid, err)
			}
			if labels, err := c.GetLabelsFromContainerID(ctx, "docker://running"); err != nil || !reflect.DeepEqual(labels, map[string]string{"app": "test"}) {
				t.Fatalf("inspect labels = %v, %v", labels, err)
			}
			wantIDs := []string{"docker://running"}
			if tc.emptyList {
				wantIDs = nil
			}
			if ids, err := c.ListContainerIDs(ctx); err != nil || !reflect.DeepEqual(ids, wantIDs) {
				t.Fatalf("list IDs = %v, %v", ids, err)
			}
			if err := c.ContainerKillByContainerID(ctx, "docker://running"); err != nil {
				t.Fatal(err)
			}
			if _, err := c.GetPidFromContainerID(ctx, "docker://stopped"); err == nil || !strings.Contains(err.Error(), "container is not running") {
				t.Fatalf("stopped container error = %v", err)
			}
			if _, err := c.GetPidFromContainerID(ctx, "docker://missing"); err == nil || !strings.Contains(err.Error(), "No such container: missing") {
				t.Fatalf("missing container error = %v", err)
			}
			server.Close()
			want := []string{
				"GET /v" + tc.apiVersion + "/containers/running/json",
				"GET /v" + tc.apiVersion + "/containers/running/json",
				"GET /v" + tc.apiVersion + "/containers/json",
				"POST /v" + tc.apiVersion + "/containers/running/kill",
				"GET /v" + tc.apiVersion + "/containers/stopped/json",
				"GET /v" + tc.apiVersion + "/containers/missing/json",
			}
			if !reflect.DeepEqual(requests, want) {
				t.Fatalf("requests = %v, want %v", requests, want)
			}
		})
	}
}
