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
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func sourceEvent(uid string) *core.Event {
	return &core.Event{
		SourceUID: &uid, ObjectID: "chaos-uid", Namespace: "default", Name: "target",
		CreatedAt: time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC),
		Kind:      "PodChaos", Type: "Normal", Reason: "Applied", Message: "fault injected",
	}
}

func TestEventSourceUIDReplay(t *testing.T) {
	store, db := newEventTestStore(t)
	ctx := context.Background()
	first := sourceEvent("source-1")
	require.NoError(t, store.Create(ctx, first))
	replay := sourceEvent("source-1")
	replay.Message = "changed after initial archival"
	require.NoError(t, store.Create(ctx, replay))
	require.Equal(t, first, replay, "replay must return the canonical persisted event")
	require.Error(t, db.Create(sourceEvent("source-1")).Error, "the database must arbitrate competing inserts")
	second := sourceEvent("source-2")
	require.NoError(t, store.Create(ctx, second))
	rows, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 2, "distinct source events with identical content must remain distinct")
	require.NotEqual(t, first.ID, second.ID)
}

func TestEventSourceUIDUnknownAppends(t *testing.T) {
	store, db := newEventTestStore(t)
	for _, unknown := range []*string{nil, nil, new(string), new(string)} {
		event := sourceEvent("unused")
		event.SourceUID = unknown
		require.NoError(t, store.Create(context.Background(), event))
		require.Nil(t, event.SourceUID)
	}
	var count int64
	require.NoError(t, db.Model(&core.Event{}).Where("source_uid IS NULL").Count(&count).Error)
	require.EqualValues(t, 4, count)
}

func TestEventSourceUIDLostCommitAcknowledgement(t *testing.T) {
	for _, failure := range []string{"acknowledgement", "lookup", "cancelled lookup"} {
		t.Run(failure, func(t *testing.T) {
			store, db := newEventTestStore(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			lostAck := errors.New("commit acknowledgement lost")
			loseNext := true
			require.NoError(t, db.Callback().Create().After("gorm:commit_or_rollback_transaction").Register("test:lose_ack", func(tx *gorm.DB) {
				if tx.Error == nil && loseNext {
					loseNext = false
					tx.AddError(lostAck)
					if failure == "cancelled lookup" {
						cancel()
					}
				}
			}))
			if failure == "lookup" {
				require.NoError(t, db.Callback().Query().Before("gorm:query").Register("test:fail_lookup", func(tx *gorm.DB) {
					tx.AddError(errors.New("lookup unavailable"))
				}))
			}
			event := sourceEvent("source-1")
			err := store.Create(ctx, event)
			if failure == "acknowledgement" {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, lostAck, "a failed recovery must preserve the insert error")
			}
			if failure == "lookup" {
				require.NoError(t, db.Callback().Query().Remove("test:fail_lookup"))
			}
			rows, err := store.List(context.Background())
			require.NoError(t, err)
			require.Len(t, rows, 1, "the acknowledgement failure must occur after a real commit")
			require.Equal(t, event.ID, rows[0].ID)
			replay := sourceEvent("source-1")
			require.NoError(t, store.Create(context.Background(), replay))
			require.Equal(t, rows[0], replay)
			rows, err = store.List(context.Background())
			require.NoError(t, err)
			require.Len(t, rows, 1)
		})
	}
}

func TestEventSourceUIDInsertFailure(t *testing.T) {
	store, db := newEventTestStore(t)
	require.NoError(t, db.Exec("CREATE TRIGGER reject_event BEFORE INSERT ON events BEGIN SELECT RAISE(ABORT, 'event rejected'); END").Error)
	var insertError error
	require.NoError(t, db.Callback().Create().After("gorm:commit_or_rollback_transaction").Register("test:capture_error", func(tx *gorm.DB) {
		insertError = tx.Error
	}))
	for _, uid := range []*string{nil, sourceEvent("source-1").SourceUID} {
		event := sourceEvent("unused")
		event.SourceUID = uid
		err := store.Create(context.Background(), event)
		require.ErrorContains(t, err, "event rejected")
		require.Same(t, insertError, err, "lookup errors must not replace the original insert failure")
	}
	rows, err := store.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, rows)
}

func TestEventSourceUIDCancelledContext(t *testing.T) {
	store, _ := newEventTestStore(t)
	require.NoError(t, store.Create(context.Background(), sourceEvent("existing")))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, uid := range []string{"existing", "new"} {
		require.ErrorIs(t, store.Create(ctx, sourceEvent(uid)), context.Canceled)
	}
	rows, err := store.List(context.Background())
	require.NoError(t, err)
	require.Len(t, rows, 1)
}

func TestEventSourceUIDConcurrentReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.db")
	const writers = 8
	stores := make([]core.EventStore, writers)
	for i := range stores {
		db, err := gorm.Open(sqlite.Open(path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		require.NoError(t, err)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.SetMaxOpenConns(1)
		t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
		stores[i], err = NewStore(db)
		require.NoError(t, err)
	}
	start := make(chan struct{})
	results := make(chan *core.Event, writers)
	failures := make(chan error, writers)
	var workers sync.WaitGroup
	for _, store := range stores {
		workers.Go(func() {
			<-start
			event := sourceEvent("same-source")
			failures <- store.Create(context.Background(), event)
			results <- event
		})
	}
	close(start)
	workers.Wait()
	close(failures)
	close(results)
	for err := range failures {
		require.NoError(t, err)
	}
	rows, err := stores[0].List(context.Background())
	require.NoError(t, err)
	require.Len(t, rows, 1)
	for event := range results {
		require.Equal(t, rows[0], event)
	}
}

func TestEventSourceUIDJSONUnchanged(t *testing.T) {
	event := sourceEvent("private-source-uid")
	withUID, err := json.Marshal(event)
	require.NoError(t, err)
	event.SourceUID = nil
	withoutUID, err := json.Marshal(event)
	require.NoError(t, err)
	require.JSONEq(t, string(withoutUID), string(withUID))
	var fields map[string]any
	require.NoError(t, json.Unmarshal(withUID, &fields))
	require.Len(t, fields, 9)
	require.NotContains(t, string(withUID), "private-source-uid")
}

func TestEventSourceUIDLegacyMigration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.Exec("CREATE TABLE events (id integer PRIMARY KEY AUTOINCREMENT, object_id text, created_at datetime, namespace text, name text, kind text, type text, reason text, message text)").Error)
	require.NoError(t, db.Exec("CREATE INDEX object_id ON events(object_id)").Error)
	for i, timestamp := range []string{"2026-10-16 08:00:00+08:00", "2026-10-15 19:00:00-05:00", "2026-10-16T00:00:00Z", "2026-10-16T00:00:00Z"} {
		require.NoError(t, db.Exec("INSERT INTO events VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", i+11, "chaos-uid", timestamp, "default", "target", "PodChaos", "Normal", "Applied", "fault injected").Error)
	}
	type legacyRow struct {
		ID                                                                uint
		ObjectID, Timestamp, Namespace, Name, Kind, Type, Reason, Message string
	}
	readHistory := func() []legacyRow {
		var rows []legacyRow
		require.NoError(t, db.Raw("SELECT id, object_id, CAST(created_at AS TEXT) AS timestamp, namespace, name, kind, type, reason, message FROM events WHERE id BETWEEN 11 AND 14 ORDER BY id").Scan(&rows).Error)
		return rows
	}
	before := readHistory()
	require.Len(t, before, 4)
	columns, err := db.Migrator().ColumnTypes("events")
	require.NoError(t, err)
	require.Len(t, columns, 9)
	require.NoError(t, db.Exec("CREATE TABLE event_updates (id integer)").Error)
	require.NoError(t, db.Exec("CREATE TRIGGER audit_event_update AFTER UPDATE ON events BEGIN INSERT INTO event_updates VALUES (NEW.id); END").Error)
	for range 2 {
		store, err := NewStore(db)
		require.NoError(t, err)
		require.Equal(t, before, readHistory(), "migration must preserve every historical value including raw timestamp text")
		var known int64
		require.NoError(t, db.Model(&core.Event{}).Where("id BETWEEN 11 AND 14 AND source_uid IS NOT NULL").Count(&known).Error)
		require.Zero(t, known, "historical identity must remain unknown")
		var updates int64
		require.NoError(t, db.Table("event_updates").Count(&updates).Error)
		require.Zero(t, updates)
		require.NoError(t, store.Create(context.Background(), sourceEvent("new-source")))
	}
	var count int64
	require.NoError(t, db.Model(&core.Event{}).Count(&count).Error)
	require.EqualValues(t, 5, count, "historical duplicates stay intact while new identified events converge")
	var indexSQL string
	require.NoError(t, db.Raw("SELECT sql FROM sqlite_master WHERE type = 'index' AND name = 'event_source_uid'").Scan(&indexSQL).Error)
	require.Equal(t, "CREATE UNIQUE INDEX `event_source_uid` ON `events`(`source_uid`) WHERE source_uid IS NOT NULL", indexSQL)
}

func TestEventSourceUIDDialectMigration(t *testing.T) {
	for _, tc := range []struct {
		name      string
		dialect   func(*sql.DB) gorm.Dialector
		addColumn string
		statement string
	}{
		{
			"mysql", func(db *sql.DB) gorm.Dialector {
				return mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true})
			},
			"ALTER TABLE `events` ADD `source_uid` varchar(128)",
			"CREATE UNIQUE INDEX `event_source_uid` ON `events`(`source_uid`)",
		},
		{
			"postgres", func(db *sql.DB) gorm.Dialector { return postgres.New(postgres.Config{Conn: db}) },
			`ALTER TABLE "events" ADD "source_uid" varchar(128)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS "event_source_uid" ON "events" ("source_uid") WHERE source_uid IS NOT NULL`,
		},
		{
			"sqlserver", func(db *sql.DB) gorm.Dialector { return sqlserver.New(sqlserver.Config{Conn: db}) },
			`ALTER TABLE "events" ADD "source_uid" nvarchar(128)`,
			`CREATE UNIQUE INDEX "event_source_uid" ON "events"("source_uid") WHERE source_uid IS NOT NULL`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer sqlDB.Close()
			db, err := gorm.Open(tc.dialect(sqlDB), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
			require.NoError(t, err)
			mock.ExpectExec(tc.addColumn).WillReturnResult(sqlmock.NewResult(0, 0))
			require.NoError(t, db.Migrator().AddColumn(&core.Event{}, "SourceUID"))
			mock.ExpectExec(tc.statement).WillReturnResult(sqlmock.NewResult(0, 0))
			require.NoError(t, db.Migrator().CreateIndex(&core.Event{}, "event_source_uid"))
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
