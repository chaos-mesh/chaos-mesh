// Copyright 2022 Chaos Mesh Authors.
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
	"errors"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type stopFunc func(context.Context) error

func (f stopFunc) Stop(ctx context.Context) error { return f(ctx) }

func desiredConfig(t *testing.T) *clientcmdapi.Config {
	t.Helper()
	cfg, err := clientcmd.Load([]byte(`apiVersion: v1
kind: Config
current-context: remote
contexts:
- name: remote
  context: {cluster: remote, user: remote}
clusters:
- name: remote
  cluster: {server: https://remote.example, proxy-url: http://proxy.example:8080, certificate-authority-data: Y2E=}
users:
- name: remote
  user: {token: old}
`))
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestEnsureConnectionLifecycle(t *testing.T) {
	ctx := context.Background()
	r := New(logr.Discard(), nil)
	var events []string
	var current client.Client
	r.start = func(_ context.Context, _ string, cfg *rest.Config) (*remoteCluster, error) {
		token := cfg.BearerToken
		events = append(events, "start:"+token)
		cfg.TLSClientConfig.CAData[0] = 'x'
		current = fake.NewClientBuilder().Build()
		return &remoteCluster{Client: current, app: stopFunc(func(context.Context) error {
			events = append(events, "stop:"+token)
			return nil
		})}, nil
	}
	first := desiredConfig(t)
	if err := r.Ensure(ctx, "remote", first); err != nil {
		t.Fatal(err)
	}
	original := current
	if err := r.Ensure(ctx, "remote", desiredConfig(t)); err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("equivalent proxy config restarted manager: %v", events)
	}
	first.AuthInfos["remote"].Token = "new"
	if err := r.Ensure(ctx, "remote", first); err != nil {
		t.Fatal(err)
	}
	if err := r.WithClient("remote", func(c client.Client) error {
		if c == original || c != current {
			t.Fatal("old client retained")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := r.Stop(ctx, "remote"); err != nil {
		t.Fatal(err)
	}
	if err := r.Ensure(ctx, "remote", first); err != nil {
		t.Fatal(err)
	}
	want := []string{"start:old", "stop:old", "start:new", "stop:new", "start:new"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events %v, want %v", events, want)
	}
}

func TestEnsureFailures(t *testing.T) {
	for _, failure := range []string{"invalid", "stop", "start"} {
		t.Run(failure, func(t *testing.T) {
			ctx := context.Background()
			r := New(logr.Discard(), nil)
			starts, stops := 0, 0
			failing := false
			boom := errors.New("failed")
			r.start = func(context.Context, string, *rest.Config) (*remoteCluster, error) {
				starts++
				if failing && failure == "start" {
					return nil, boom
				}
				return &remoteCluster{app: stopFunc(func(context.Context) error {
					stops++
					if failing && failure == "stop" {
						return boom
					}
					return nil
				})}, nil
			}
			cfg := desiredConfig(t)
			if err := r.Ensure(ctx, "remote", cfg); err != nil {
				t.Fatal(err)
			}
			cfg.AuthInfos["remote"].Token = "new"
			if failure == "invalid" {
				cfg.CurrentContext = "missing"
			}
			failing = true
			if err := r.Ensure(ctx, "remote", cfg); err == nil {
				t.Fatal("expected failure")
			}
			switch failure {
			case "invalid":
				if starts != 1 || stops != 0 {
					t.Fatalf("invalid input disturbed live manager: %d/%d", starts, stops)
				}
			case "stop":
				if starts != 1 || stops != 1 {
					t.Fatalf("replacement started despite failed stop: %d/%d", starts, stops)
				}
			case "start":
				if err := r.WithClient("remote", func(client.Client) error { return nil }); !errors.Is(err, ErrNotExist) {
					t.Fatalf("failed start published entry: %v", err)
				}
			}
			failing = false
			cfg.CurrentContext = "remote"
			if err := r.Ensure(ctx, "remote", cfg); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEnsureTLSAndProxyChanges(t *testing.T) {
	for _, change := range []string{"TLS", "proxy"} {
		t.Run(change, func(t *testing.T) {
			ctx := context.Background()
			r := New(logr.Discard(), nil)
			starts, stops := 0, 0
			r.start = func(context.Context, string, *rest.Config) (*remoteCluster, error) {
				starts++
				return &remoteCluster{app: stopFunc(func(context.Context) error { stops++; return nil })}, nil
			}
			cfg := desiredConfig(t)
			if err := r.Ensure(ctx, "remote", cfg); err != nil {
				t.Fatal(err)
			}
			if change == "TLS" {
				cfg.Clusters["remote"].CertificateAuthorityData[0] = 'n'
			} else {
				cfg.Clusters["remote"].ProxyURL = "http://new-proxy.example:8080"
			}
			if err := r.Ensure(ctx, "remote", cfg); err != nil {
				t.Fatal(err)
			}
			if starts != 2 || stops != 1 {
				t.Fatalf("change was not applied: starts=%d stops=%d", starts, stops)
			}
		})
	}
}

func TestSpawnRemainsStrict(t *testing.T) {
	ctx := context.Background()
	r := New(logr.Discard(), nil)
	starts, stops := 0, 0
	r.start = func(context.Context, string, *rest.Config) (*remoteCluster, error) {
		starts++
		return &remoteCluster{app: stopFunc(func(context.Context) error { stops++; return nil })}, nil
	}
	if err := r.Spawn("remote", &rest.Config{}); err != nil {
		t.Fatal(err)
	}
	if err := r.Spawn("remote", &rest.Config{}); !errors.Is(err, ErrAlreadyExist) {
		t.Fatalf("duplicate Spawn: %v", err)
	}
	if err := r.Ensure(ctx, "remote", desiredConfig(t)); err != nil {
		t.Fatal(err)
	}
	if starts != 2 || stops != 1 {
		t.Fatalf("Ensure trusted unknown Spawn config: starts=%d stops=%d", starts, stops)
	}
}
