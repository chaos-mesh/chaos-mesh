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
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestWithClientDoesNotBlockOtherClusters(t *testing.T) {
	registry := New(logr.Discard(), nil)
	registry.clusters["a"] = &remoteCluster{gate: make(chan struct{}, 1), state: clusterReady}
	registry.clusters["b"] = &remoteCluster{gate: make(chan struct{}, 1), state: clusterReady}
	entered := make(chan struct{})
	release := make(chan struct{})
	aDone := make(chan error, 1)
	go func() {
		aDone <- registry.WithClient("a", func(client.Client) error { close(entered); <-release; return nil })
	}()
	<-entered
	bDone := make(chan error, 1)
	go func() { bDone <- registry.WithClient("b", func(client.Client) error { return nil }) }()
	progressed := false
	select {
	case err := <-bDone:
		require.NoError(t, err)
		progressed = true
	case <-time.After(time.Second):
	}
	close(release)
	require.NoError(t, <-aDone)
	if !progressed {
		require.NoError(t, <-bDone)
	}
	require.True(t, progressed, "remote B must remain available while an A callback is blocked")
}

func testApp(stop func(context.Context) error) *fx.App {
	return fx.New(fx.NopLogger, fx.Invoke(func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{OnStop: stop})
	}))
}

func TestStopDrainsCallbackAndReservesNameUntilActualShutdown(t *testing.T) {
	registry := New(logr.Discard(), nil)
	stopping, releaseStop := make(chan struct{}), make(chan struct{})
	defer close(releaseStop)
	app := testApp(func(context.Context) error { close(stopping); <-releaseStop; return nil })
	require.NoError(t, registry.spawn("a", func() (*fx.App, client.Client) { return app, nil }))
	entered, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- registry.WithClient("a", func(client.Client) error { close(entered); <-release; return nil })
	}()
	<-entered
	require.ErrorIs(t, registry.WithClient("a", func(client.Client) error { t.Error("second callback admitted"); return nil }), ErrBusy)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, registry.Stop(ctx, "a"), context.DeadlineExceeded)
	require.ErrorIs(t, registry.WithClient("a", func(client.Client) error { t.Error("callback admitted after stop intent"); return nil }), ErrBusy)
	select {
	case <-stopping:
		t.Error("shutdown started before callback finished")
	default:
	}
	close(release)
	require.NoError(t, <-done)
	<-stopping
	require.ErrorIs(t, registry.spawn("a", func() (*fx.App, client.Client) { t.Error("replacement constructed during shutdown"); return nil, nil }), ErrBusy)
	require.NoError(t, registry.spawn("b", func() (*fx.App, client.Client) { return testApp(nil), nil }))
	require.NoError(t, registry.Stop(context.Background(), "b"))
}

func TestStopLateCompletionAllowsReplacement(t *testing.T) {
	registry := New(logr.Discard(), nil)
	release := make(chan struct{})
	require.NoError(t, registry.spawn("a", func() (*fx.App, client.Client) {
		return testApp(func(context.Context) error { <-release; return nil }), nil
	}))
	old := registry.clusters["a"]
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, registry.Stop(ctx, "a"), context.DeadlineExceeded)
	joined := make(chan error, 1)
	waiter := &observedContext{Context: context.Background(), observed: make(chan struct{})}
	go func() { joined <- registry.Stop(waiter, "a") }()
	<-waiter.observed
	close(release)
	require.NoError(t, <-joined)
	require.NoError(t, registry.spawn("a", func() (*fx.App, client.Client) { return testApp(nil), nil }))
	require.Equal(t, clusterStopped, old.state)
	require.NotSame(t, old, registry.clusters["a"])
	require.NoError(t, registry.WithClient("a", func(client.Client) error { return nil }))
	require.NoError(t, registry.Stop(context.Background(), "a"))
}

func TestFailedShutdownDoesNotAllowReuse(t *testing.T) {
	registry := New(logr.Discard(), nil)
	failure := errors.New("stop failed")
	require.NoError(t, registry.spawn("a", func() (*fx.App, client.Client) { return testApp(func(context.Context) error { return failure }), nil }))
	require.ErrorIs(t, registry.Stop(context.Background(), "a"), failure)
	require.ErrorIs(t, registry.Stop(context.Background(), "a"), failure)
	require.ErrorIs(t, registry.WithClient("a", func(client.Client) error { t.Error("failed shutdown reopened"); return nil }), ErrBusy)
	require.ErrorIs(t, registry.spawn("a", func() (*fx.App, client.Client) { t.Error("failed shutdown replaced"); return nil, nil }), ErrBusy)
}

func TestStartupReservationAndFailure(t *testing.T) {
	registry := New(logr.Discard(), nil)
	entered, release := make(chan struct{}), make(chan struct{})
	failure := errors.New("start failed")
	done := make(chan error, 1)
	go func() {
		done <- registry.spawn("a", func() (*fx.App, client.Client) {
			close(entered)
			<-release
			return fx.New(fx.NopLogger, fx.Invoke(func() error { return failure })), nil
		})
	}()
	<-entered
	require.ErrorIs(t, registry.spawn("a", func() (*fx.App, client.Client) { t.Error("duplicate startup"); return nil, nil }), ErrBusy)
	require.ErrorIs(t, registry.WithClient("a", func(client.Client) error { t.Error("unfinished startup exposed"); return nil }), ErrBusy)
	require.NoError(t, registry.spawn("b", func() (*fx.App, client.Client) { return testApp(nil), nil }))
	close(release)
	require.ErrorIs(t, <-done, failure)
	require.NoError(t, registry.spawn("a", func() (*fx.App, client.Client) { return testApp(nil), nil }))
	require.ErrorIs(t, registry.spawn("a", func() (*fx.App, client.Client) { t.Error("ready entry replaced"); return nil, nil }), ErrAlreadyExist)
	require.NoError(t, registry.Stop(context.Background(), "a"))
	require.NoError(t, registry.Stop(context.Background(), "b"))
}

type observedContext struct {
	context.Context
	observed chan struct{}
	once     sync.Once
}

func (ctx *observedContext) Done() <-chan struct{} {
	ctx.once.Do(func() { close(ctx.observed) })
	return ctx.Context.Done()
}

type testManager struct {
	ctrl.Manager
	config *rest.Config
	scheme *runtime.Scheme
	start  func(context.Context) error
}

func (m testManager) GetConfig() *rest.Config         { return m.config }
func (m testManager) GetScheme() *runtime.Scheme      { return m.scheme }
func (m testManager) Start(ctx context.Context) error { return m.start(ctx) }

func TestManagerStartsOnlyWithLifecycle(t *testing.T) {
	started, stopped := make(chan struct{}), make(chan struct{})
	mgr := testManager{start: func(ctx context.Context) error { close(started); <-ctx.Done(); close(stopped); return nil }}
	app := fx.New(fx.NopLogger, fx.Invoke(func(lc fx.Lifecycle) error { return run(lc, mgr, logr.Discard()) }))
	select {
	case <-started:
		t.Fatal("construction started manager before lifecycle")
	default:
	}
	require.NoError(t, app.Start(context.Background()))
	<-started
	require.NoError(t, app.Stop(context.Background()))
	<-stopped
}

func TestFailedStartDrainsAlreadyStartedManager(t *testing.T) {
	registry := New(logr.Discard(), nil)
	started, stopping, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	mgr := testManager{start: func(ctx context.Context) error { close(started); <-ctx.Done(); close(stopping); <-release; return nil }}
	failure := errors.New("later hook failed")
	done := make(chan error, 1)
	go func() {
		done <- registry.spawn("a", func() (*fx.App, client.Client) {
			app := fx.New(fx.NopLogger, fx.Invoke(func(lc fx.Lifecycle) error {
				if err := run(lc, mgr, logr.Discard()); err != nil {
					return err
				}
				lc.Append(fx.Hook{OnStart: func(context.Context) error { <-started; return failure }})
				return nil
			}))
			return app, nil
		})
	}()
	<-stopping
	require.ErrorIs(t, registry.spawn("a", func() (*fx.App, client.Client) { t.Error("replacement overlaps startup cleanup"); return nil, nil }), ErrBusy)
	close(release)
	require.ErrorIs(t, <-done, failure)
	require.NoError(t, registry.spawn("a", func() (*fx.App, client.Client) { return testApp(nil), nil }))
	require.NoError(t, registry.Stop(context.Background(), "a"))
}

func TestForwardingConfigPreservesStricterTimeoutAndInput(t *testing.T) {
	for _, timeout := range []time.Duration{0, time.Millisecond, time.Minute} {
		input := &rest.Config{Timeout: timeout}
		copied := forwardingConfig(input)
		require.Equal(t, timeout, input.Timeout)
		require.NotSame(t, input, copied)
		expected := remoteRequestTimeout
		if timeout > 0 && timeout < expected {
			expected = timeout
		}
		require.Equal(t, expected, copied.Timeout)
	}
}

func TestForwardingClientReadsLiveAndHonorsCancellation(t *testing.T) {
	entered := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if serveForwardingDiscovery(w, r) {
			return
		}
		entered <- struct{}{}
		<-r.Context().Done()
	}))
	defer server.Close()
	config := &rest.Config{Host: server.URL}
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	c, err := forwardingClient(testManager{config: config, scheme: scheme})
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	err = c.Get(ctx, client.ObjectKey{Namespace: "test", Name: "experiment"}, &v1alpha1.PodChaos{})
	require.Error(t, err)
	require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
	select {
	case <-entered:
	default:
		t.Fatal("read did not reach live API")
	}
}

func TestDiscoveryUsesTransportTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	config := &rest.Config{Host: server.URL, Timeout: 30 * time.Millisecond}
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	c, err := forwardingClient(testManager{config: config, scheme: scheme})
	require.NoError(t, err)
	started := time.Now()
	err = c.Get(context.Background(), client.ObjectKey{Namespace: "test", Name: "experiment"}, &v1alpha1.PodChaos{})
	require.Error(t, err)
	require.Less(t, time.Since(started), time.Second, "discovery must not wait indefinitely without reconcile context")
}

func TestRemoteAppConstructsNamedForwardingClient(t *testing.T) {
	registry := New(logr.Discard(), nil)
	app, c := registry.newApp("test", &rest.Config{Host: "http://127.0.0.1:1"})
	require.NoError(t, app.Err())
	require.NotNil(t, c)
	require.NoError(t, app.Stop(context.Background()))
}

func TestStopDuringStartupKeepsAdmissionClosed(t *testing.T) {
	registry := New(logr.Discard(), nil)
	entered, release := make(chan struct{}), make(chan struct{})
	spawned := make(chan error, 1)
	go func() {
		spawned <- registry.spawn("a", func() (*fx.App, client.Client) { close(entered); <-release; return testApp(nil), nil })
	}()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	require.ErrorIs(t, registry.Stop(ctx, "a"), context.DeadlineExceeded)
	registry.lock.Lock()
	entry := registry.clusters["a"]
	registry.lock.Unlock()
	require.ErrorIs(t, registry.WithClient("a", func(client.Client) error { t.Error("stopping startup exposed"); return nil }), ErrBusy)
	close(release)
	require.NoError(t, <-spawned)
	<-entry.stopped
	require.ErrorIs(t, registry.WithClient("a", func(client.Client) error { t.Error("stopped startup exposed"); return nil }), ErrNotExist)
}

func TestCanceledStopDoesNotCloseAdmission(t *testing.T) {
	registry := New(logr.Discard(), nil)
	require.NoError(t, registry.spawn("a", func() (*fx.App, client.Client) { return testApp(nil), nil }))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, registry.Stop(ctx, "a"), context.Canceled)
	require.NoError(t, registry.WithClient("a", func(client.Client) error { return nil }))
	require.NoError(t, registry.Stop(context.Background(), "a"))
}
