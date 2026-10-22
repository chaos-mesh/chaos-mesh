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

package ttlcontroller

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"testing/synctest"
	"time"

	"github.com/go-logr/logr"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type eventCleanup struct {
	core.EventStore
	cleanup func(context.Context, time.Duration) error
}

func (s eventCleanup) DeleteByDuration(ctx context.Context, ttl time.Duration) error {
	return s.cleanup(ctx, ttl)
}

type experimentCleanup struct {
	core.ExperimentStore
	cleanup func(context.Context, time.Duration) error
}

func (s experimentCleanup) DeleteByFinishTime(ctx context.Context, ttl time.Duration) error {
	return s.cleanup(ctx, ttl)
}

type scheduleCleanup struct {
	core.ScheduleStore
	cleanup func(context.Context, time.Duration) error
}

func (s scheduleCleanup) DeleteByFinishTime(ctx context.Context, ttl time.Duration) error {
	return s.cleanup(ctx, ttl)
}

type workflowCleanup struct {
	core.WorkflowStore
	cleanup func(context.Context, time.Duration) error
}

func (s workflowCleanup) DeleteByFinishTime(ctx context.Context, ttl time.Duration) error {
	return s.cleanup(ctx, ttl)
}

func TestRegisterPropagatesCancellationAndStops(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		type contextKey struct{}
		ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "worker"))
		defer cancel()
		entered := make(chan context.Context, 1)
		release := make(chan struct{})
		calls := 0
		cleanup := func(ctx context.Context, _ time.Duration) error {
			if ctx.Value(contextKey{}) != "worker" {
				t.Error("cleanup lost context value")
			}
			return ctx.Err()
		}
		c := NewController(eventCleanup{cleanup: func(ctx context.Context, _ time.Duration) error {
			calls++
			entered <- ctx
			<-release
			return ctx.Err()
		}}, experimentCleanup{cleanup: cleanup}, scheduleCleanup{cleanup: cleanup}, workflowCleanup{cleanup: cleanup}, &config.TTLConfig{ResyncPeriod: time.Nanosecond}, logr.Discard())
		Register(ctx, c)
		received := <-entered
		cancel()
		if !errors.Is(received.Err(), context.Canceled) {
			t.Errorf("running cleanup context error = %v, want cancellation", received.Err())
		}
		close(release)
		synctest.Wait()
		if calls != 1 {
			t.Fatalf("cleanup ran %d times after cancellation", calls)
		}
		Register(ctx, c)
		synctest.Wait()
		if calls != 1 {
			t.Fatalf("already canceled controller started cleanup, calls %d", calls)
		}
	})
}

func TestWorkerContinuesAfterEachStoreFailureAndRecovers(t *testing.T) {
	for _, failed := range []string{"event", "experiment", "schedule", "workflow"} {
		t.Run(failed, func(t *testing.T) {
			var called []string
			failing := true
			cleanup := func(name string) func(context.Context, time.Duration) error {
				return func(context.Context, time.Duration) error {
					called = append(called, name)
					if failing && name == failed {
						return errors.New("store unavailable")
					}
					return nil
				}
			}
			c := NewController(eventCleanup{cleanup: cleanup("event")}, experimentCleanup{cleanup: cleanup("experiment")}, scheduleCleanup{cleanup: cleanup("schedule")}, workflowCleanup{cleanup: cleanup("workflow")}, &config.TTLConfig{}, logr.Discard())
			for range 2 {
				called = nil
				c.runWorker(context.Background())
				counts := map[string]int{}
				for _, name := range called {
					counts[name]++
				}
				if !reflect.DeepEqual(counts, map[string]int{"event": 1, "experiment": 1, "schedule": 1, "workflow": 1}) {
					t.Fatalf("stores called = %v", called)
				}
				failing = false
			}
		})
	}
}

func TestCanceledWorkerDoesNotStartCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := 0
	cleanup := func(context.Context, time.Duration) error { called++; return nil }
	c := NewController(eventCleanup{cleanup: cleanup}, experimentCleanup{cleanup: cleanup}, scheduleCleanup{cleanup: cleanup}, workflowCleanup{cleanup: cleanup}, &config.TTLConfig{}, logr.Discard())
	c.runWorker(ctx)
	if called != 0 {
		t.Fatalf("canceled worker started %d cleanups", called)
	}
}
