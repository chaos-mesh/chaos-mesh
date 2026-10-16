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

package store_test

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	eventstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/event"
	experimentstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/experiment"
	schedulestore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/schedule"
	workflowstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/workflow"
)

type persistenceFixture struct {
	db          *gorm.DB
	events      core.EventStore
	experiments core.ExperimentStore
	schedules   core.ScheduleStore
	workflows   core.WorkflowStore
	event       *core.Event
	experiment  *core.Experiment
	schedule    *core.Schedule
	workflow    *core.WorkflowEntity
}

func newPersistenceFixture(t *testing.T) persistenceFixture {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	f := persistenceFixture{db: db}
	f.events, err = eventstore.NewStore(db)
	require.NoError(t, err)
	f.experiments, err = experimentstore.NewStore(db)
	require.NoError(t, err)
	f.schedules, err = schedulestore.NewStore(db)
	require.NoError(t, err)
	f.workflows, err = workflowstore.NewStore(db)
	require.NoError(t, err)
	finish := time.Now().Add(-2 * time.Hour)
	f.event = &core.Event{ObjectID: "target", CreatedAt: finish}
	f.experiment = &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: "target", Name: "target", Namespace: "team", Archived: true, FinishTime: &finish}, Experiment: `{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"PodChaos","metadata":{}}`}
	f.schedule = &core.Schedule{ScheduleMeta: core.ScheduleMeta{UID: "target", Name: "target", Namespace: "team", Archived: true, FinishTime: &finish}}
	f.workflow = &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "target", Name: "target", Namespace: "team", Archived: true, FinishTime: &finish}}
	for _, row := range []any{f.event, f.experiment, f.schedule, f.workflow} {
		require.NoError(t, db.Create(row).Error)
	}
	return f
}

func TestStoreCancellationAndRecovery(t *testing.T) {
	operations := map[string]func(persistenceFixture, context.Context) error{
		"event/list": func(f persistenceFixture, ctx context.Context) error { _, err := f.events.List(ctx); return err },

		"event/list-uid": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.events.ListByUID(ctx, "target")
			return err
		},
		"event/list-uids": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.events.ListByUIDList(ctx, []string{"target"})
			return err
		},
		"event/list-experiment": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.events.ListByExperiment(ctx, "team", "target", "PodChaos")
			return err
		},
		"event/filter": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.events.ListByFilter(ctx, core.Filter{})
			return err
		},
		"event/uids-filter": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.events.ListByUIDListWithFilter(ctx, []string{"target"}, core.Filter{})
			return err
		},
		"event/find": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.events.Find(ctx, f.event.ID)
			return err
		},
		"event/create": func(f persistenceFixture, ctx context.Context) error {
			return f.events.Create(ctx, &core.Event{ObjectID: "new"})
		},
		"event/delete-uid": func(f persistenceFixture, ctx context.Context) error { return f.events.DeleteByUID(ctx, "target") },
		"event/delete-uids": func(f persistenceFixture, ctx context.Context) error {
			return f.events.DeleteByUIDList(ctx, []string{"target"})
		},
		"event/ttl": func(f persistenceFixture, ctx context.Context) error {
			return f.events.DeleteByDuration(ctx, time.Hour)
		},
		"experiment/list-meta": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.experiments.ListMeta(ctx, "", "", "", true)
			return err
		},
		"experiment/find": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.experiments.FindByUID(ctx, "target")
			return err
		},
		"experiment/find-meta": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.experiments.FindMetaByUID(ctx, "target")
			return err
		},
		"experiment/save": func(f persistenceFixture, ctx context.Context) error { return f.experiments.Save(ctx, f.experiment) },
		"experiment/delete-uids": func(f persistenceFixture, ctx context.Context) error {
			return f.experiments.DeleteByUIDs(ctx, []string{"target"})
		},
		"experiment/ttl": func(f persistenceFixture, ctx context.Context) error {
			return f.experiments.DeleteByFinishTime(ctx, time.Hour)
		},
		"experiment/archive": func(f persistenceFixture, ctx context.Context) error {
			return f.experiments.Archive(ctx, "team", "target")
		},
		"experiment/delete": func(f persistenceFixture, ctx context.Context) error { return f.experiments.Delete(ctx, f.experiment) },
		"experiment/delete-incomplete": func(f persistenceFixture, ctx context.Context) error {
			return f.experiments.DeleteIncompleteExperiments(ctx)
		},
		"schedule/list-meta": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.schedules.ListMeta(ctx, "", "", true)
			return err
		},
		"schedule/find": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.schedules.FindByUID(ctx, "target")
			return err
		},
		"schedule/find-meta": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.schedules.FindMetaByUID(ctx, "target")
			return err
		},
		"schedule/save": func(f persistenceFixture, ctx context.Context) error { return f.schedules.Save(ctx, f.schedule) },
		"schedule/delete-uids": func(f persistenceFixture, ctx context.Context) error {
			return f.schedules.DeleteByUIDs(ctx, []string{"target"})
		},
		"schedule/ttl": func(f persistenceFixture, ctx context.Context) error {
			return f.schedules.DeleteByFinishTime(ctx, time.Hour)
		},
		"schedule/archive": func(f persistenceFixture, ctx context.Context) error {
			return f.schedules.Archive(ctx, "team", "target")
		},
		"schedule/delete": func(f persistenceFixture, ctx context.Context) error { return f.schedules.Delete(ctx, f.schedule) },
		"schedule/delete-incomplete": func(f persistenceFixture, ctx context.Context) error {
			return f.schedules.DeleteIncompleteSchedules(ctx)
		},
		"workflow/list-meta": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.workflows.ListMeta(ctx, "", "", true)
			return err
		},
		"workflow/find": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.workflows.FindByUID(ctx, "target")
			return err
		},
		"workflow/find-meta": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.workflows.FindMetaByUID(ctx, "target")
			return err
		},
		"workflow/save": func(f persistenceFixture, ctx context.Context) error { return f.workflows.Save(ctx, f.workflow) },
		"workflow/delete-uids": func(f persistenceFixture, ctx context.Context) error {
			return f.workflows.DeleteByUIDs(ctx, []string{"target"})
		},
		"workflow/ttl": func(f persistenceFixture, ctx context.Context) error {
			return f.workflows.DeleteByFinishTime(ctx, time.Hour)
		},
		"experiment/list-managed": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.experiments.ListArchivedManagedByNamespaceName(ctx, "team", "target")
			return err
		},
		"workflow/list": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.workflows.List(ctx, "", "", true)
			return err
		},
		"workflow/find-id": func(f persistenceFixture, ctx context.Context) error {
			_, err := f.workflows.FindByID(ctx, f.workflow.ID)
			return err
		},
		"workflow/delete-uid": func(f persistenceFixture, ctx context.Context) error { return f.workflows.DeleteByUID(ctx, "target") },
		"workflow/archive": func(f persistenceFixture, ctx context.Context) error {
			return f.workflows.MarkAsArchived(ctx, "team", "target")
		},
		"workflow/archive-uid": func(f persistenceFixture, ctx context.Context) error {
			return f.workflows.MarkAsArchivedWithUID(ctx, "target")
		},
	}
	for name, operation := range operations {
		for _, expired := range []bool{false, true} {
			state := "canceled"
			if expired {
				state = "deadline"
			}
			t.Run(name+"/"+state, func(t *testing.T) {
				f := newPersistenceFixture(t)
				ctx, cancel := context.WithCancel(context.Background())
				want := context.Canceled
				if expired {
					cancel()
					ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
					want = context.DeadlineExceeded
				}
				cancel()
				require.ErrorIs(t, operation(f, ctx), want)
				for _, row := range []any{&core.Event{}, &core.Experiment{}, &core.Schedule{}, &core.WorkflowEntity{}} {
					var count int64
					require.NoError(t, f.db.Model(row).Count(&count).Error)
					require.EqualValues(t, 1, count)
				}
				require.NoError(t, operation(f, context.Background()))
			})
		}
	}
}

type retentionCase struct {
	name    string
	model   any
	cleanup func(context.Context, time.Duration) error
}

func retentionCases(f persistenceFixture) []retentionCase {
	return []retentionCase{{"experiment", &core.Experiment{}, f.experiments.DeleteByFinishTime}, {"schedule", &core.Schedule{}, f.schedules.DeleteByFinishTime}, {"workflow", &core.WorkflowEntity{}, f.workflows.DeleteByFinishTime}}
}

func TestRetentionPartialFailureAndRetry(t *testing.T) {
	for _, kind := range []string{"experiment", "schedule", "workflow"} {
		t.Run(kind, func(t *testing.T) {
			f := newPersistenceFixture(t)
			var tc retentionCase
			for _, candidate := range retentionCases(f) {
				if candidate.name == kind {
					tc = candidate
				}
			}
			finish := time.Now().Add(-3 * time.Hour)
			for _, uid := range []string{"second", "third"} {
				var row any
				switch kind {
				case "experiment":
					row = &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: uid, Archived: true, FinishTime: &finish}}
				case "schedule":
					row = &core.Schedule{ScheduleMeta: core.ScheduleMeta{UID: uid, Archived: true, FinishTime: &finish}}
				case "workflow":
					row = &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: uid, Archived: true, FinishTime: &finish}}
				}
				require.NoError(t, f.db.Create(row).Error)
			}
			failure := errors.New("second delete failed")
			deletes := 0
			require.NoError(t, f.db.Callback().Delete().Before("gorm:delete").Register("test:fail-second", func(tx *gorm.DB) {
				deletes++
				if deletes == 2 {
					tx.AddError(failure)
				}
			}))
			require.ErrorIs(t, tc.cleanup(context.Background(), time.Hour), failure)
			var remaining int64
			require.NoError(t, f.db.Model(tc.model).Count(&remaining).Error)
			require.EqualValues(t, 2, remaining)
			require.NoError(t, f.db.Callback().Delete().Remove("test:fail-second"))
			for range 2 {
				require.NoError(t, tc.cleanup(context.Background(), time.Hour))
				require.NoError(t, f.db.Model(tc.model).Count(&remaining).Error)
				require.Zero(t, remaining)
			}
		})
	}
}

func TestRetentionCancellationBetweenSelectionAndDeletion(t *testing.T) {
	for _, kind := range []string{"experiment", "schedule", "workflow"} {
		t.Run(kind, func(t *testing.T) {
			f := newPersistenceFixture(t)
			var tc retentionCase
			for _, candidate := range retentionCases(f) {
				if candidate.name == kind {
					tc = candidate
				}
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			require.NoError(t, f.db.Callback().Query().After("gorm:query").Register("test:cancel-after-selection", func(tx *gorm.DB) { cancel() }))
			require.ErrorIs(t, tc.cleanup(ctx, time.Hour), context.Canceled)
			require.NoError(t, f.db.Callback().Query().Remove("test:cancel-after-selection"))
			var remaining int64
			require.NoError(t, f.db.Model(tc.model).Count(&remaining).Error)
			require.EqualValues(t, 1, remaining)
			require.NoError(t, tc.cleanup(context.Background(), time.Hour))
			require.NoError(t, f.db.Model(tc.model).Count(&remaining).Error)
			require.Zero(t, remaining)
		})
	}
}

func TestRetentionExactBoundary(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		f := newPersistenceFixture(t)
		now := time.Now().UTC()
		for _, tc := range retentionCases(f) {
			require.NoError(t, f.db.Model(tc.model).Where("uid = ?", "target").Update("finish_time", now.Add(-time.Hour)).Error)
			require.NoError(t, tc.cleanup(context.Background(), time.Hour))
			var remaining int64
			require.NoError(t, f.db.Model(tc.model).Count(&remaining).Error)
			require.EqualValues(t, 1, remaining, "%s exact boundary", tc.name)
			require.NoError(t, f.db.Model(tc.model).Where("uid = ?", "target").Update("finish_time", now.Add(-time.Hour-time.Nanosecond)).Error)
			require.NoError(t, tc.cleanup(context.Background(), time.Hour))
			require.NoError(t, f.db.Model(tc.model).Count(&remaining).Error)
			require.Zero(t, remaining, "%s expired boundary", tc.name)
		}
		require.NoError(t, f.db.Model(&core.Event{}).Where("object_id = ?", "target").Update("created_at", now.Add(-time.Hour)).Error)
		require.NoError(t, f.events.Create(context.Background(), &core.Event{ObjectID: "future", CreatedAt: now.Add(-time.Hour + time.Nanosecond)}))
		require.NoError(t, f.events.DeleteByDuration(context.Background(), time.Hour))
		rows, err := f.events.List(context.Background())
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, "future", rows[0].ObjectID)
	})
}

func TestRetentionForwardsContextToSelection(t *testing.T) {
	for _, kind := range []string{"experiment", "schedule", "workflow"} {
		t.Run(kind, func(t *testing.T) {
			f := newPersistenceFixture(t)
			var tc retentionCase
			for _, candidate := range retentionCases(f) {
				if candidate.name == kind {
					tc = candidate
				}
			}
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "cleanup")
			require.NoError(t, f.db.Callback().Query().Before("gorm:query").Register("test:context", func(tx *gorm.DB) {
				if tx.Statement.Context.Value(contextKey{}) != "cleanup" {
					tx.AddError(errors.New("selection lost request context"))
				}
			}))
			require.NoError(t, tc.cleanup(ctx, time.Hour))
			require.NoError(t, f.db.Callback().Query().Remove("test:context"))
			var remaining int64
			require.NoError(t, f.db.Model(tc.model).Count(&remaining).Error)
			require.Zero(t, remaining)
		})
	}
}
