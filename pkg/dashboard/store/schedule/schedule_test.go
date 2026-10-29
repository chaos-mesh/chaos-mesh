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

package schedule

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func newTestStore(t *testing.T) (core.ScheduleStore, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	store, err := NewStore(db)
	require.NoError(t, err)
	return store, db
}

func saveRecords(t *testing.T, store core.ScheduleStore, records ...*core.Schedule) {
	t.Helper()
	for _, record := range records {
		require.NoError(t, store.Save(context.Background(), record))
		require.NotZero(t, record.ID)
	}
}

func TestStoreSaveAndFind(t *testing.T) {
	store, db := newTestStore(t)
	ctx := context.Background()
	record := &core.Schedule{UID: "target", Namespace: "team-a", Name: "chaos", Archived: true, Schedule: `{"version":1}`}
	other := &core.Schedule{UID: "other", Namespace: "team-b", Name: "chaos", Schedule: `{"version":2}`}
	saveRecords(t, store, record, other)
	found, err := store.FindByUID(ctx, record.UID)
	require.NoError(t, err)
	require.Equal(t, record.ID, found.ID)
	require.Equal(t, record.Schedule, found.Schedule)
	require.Equal(t, record.Namespace, found.Namespace)
	require.True(t, found.Archived)
	metadata, err := store.FindMetaByUID(ctx, record.UID)
	require.NoError(t, err)
	require.Equal(t, found.ScheduleMeta, *metadata)

	record.Schedule = `{"version":3}`
	record.Name = ""
	record.Archived = false
	saveRecords(t, store, record)
	found, err = store.FindByUID(ctx, record.UID)
	require.NoError(t, err)
	require.Equal(t, record.ID, found.ID)
	require.Equal(t, record.Schedule, found.Schedule)
	require.Empty(t, found.Name)
	require.False(t, found.Archived)
	var count int64
	require.NoError(t, db.Model(&core.Schedule{}).Count(&count).Error)
	require.EqualValues(t, 2, count)
	found, err = store.FindByUID(ctx, other.UID)
	require.NoError(t, err)
	require.Equal(t, other.Schedule, found.Schedule)
	found, err = store.FindByUID(ctx, "missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, found)
	metadata, err = store.FindMetaByUID(ctx, "missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, metadata)
}

func TestStoreListMetaFilters(t *testing.T) {
	store, _ := newTestStore(t)
	records := []*core.Schedule{
		{UID: "match", Namespace: "team-a", Name: "chaos"},
		{UID: "namespace", Namespace: "team-b", Name: "chaos"},
		{UID: "name", Namespace: "team-a", Name: "other"},
		{UID: "archived", Namespace: "team-a", Name: "chaos", Archived: true},
		{UID: "quoted", Namespace: "team-a", Name: "chaos' OR 1=1 --"},
	}

	saveRecords(t, store, records...)
	cases := []struct {
		name, namespace, resourceName string
		archived                      bool
		want                          []string
	}{
		{"all active", "", "", false, []string{"match", "name", "namespace", "quoted"}},
		{"namespace", "team-a", "", false, []string{"match", "name", "quoted"}},
		{"name", "", "chaos", false, []string{"match", "namespace"}},
		{"combined", "team-a", "chaos", false, []string{"match"}},
		{"archived", "team-a", "chaos", true, []string{"archived"}},
		{"no matches", "missing", "", false, []string{}},
		{"literal name", "team-a", "chaos' OR 1=1 --", false, []string{"quoted"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			records, err := store.ListMeta(context.Background(), tc.namespace, tc.resourceName, tc.archived)
			require.NoError(t, err)
			got := make([]string, 0, len(records))
			for _, record := range records {
				got = append(got, record.UID)
			}
			sort.Strings(got)
			require.Equal(t, tc.want, got)
		})
	}

}

func TestStoreArchive(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	finished := time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Second)
	records := []*core.Schedule{
		{UID: "unfinished", Namespace: "team-a", Name: "chaos"},
		{UID: "finished", Namespace: "team-a", Name: "chaos", FinishTime: &finished},
		{UID: "archived", Namespace: "team-a", Name: "chaos", Archived: true, FinishTime: &finished},
		{UID: "archived unfinished", Namespace: "team-a", Name: "chaos", Archived: true},
		{UID: "namespace", Namespace: "team-b", Name: "chaos"},
		{UID: "name", Namespace: "team-a", Name: "other"},
	}
	saveRecords(t, store, records...)
	before := time.Now().Add(-time.Second)
	require.NoError(t, store.Archive(ctx, "team-a", "chaos"))
	unfinished, err := store.FindByUID(ctx, "unfinished")
	require.NoError(t, err)
	require.True(t, unfinished.Archived)
	require.NotNil(t, unfinished.FinishTime)
	require.False(t, unfinished.FinishTime.Before(before))
	require.False(t, unfinished.FinishTime.After(time.Now()))
	require.NoError(t, store.Archive(ctx, "team-a", "chaos"))
	require.NoError(t, store.Archive(ctx, "missing", "chaos"))
	for _, original := range records {
		found, err := store.FindByUID(ctx, original.UID)
		require.NoError(t, err)
		switch original.UID {
		case "unfinished":
			require.True(t, found.FinishTime.Equal(*unfinished.FinishTime))
		case "archived unfinished":
			require.True(t, found.Archived)
			require.Nil(t, found.FinishTime)
		case "finished", "archived":
			require.True(t, found.Archived)
			require.NotNil(t, found.FinishTime)
			require.True(t, found.FinishTime.Equal(finished))
		default:
			require.False(t, found.Archived)
			require.Nil(t, found.FinishTime)
		}
	}
}

func TestStoreRetention(t *testing.T) {
	store, db := newTestStore(t)
	ctx := context.Background()
	old := time.Now().Add(-48 * time.Hour)
	recent := time.Now().Add(-time.Minute)
	records := []*core.Schedule{
		{UID: "expired", Archived: true, FinishTime: &old},
		{UID: "recent", Archived: true, FinishTime: &recent},
		{UID: "unfinished", Archived: true},
		{UID: "active", FinishTime: &old},
	}
	saveRecords(t, store, records...)
	require.NoError(t, store.DeleteByFinishTime(ctx, 24*time.Hour))
	require.NoError(t, store.DeleteByFinishTime(ctx, 24*time.Hour))
	var remaining []core.Schedule
	require.NoError(t, db.Unscoped().Find(&remaining).Error)
	got := make([]string, 0, len(remaining))
	for _, record := range remaining {
		got = append(got, record.UID)
	}
	require.ElementsMatch(t, []string{"recent", "unfinished", "active"}, got)
}

func TestStoreDelete(t *testing.T) {
	for _, operation := range []string{"empty uid list", "uid list", "single", "incomplete"} {
		t.Run(operation, func(t *testing.T) {
			store, db := newTestStore(t)
			ctx := context.Background()
			finished := time.Now().Add(-time.Hour)
			first := &core.Schedule{UID: "first", FinishTime: &finished}
			second := &core.Schedule{UID: "second", Archived: true}
			third := &core.Schedule{UID: "third"}
			saveRecords(t, store, first, second, third)
			var want []string
			switch operation {
			case "empty uid list":
				require.NoError(t, store.DeleteByUIDs(ctx, nil))
				require.NoError(t, store.DeleteByUIDs(ctx, []string{}))
				want = []string{"first", "second", "third"}
			case "uid list":
				require.NoError(t, store.DeleteByUIDs(ctx, []string{"first", "second", "missing"}))
				want = []string{"third"}
			case "single":
				require.NoError(t, store.Delete(ctx, first))
				require.NoError(t, store.Delete(ctx, first))
				want = []string{"second", "third"}
			case "incomplete":
				require.NoError(t, store.DeleteIncompleteSchedules(ctx))
				want = []string{"first"}
			}
			var remaining []core.Schedule
			require.NoError(t, db.Unscoped().Find(&remaining).Error)
			got := make([]string, 0, len(remaining))
			for _, record := range remaining {
				got = append(got, record.UID)
			}
			require.ElementsMatch(t, want, got)
		})
	}
}

func TestStoreDatabaseErrors(t *testing.T) {
	store, db := newTestStore(t)
	require.NoError(t, db.Migrator().DropTable(&core.Schedule{}))
	ctx := context.Background()
	_, err := store.ListMeta(ctx, "", "", false)
	require.Error(t, err)
	_, err = store.FindByUID(ctx, "missing")
	require.Error(t, err)
	require.False(t, errors.Is(err, gorm.ErrRecordNotFound))
	_, err = store.FindMetaByUID(ctx, "missing")
	require.Error(t, err)
	require.Error(t, store.Save(ctx, &core.Schedule{}))
	require.Error(t, store.Archive(ctx, "team-a", "chaos"))
	require.Error(t, store.DeleteByUIDs(ctx, []string{"first"}))
	require.Error(t, store.DeleteByFinishTime(ctx, time.Hour))
	require.Error(t, store.DeleteIncompleteSchedules(ctx))
	require.Error(t, store.Delete(ctx, &core.Schedule{ID: 1}))
}

func TestStoreMigrationError(t *testing.T) {
	_, db := newTestStore(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	store, err := NewStore(db)
	require.Error(t, err)
	require.Nil(t, store)
}

func TestStoreRetentionDeleteError(t *testing.T) {
	store, db := newTestStore(t)
	expired := time.Now().Add(-48 * time.Hour)
	record := &core.Schedule{UID: "expired", Archived: true, FinishTime: &expired}
	saveRecords(t, store, record)
	require.NoError(t, db.Exec("CREATE TRIGGER reject_delete BEFORE DELETE ON schedules BEGIN SELECT RAISE(ABORT, 'delete rejected'); END").Error)
	require.Error(t, store.DeleteByFinishTime(context.Background(), time.Hour))
	found, err := store.FindByUID(context.Background(), record.UID)
	require.NoError(t, err)
	require.Equal(t, record.ID, found.ID)
}

func TestStartupDeletesIncompleteRecords(t *testing.T) {
	store, _ := newTestStore(t)
	finished := time.Now().Add(-time.Hour)
	incomplete := &core.Schedule{UID: "incomplete"}
	complete := &core.Schedule{UID: "complete", FinishTime: &finished}
	saveRecords(t, store, incomplete, complete)
	DeleteIncompleteSchedules(store, nil)
	_, err := store.FindByUID(context.Background(), incomplete.UID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	found, err := store.FindByUID(context.Background(), complete.UID)
	require.NoError(t, err)
	require.Equal(t, complete.ID, found.ID)
}
