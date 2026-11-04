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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func awaitSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatal("operation did not reach barrier")
	}
}

func awaitResult(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("operation did not complete")
		return nil
	}
}

func runAsync(f func() error) <-chan error {
	done := make(chan error, 1)
	go func() { done <- f() }()
	return done
}

func TestEnsureLifecycleDoesNotBlockOtherClusters(t *testing.T) {
	for _, phase := range []string{"build", "start", "stop", "drain", "cleanup"} {
		t.Run(phase, func(t *testing.T) {
			r := New(logr.Discard(), nil)
			blocked, release, stopped := make(chan struct{}), make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(release) }) }
			defer unblock()
			aClient, bClient := fake.NewClientBuilder().Build(), fake.NewClientBuilder().Build()
			failure := errors.New("startup failed")
			r.build = func(name string, cfg *rest.Config) (appLifecycle, client.Client) {
				if name == "b" {
					return lifecycleFuncs{}, bClient
				}
				if cfg.BearerToken != "old" {
					return lifecycleFuncs{}, aClient
				}
				if phase == "build" {
					close(blocked)
					<-release
				}
				return lifecycleFuncs{start: func(context.Context) error {
					if phase == "start" {
						close(blocked)
						<-release
					}
					if phase == "cleanup" {
						return failure
					}
					return nil
				}, stop: func(context.Context) error {
					close(stopped)
					if phase == "stop" || phase == "cleanup" {
						close(blocked)
						<-release
					}
					return nil
				}}, aClient
			}
			cfg := desiredConfig(t)
			require.NoError(t, r.Ensure(context.Background(), "b", cfg))
			var aDone <-chan error
			if phase == "stop" || phase == "drain" {
				require.NoError(t, r.Ensure(context.Background(), "a", cfg))
				var callbackDone <-chan error
				if phase == "drain" {
					callbackDone = runAsync(func() error {
						return r.WithClient("a", func(client.Client) error { close(blocked); <-release; return nil })
					})
					awaitSignal(t, blocked)
				}
				changed := cfg.DeepCopy()
				changed.AuthInfos["remote"].Token = "new"
				observed := &observedContext{Context: context.Background(), observed: make(chan struct{})}
				aDone = runAsync(func() error { return r.Ensure(observed, "a", changed) })
				if phase == "drain" {
					awaitSignal(t, observed.observed)
					select {
					case <-stopped:
						t.Fatal("manager stopped before callback drained")
					default:
					}
					defer func() { require.NoError(t, awaitResult(t, callbackDone)) }()
				} else {
					awaitSignal(t, blocked)
				}
			} else {
				aDone = runAsync(func() error { return r.Ensure(context.Background(), "a", cfg) })
				awaitSignal(t, blocked)
			}
			require.ErrorIs(t, r.WithClient("a", func(client.Client) error { return nil }), ErrBusy)
			bDone := runAsync(func() error {
				if err := r.WithClient("b", func(c client.Client) error {
					if c != bClient {
						return errors.New("wrong B client")
					}
					return nil
				}); err != nil {
					return err
				}
				changed := cfg.DeepCopy()
				changed.AuthInfos["remote"].Token = "b-new"
				if err := r.Ensure(context.Background(), "b", changed); err != nil {
					return err
				}
				if err := r.Stop(context.Background(), "b"); err != nil {
					return err
				}
				if err := r.Spawn("b", &rest.Config{}); err != nil {
					return err
				}
				return r.Stop(context.Background(), "b")
			})
			require.NoError(t, awaitResult(t, bDone))
			unblock()
			err := awaitResult(t, aDone)
			if phase == "cleanup" {
				require.ErrorIs(t, err, failure)
				require.ErrorIs(t, r.WithClient("a", func(client.Client) error { return nil }), ErrNotExist)
			} else {
				require.NoError(t, err)
				require.NoError(t, r.WithClient("a", func(c client.Client) error { require.Same(t, aClient, c); return nil }))
				require.NoError(t, r.Stop(context.Background(), "a"))
			}
		})
	}
}

func TestConcurrentEnsureKeepsSingleStartingManager(t *testing.T) {
	for _, token := range []string{"old", "different"} {
		t.Run(token, func(t *testing.T) {
			r := New(logr.Discard(), nil)
			entered, release := make(chan struct{}), make(chan struct{})
			var builds atomic.Int32
			oldClient, newClient := fake.NewClientBuilder().Build(), fake.NewClientBuilder().Build()
			r.build = func(string, *rest.Config) (appLifecycle, client.Client) {
				if builds.Add(1) == 1 {
					close(entered)
					<-release
					return lifecycleFuncs{}, oldClient
				}
				return lifecycleFuncs{}, newClient
			}
			cfg := desiredConfig(t)
			done := runAsync(func() error { return r.Ensure(context.Background(), "a", cfg) })
			awaitSignal(t, entered)
			contender := cfg.DeepCopy()
			contender.AuthInfos["remote"].Token = token
			require.ErrorIs(t, r.Ensure(context.Background(), "a", contender), ErrBusy)
			require.ErrorIs(t, r.Spawn("a", &rest.Config{}), ErrBusy)
			close(release)
			require.NoError(t, awaitResult(t, done))
			require.NoError(t, r.Ensure(context.Background(), "a", contender))
			expectedClient, expectedBuilds := oldClient, int32(1)
			if token != "old" {
				expectedClient, expectedBuilds = newClient, 2
			}
			require.Equal(t, expectedBuilds, builds.Load())
			require.NoError(t, r.WithClient("a", func(c client.Client) error { require.Same(t, expectedClient, c); return nil }))
			require.NoError(t, r.Stop(context.Background(), "a"))
		})
	}
}

func TestCanceledEnsureWaitsForActualCleanup(t *testing.T) {
	for _, phase := range []string{"build", "start", "cleanup", "old-stop"} {
		t.Run(phase, func(t *testing.T) {
			r := New(logr.Discard(), nil)
			entered, release, cleanup, releaseCleanup := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
			var builds atomic.Int32
			failure := errors.New("start failed")
			expectedClient := fake.NewClientBuilder().Build()
			r.build = func(string, *rest.Config) (appLifecycle, client.Client) {
				if builds.Add(1) > 1 {
					return lifecycleFuncs{}, expectedClient
				}
				if phase == "build" {
					close(entered)
					<-release
				}
				return lifecycleFuncs{start: func(context.Context) error {
					if phase == "start" {
						close(entered)
						<-release
					}
					if phase == "cleanup" {
						return failure
					}
					return nil
				}, stop: func(context.Context) error {
					close(cleanup)
					<-releaseCleanup
					return nil
				}}, expectedClient
			}
			cfg := desiredConfig(t)
			if phase == "old-stop" {
				require.NoError(t, r.Ensure(context.Background(), "a", cfg))
				cfg.AuthInfos["remote"].Token = "new"
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := runAsync(func() error { return r.Ensure(ctx, "a", cfg) })
			if phase == "cleanup" || phase == "old-stop" {
				awaitSignal(t, cleanup)
			} else {
				awaitSignal(t, entered)
			}
			cancel()
			require.ErrorIs(t, awaitResult(t, done), context.Canceled)
			require.ErrorIs(t, r.Ensure(context.Background(), "a", cfg), ErrBusy)
			require.ErrorIs(t, r.Spawn("a", &rest.Config{}), ErrBusy)
			if phase == "build" || phase == "start" {
				close(release)
				awaitSignal(t, cleanup)
			}
			require.ErrorIs(t, r.WithClient("a", func(client.Client) error { return nil }), ErrBusy)
			r.lock.Lock()
			entry := r.clusters["a"]
			r.lock.Unlock()
			close(releaseCleanup)
			awaitSignal(t, entry.stopped)
			require.NoError(t, r.Ensure(context.Background(), "a", cfg))
			require.Equal(t, int32(2), builds.Load())
			require.NoError(t, r.WithClient("a", func(c client.Client) error { require.Same(t, expectedClient, c); return nil }))
			require.NoError(t, r.Stop(context.Background(), "a"))
		})
	}
}

type pausedWaitContext struct {
	context.Context
	entered chan struct{}
	release chan struct{}
}

func (ctx *pausedWaitContext) Done() <-chan struct{} {
	close(ctx.entered)
	<-ctx.release
	return ctx.Context.Done()
}

func TestOldLifecycleWaiterCannotStopNewEntry(t *testing.T) {
	for _, operation := range []string{"ensure-different", "ensure-equal", "stop"} {
		t.Run(operation, func(t *testing.T) {
			r := New(logr.Discard(), nil)
			var stops atomic.Int32
			expectedClient := fake.NewClientBuilder().Build()
			r.build = func(string, *rest.Config) (appLifecycle, client.Client) {
				return lifecycleFuncs{stop: func(context.Context) error { stops.Add(1); return nil }}, expectedClient
			}
			cfg := desiredConfig(t)
			require.NoError(t, r.Ensure(context.Background(), "a", cfg))
			r.lock.Lock()
			old := r.clusters["a"]
			r.lock.Unlock()
			changed := cfg.DeepCopy()
			changed.AuthInfos["remote"].Token = "new"
			paused := &pausedWaitContext{Context: context.Background(), entered: make(chan struct{}), release: make(chan struct{})}
			done := runAsync(func() error {
				if operation == "stop" {
					return r.Stop(paused, "a")
				}
				return r.Ensure(paused, "a", changed)
			})
			awaitSignal(t, paused.entered)
			awaitSignal(t, old.stopped)
			replacement := cfg
			if operation == "ensure-equal" {
				replacement = changed
			}
			require.NoError(t, r.Ensure(context.Background(), "a", replacement))
			close(paused.release)
			err := awaitResult(t, done)
			if operation == "ensure-different" {
				require.ErrorIs(t, err, ErrBusy)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, int32(1), stops.Load())
			require.NoError(t, r.WithClient("a", func(c client.Client) error { require.Same(t, expectedClient, c); return nil }))
			require.NoError(t, r.Stop(context.Background(), "a"))
		})
	}
}

func TestEnsureEqualConfigDuringCallbackAndCanceledReplacement(t *testing.T) {
	r := New(logr.Discard(), nil)
	var builds atomic.Int32
	expectedClient := fake.NewClientBuilder().Build()
	r.build = func(string, *rest.Config) (appLifecycle, client.Client) {
		builds.Add(1)
		return lifecycleFuncs{}, expectedClient
	}
	cfg := desiredConfig(t)
	require.NoError(t, r.Ensure(context.Background(), "a", cfg))
	entered, release := make(chan struct{}), make(chan struct{})
	done := runAsync(func() error {
		return r.WithClient("a", func(client.Client) error { close(entered); <-release; return nil })
	})
	awaitSignal(t, entered)
	require.NoError(t, r.Ensure(context.Background(), "a", cfg))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	changed := cfg.DeepCopy()
	changed.AuthInfos["remote"].Token = "new"
	require.ErrorIs(t, r.Ensure(ctx, "a", changed), context.Canceled)
	close(release)
	require.NoError(t, awaitResult(t, done))
	require.Equal(t, int32(1), builds.Load())
	require.NoError(t, r.WithClient("a", func(c client.Client) error { require.Same(t, expectedClient, c); return nil }))
	require.NoError(t, r.Stop(context.Background(), "a"))
}

func TestStopDuringEnsureDoesNotReportReady(t *testing.T) {
	for _, phase := range []string{"build", "start"} {
		t.Run(phase, func(t *testing.T) {
			r := New(logr.Discard(), nil)
			entered, release := make(chan struct{}), make(chan struct{})
			var builds atomic.Int32
			expectedClient := fake.NewClientBuilder().Build()
			r.build = func(string, *rest.Config) (appLifecycle, client.Client) {
				if builds.Add(1) > 1 {
					return lifecycleFuncs{}, expectedClient
				}
				if phase == "build" {
					close(entered)
					<-release
				}
				return lifecycleFuncs{start: func(context.Context) error {
					if phase == "start" {
						close(entered)
						<-release
					}
					return nil
				}}, expectedClient
			}
			cfg := desiredConfig(t)
			ensured := runAsync(func() error { return r.Ensure(context.Background(), "a", cfg) })
			awaitSignal(t, entered)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			observed := &observedContext{Context: ctx, observed: make(chan struct{})}
			stopped := runAsync(func() error { return r.Stop(observed, "a") })
			awaitSignal(t, observed.observed)
			cancel()
			require.ErrorIs(t, awaitResult(t, stopped), context.Canceled)
			require.ErrorIs(t, r.WithClient("a", func(client.Client) error { return nil }), ErrBusy)
			require.ErrorIs(t, r.Ensure(context.Background(), "a", cfg), ErrBusy)
			r.lock.Lock()
			entry := r.clusters["a"]
			r.lock.Unlock()
			close(release)
			require.ErrorIs(t, awaitResult(t, ensured), ErrBusy)
			awaitSignal(t, entry.stopped)
			require.NoError(t, r.Ensure(context.Background(), "a", cfg))
			require.NoError(t, r.WithClient("a", func(c client.Client) error { require.Same(t, expectedClient, c); return nil }))
			require.Equal(t, int32(2), builds.Load())
			require.NoError(t, r.Stop(context.Background(), "a"))
		})
	}
}

func TestEnsurePreservesStartupAndCleanupFailures(t *testing.T) {
	r := New(logr.Discard(), nil)
	startFailure, stopFailure := errors.New("startup failed"), errors.New("cleanup failed")
	var starts, stops atomic.Int32
	r.build = func(string, *rest.Config) (appLifecycle, client.Client) {
		return lifecycleFuncs{start: func(context.Context) error { starts.Add(1); return startFailure }, stop: func(context.Context) error { stops.Add(1); return stopFailure }}, nil
	}
	cfg := desiredConfig(t)
	err := r.Ensure(context.Background(), "a", cfg)
	require.ErrorIs(t, err, startFailure)
	require.ErrorIs(t, err, stopFailure)
	require.ErrorIs(t, r.Ensure(context.Background(), "a", cfg), stopFailure)
	require.ErrorIs(t, r.Stop(context.Background(), "a"), stopFailure)
	require.ErrorIs(t, r.Spawn("a", &rest.Config{}), ErrBusy)
	require.ErrorIs(t, r.WithClient("a", func(client.Client) error { return nil }), ErrBusy)
	require.Equal(t, int32(1), starts.Load())
	require.Equal(t, int32(1), stops.Load())
}
