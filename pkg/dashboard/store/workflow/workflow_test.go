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

package workflow

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

func newTestStore(t *testing.T) (core.WorkflowStore, *gorm.DB) {
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

func saveRecords(t *testing.T, store core.WorkflowStore, records ...*core.WorkflowEntity) {
	t.Helper()
	for _, record := range records {
		require.NoError(t, store.Save(context.Background(), record))
		require.NotZero(t, record.ID)
	}
}

func TestStoreSaveAndFind(t *testing.T) {
	store, db := newTestStore(t)
	ctx := context.Background()
	record := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "target", Namespace: "team-a", Name: "chaos", Archived: true}, Workflow: `{"version":1}`}
	other := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "other", Namespace: "team-b", Name: "chaos"}, Workflow: `{"version":2}`}
	saveRecords(t, store, record, other)
	found, err := store.FindByUID(ctx, record.UID)
	require.NoError(t, err)
	require.Equal(t, record.ID, found.ID)
	require.Equal(t, record.Workflow, found.Workflow)
	require.Equal(t, record.Namespace, found.Namespace)
	require.True(t, found.Archived)
	metadata, err := store.FindMetaByUID(ctx, record.UID)
	require.NoError(t, err)
	require.Equal(t, found.WorkflowMeta, *metadata)

	record.Workflow = `{"version":3}`
	record.Name = ""
	record.Archived = false
	saveRecords(t, store, record)
	found, err = store.FindByUID(ctx, record.UID)
	require.NoError(t, err)
	require.Equal(t, record.ID, found.ID)
	require.Equal(t, record.Workflow, found.Workflow)
	require.Empty(t, found.Name)
	require.False(t, found.Archived)
	var count int64
	require.NoError(t, db.Model(&core.WorkflowEntity{}).Count(&count).Error)
	require.EqualValues(t, 2, count)
	found, err = store.FindByUID(ctx, other.UID)
	require.NoError(t, err)
	require.Equal(t, other.Workflow, found.Workflow)
	found, err = store.FindByUID(ctx, "missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, found)
	metadata, err = store.FindMetaByUID(ctx, "missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, metadata)
}

func TestStoreListMetaFilters(t *testing.T) {
	store, _ := newTestStore(t)
	records := []*core.WorkflowEntity{
		{WorkflowMeta: core.WorkflowMeta{UID: "match", Namespace: "team-a", Name: "chaos"}},
		{WorkflowMeta: core.WorkflowMeta{UID: "namespace", Namespace: "team-b", Name: "chaos"}},
		{WorkflowMeta: core.WorkflowMeta{UID: "name", Namespace: "team-a", Name: "other"}},
		{WorkflowMeta: core.WorkflowMeta{UID: "archived", Namespace: "team-a", Name: "chaos", Archived: true}},
		{WorkflowMeta: core.WorkflowMeta{UID: "quoted", Namespace: "team-a", Name: "chaos' OR 1=1 --"}},
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
	records := []*core.WorkflowEntity{
		{WorkflowMeta: core.WorkflowMeta{UID: "unfinished", Namespace: "team-a", Name: "chaos"}},
		{WorkflowMeta: core.WorkflowMeta{UID: "finished", Namespace: "team-a", Name: "chaos", FinishTime: &finished}},
		{WorkflowMeta: core.WorkflowMeta{UID: "archived", Namespace: "team-a", Name: "chaos", Archived: true, FinishTime: &finished}},
		{WorkflowMeta: core.WorkflowMeta{UID: "archived unfinished", Namespace: "team-a", Name: "chaos", Archived: true}},
		{WorkflowMeta: core.WorkflowMeta{UID: "namespace", Namespace: "team-b", Name: "chaos"}},
		{WorkflowMeta: core.WorkflowMeta{UID: "name", Namespace: "team-a", Name: "other"}},
	}
	saveRecords(t, store, records...)
	before := time.Now().Add(-time.Second)
	require.NoError(t, store.MarkAsArchived(ctx, "team-a", "chaos"))
	unfinished, err := store.FindByUID(ctx, "unfinished")
	require.NoError(t, err)
	require.True(t, unfinished.Archived)
	require.NotNil(t, unfinished.FinishTime)
	require.False(t, unfinished.FinishTime.Before(before))
	require.False(t, unfinished.FinishTime.After(time.Now()))
	require.NoError(t, store.MarkAsArchived(ctx, "team-a", "chaos"))
	require.NoError(t, store.MarkAsArchived(ctx, "missing", "chaos"))
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
	records := []*core.WorkflowEntity{
		{WorkflowMeta: core.WorkflowMeta{UID: "expired", Archived: true, FinishTime: &old}},
		{WorkflowMeta: core.WorkflowMeta{UID: "recent", Archived: true, FinishTime: &recent}},
		{WorkflowMeta: core.WorkflowMeta{UID: "unfinished", Archived: true}},
		{WorkflowMeta: core.WorkflowMeta{UID: "active", FinishTime: &old}},
	}
	saveRecords(t, store, records...)
	require.NoError(t, store.DeleteByFinishTime(ctx, 24*time.Hour))
	require.NoError(t, store.DeleteByFinishTime(ctx, 24*time.Hour))
	var remaining []core.WorkflowEntity
	require.NoError(t, db.Unscoped().Find(&remaining).Error)
	got := make([]string, 0, len(remaining))
	for _, record := range remaining {
		got = append(got, record.UID)
	}
	require.ElementsMatch(t, []string{"recent", "unfinished", "active"}, got)
}

func TestStoreDelete(t *testing.T) {
	for _, operation := range []string{"empty uid list", "uid list", "single"} {
		t.Run(operation, func(t *testing.T) {
			store, db := newTestStore(t)
			ctx := context.Background()
			finished := time.Now().Add(-time.Hour)
			first := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "first", FinishTime: &finished}}
			second := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "second", Archived: true}}
			third := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "third"}}
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
				require.NoError(t, store.DeleteByUID(ctx, first.UID))
				require.NoError(t, store.DeleteByUID(ctx, first.UID))
				want = []string{"second", "third"}

			}
			var remaining []core.WorkflowEntity
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
	require.NoError(t, db.Migrator().DropTable(&core.WorkflowEntity{}))
	ctx := context.Background()
	_, err := store.ListMeta(ctx, "", "", false)
	require.Error(t, err)
	_, err = store.FindByUID(ctx, "missing")
	require.Error(t, err)
	require.False(t, errors.Is(err, gorm.ErrRecordNotFound))
	_, err = store.FindMetaByUID(ctx, "missing")
	require.Error(t, err)
	require.Error(t, store.Save(ctx, &core.WorkflowEntity{}))
	require.Error(t, store.MarkAsArchived(ctx, "team-a", "chaos"))
	require.Error(t, store.DeleteByUIDs(ctx, []string{"first"}))
	require.Error(t, store.DeleteByFinishTime(ctx, time.Hour))
	require.Error(t, store.DeleteByUID(ctx, "missing"))
	require.Error(t, store.MarkAsArchivedWithUID(ctx, "missing"))
	_, err = store.FindByID(ctx, 1)
	require.Error(t, err)
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

func TestStoreArchiveByUIDPreservesCompletionAndSetsFinishTime(t *testing.T) {
	for _, hasFinishTime := range []bool{false, true} {
		t.Run(map[bool]string{false: "unfinished", true: "finished"}[hasFinishTime], func(t *testing.T) {
			store, _ := newTestStore(t)
			ctx := context.Background()
			completed := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
			finished := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
			target := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "target", Namespace: "team-a", Name: "chaos", EndTime: completed}}
			other := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "other", Namespace: "team-a", Name: "chaos"}}
			if hasFinishTime {
				target.FinishTime = &finished
			}
			saveRecords(t, store, target, other)
			before := time.Now().Add(-time.Second)
			require.NoError(t, store.MarkAsArchivedWithUID(ctx, target.UID))
			found, err := store.FindByUID(ctx, target.UID)
			require.NoError(t, err)
			require.True(t, found.Archived)
			require.Equal(t, completed, found.EndTime)
			require.NotNil(t, found.FinishTime)
			if hasFinishTime {
				require.True(t, found.FinishTime.Equal(finished))
			} else {
				require.False(t, found.FinishTime.Before(before))
				require.False(t, found.FinishTime.After(time.Now()))
			}
			require.NoError(t, store.MarkAsArchivedWithUID(ctx, target.UID))
			again, err := store.FindByUID(ctx, target.UID)
			require.NoError(t, err)
			require.Equal(t, found.WorkflowMeta, again.WorkflowMeta)
			require.NoError(t, store.MarkAsArchivedWithUID(ctx, "missing"))
			untouched, err := store.FindByUID(ctx, other.UID)
			require.NoError(t, err)
			require.False(t, untouched.Archived)
			require.Empty(t, untouched.EndTime)
			require.Nil(t, untouched.FinishTime)
		})
	}
}

func TestStoreListAndFindByID(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	first := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "first", Namespace: "team-a", Name: "chaos"}, Workflow: `{"version":1}`}
	second := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "second", Namespace: "team-a", Name: "chaos"}, Workflow: `{"version":2}`}
	archived := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "archived", Namespace: "team-a", Name: "chaos", Archived: true}}
	saveRecords(t, store, first, second, archived)
	found, err := store.FindByID(ctx, second.ID)
	require.NoError(t, err)
	require.Equal(t, second.UID, found.UID)
	require.Equal(t, second.Workflow, found.Workflow)
	found, err = store.FindByID(ctx, archived.ID+1)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, found)
	records, err := store.List(ctx, "team-a", "chaos", false)
	require.NoError(t, err)
	require.Len(t, records, 2)
	require.ElementsMatch(t, []string{first.UID, second.UID}, []string{records[0].UID, records[1].UID})
	require.ElementsMatch(t, []string{first.Workflow, second.Workflow}, []string{records[0].Workflow, records[1].Workflow})
	records[0].Name = "changed"
	require.Equal(t, "chaos", records[1].Name)
}

func TestStoreRetentionDeleteError(t *testing.T) {
	store, db := newTestStore(t)
	expired := time.Now().Add(-48 * time.Hour)
	record := &core.WorkflowEntity{WorkflowMeta: core.WorkflowMeta{UID: "expired", Archived: true, FinishTime: &expired}}
	saveRecords(t, store, record)
	require.NoError(t, db.Exec("CREATE TRIGGER reject_delete BEFORE DELETE ON workflow_entities BEGIN SELECT RAISE(ABORT, 'delete rejected'); END").Error)
	require.Error(t, store.DeleteByFinishTime(context.Background(), time.Hour))
	found, err := store.FindByUID(context.Background(), record.UID)
	require.NoError(t, err)
	require.Equal(t, record.ID, found.ID)
}
