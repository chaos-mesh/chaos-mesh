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

package event

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func newFileEventTestStore(t *testing.T) (core.EventStore, *gorm.DB, *gorm.DB) {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "events.db")
	open := func() *gorm.DB {
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		require.NoError(t, err)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.SetMaxOpenConns(1)
		t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
		return db
	}
	db, other := open(), open()
	require.NoError(t, db.Exec("PRAGMA journal_mode=WAL").Error)
	store, err := NewStore(db)
	require.NoError(t, err)
	return store, db, other
}

func TestEventTimeCancellation(t *testing.T) {
	for _, operation := range []string{"normal list", "zero limit", "reversed bounds", "cleanup"} {
		t.Run(operation, func(t *testing.T) {
			store, _ := newEventTestStore(t)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			filter := core.Filter{Start: "2026-10-16 00:00:00"}
			if operation == "zero limit" {
				filter.Limit = "0"
			}
			if operation == "reversed bounds" {
				filter.End = "2026-10-15 00:00:00"
			}
			if operation == "cleanup" {
				require.ErrorIs(t, store.DeleteByDuration(ctx, time.Hour), context.Canceled)
				return
			}
			got, err := store.ListByFilter(ctx, filter)
			require.ErrorIs(t, err, context.Canceled)
			require.Nil(t, got)
		})
	}
	t.Run("during cleanup", func(t *testing.T) {
		store, db, other := newFileEventTestStore(t)
		for id := 1; id <= 600; id++ {
			require.NoError(t, db.Exec("INSERT INTO events (id, created_at) VALUES (?, ?)", id, "2020-01-01T00:00:00Z").Error)
		}
		require.NoError(t, db.Exec("INSERT INTO events (id, created_at) VALUES (601, NULL)").Error)
		before := storedEventTimes(t, db)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		require.NoError(t, db.Callback().Delete().After("gorm:delete").Register("cancel_cleanup", func(*gorm.DB) { cancel() }))
		require.Error(t, store.DeleteByDuration(ctx, time.Hour))
		require.ErrorIs(t, ctx.Err(), context.Canceled)
		require.Equal(t, before, storedEventTimes(t, other))
		var nullCount int
		require.NoError(t, other.Raw("SELECT COUNT(*) FROM events WHERE id = 601 AND created_at IS NULL").Scan(&nullCount).Error)
		require.Equal(t, 1, nullCount)
		require.NoError(t, db.Callback().Delete().Remove("cancel_cleanup"))
		require.NoError(t, store.DeleteByDuration(context.Background(), time.Hour))
		got, err := store.List(context.Background())
		require.NoError(t, err)
		require.Equal(t, []uint{601}, eventIDs(got))
	})
}

func TestEventTimeFilteringSnapshot(t *testing.T) {
	store, db, other := newFileEventTestStore(t)
	for id := 1; id <= 600; id++ {
		require.NoError(t, db.Exec("INSERT INTO events (id, created_at, message) VALUES (?, ?, ?)", id, "2026-10-16 08:00:00+08:00", "original").Error)
	}
	changed := false
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("concurrent_writer", func(tx *gorm.DB) {
		if changed || tx.Statement.Table != "events" {
			return
		}
		changed = true
		require.NoError(t, other.Exec("DELETE FROM events WHERE id = 550").Error)
		require.NoError(t, other.Exec("UPDATE events SET message = 'updated' WHERE id = 400").Error)
		require.NoError(t, other.Exec("INSERT INTO events (id, created_at, message) VALUES (9999, '2026-10-16T00:00:00Z', 'new')").Error)
	}))
	got, err := store.ListByFilter(context.Background(), core.Filter{Start: "2026-10-16 00:00:00", End: "2026-10-16 00:00:00", Limit: "258"})
	require.NoError(t, err)
	require.True(t, changed)
	require.Len(t, got, 258)
	for i, event := range got {
		require.Equal(t, uint(600-i), event.ID)
		require.Equal(t, "original", event.Message)
	}
	var deleted int
	require.NoError(t, other.Raw("SELECT COUNT(*) FROM events WHERE id = 550").Scan(&deleted).Error)
	require.Zero(t, deleted)
}
