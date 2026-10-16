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

package experiment

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

func newTestStore(t *testing.T) (core.ExperimentStore, *gorm.DB) {
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

func saveRecords(t *testing.T, store core.ExperimentStore, records ...*core.Experiment) {
	t.Helper()
	for _, record := range records {
		require.NoError(t, store.Save(context.Background(), record))
		require.NotZero(t, record.ID)
	}
}

func TestStoreSaveAndFind(t *testing.T) {
	store, db := newTestStore(t)
	ctx := context.Background()
	record := &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: "target", Namespace: "team-a", Name: "chaos", Archived: true}, Experiment: `{"version":1}`}
	other := &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: "other", Namespace: "team-b", Name: "chaos"}, Experiment: `{"version":2}`}
	saveRecords(t, store, record, other)
	found, err := store.FindByUID(ctx, record.UID)
	require.NoError(t, err)
	require.Equal(t, record.ID, found.ID)
	require.Equal(t, record.Experiment, found.Experiment)
	require.Equal(t, record.Namespace, found.Namespace)
	require.True(t, found.Archived)
	metadata, err := store.FindMetaByUID(ctx, record.UID)
	require.NoError(t, err)
	require.Equal(t, found.ExperimentMeta, *metadata)

	record.Experiment = `{"version":3}`
	record.Name = ""
	record.Archived = false
	saveRecords(t, store, record)
	found, err = store.FindByUID(ctx, record.UID)
	require.NoError(t, err)
	require.Equal(t, record.ID, found.ID)
	require.Equal(t, record.Experiment, found.Experiment)
	require.Empty(t, found.Name)
	require.False(t, found.Archived)
	var count int64
	require.NoError(t, db.Model(&core.Experiment{}).Count(&count).Error)
	require.EqualValues(t, 2, count)
	found, err = store.FindByUID(ctx, other.UID)
	require.NoError(t, err)
	require.Equal(t, other.Experiment, found.Experiment)
	found, err = store.FindByUID(ctx, "missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, found)
	metadata, err = store.FindMetaByUID(ctx, "missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, metadata)
}

func TestStoreListMetaFilters(t *testing.T) {
	store, _ := newTestStore(t)
	records := []*core.Experiment{
		{ExperimentMeta: core.ExperimentMeta{UID: "match", Namespace: "team-a", Name: "chaos"}},
		{ExperimentMeta: core.ExperimentMeta{UID: "namespace", Namespace: "team-b", Name: "chaos"}},
		{ExperimentMeta: core.ExperimentMeta{UID: "name", Namespace: "team-a", Name: "other"}},
		{ExperimentMeta: core.ExperimentMeta{UID: "archived", Namespace: "team-a", Name: "chaos", Archived: true}},
		{ExperimentMeta: core.ExperimentMeta{UID: "quoted", Namespace: "team-a", Name: "chaos' OR 1=1 --"}},
	}
	for _, record := range records {
		record.Kind = "PodChaos"
	}
	records[2].Kind = "NetworkChaos"
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
			records, err := store.ListMeta(context.Background(), "", tc.namespace, tc.resourceName, tc.archived)
			require.NoError(t, err)
			got := make([]string, 0, len(records))
			for _, record := range records {
				got = append(got, record.UID)
			}
			sort.Strings(got)
			require.Equal(t, tc.want, got)
		})
	}
	filtered, err := store.ListMeta(context.Background(), "NetworkChaos", "team-a", "", false)
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	require.Equal(t, "name", filtered[0].UID)
	filtered, err = store.ListMeta(context.Background(), "NetworkChaos", "team-b", "", false)
	require.NoError(t, err)
	require.Empty(t, filtered)
}

func TestStoreArchive(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	finished := time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Second)
	records := []*core.Experiment{
		{ExperimentMeta: core.ExperimentMeta{UID: "unfinished", Namespace: "team-a", Name: "chaos"}},
		{ExperimentMeta: core.ExperimentMeta{UID: "finished", Namespace: "team-a", Name: "chaos", FinishTime: &finished}},
		{ExperimentMeta: core.ExperimentMeta{UID: "archived", Namespace: "team-a", Name: "chaos", Archived: true, FinishTime: &finished}},
		{ExperimentMeta: core.ExperimentMeta{UID: "archived unfinished", Namespace: "team-a", Name: "chaos", Archived: true}},
		{ExperimentMeta: core.ExperimentMeta{UID: "namespace", Namespace: "team-b", Name: "chaos"}},
		{ExperimentMeta: core.ExperimentMeta{UID: "name", Namespace: "team-a", Name: "other"}},
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
	records := []*core.Experiment{
		{ExperimentMeta: core.ExperimentMeta{UID: "expired", Archived: true, FinishTime: &old}},
		{ExperimentMeta: core.ExperimentMeta{UID: "recent", Archived: true, FinishTime: &recent}},
		{ExperimentMeta: core.ExperimentMeta{UID: "unfinished", Archived: true}},
		{ExperimentMeta: core.ExperimentMeta{UID: "active", FinishTime: &old}},
	}
	saveRecords(t, store, records...)
	require.NoError(t, store.DeleteByFinishTime(ctx, 24*time.Hour))
	require.NoError(t, store.DeleteByFinishTime(ctx, 24*time.Hour))
	var remaining []core.Experiment
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
			first := &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: "first", FinishTime: &finished}}
			second := &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: "second", Archived: true}}
			third := &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: "third"}}
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
				require.NoError(t, store.DeleteIncompleteExperiments(ctx))
				want = []string{"first"}
			}
			var remaining []core.Experiment
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
	require.NoError(t, db.Migrator().DropTable(&core.Experiment{}))
	ctx := context.Background()
	_, err := store.ListMeta(ctx, "", "", "", false)
	require.Error(t, err)
	_, err = store.FindByUID(ctx, "missing")
	require.Error(t, err)
	require.False(t, errors.Is(err, gorm.ErrRecordNotFound))
	_, err = store.FindMetaByUID(ctx, "missing")
	require.Error(t, err)
	require.Error(t, store.Save(ctx, &core.Experiment{}))
	require.Error(t, store.Archive(ctx, "team-a", "chaos"))
	require.Error(t, store.DeleteByUIDs(ctx, []string{"first"}))
	require.Error(t, store.DeleteByFinishTime(ctx, time.Hour))
	require.Error(t, store.DeleteIncompleteExperiments(ctx))
	require.Error(t, store.Delete(ctx, &core.Experiment{ExperimentMeta: core.ExperimentMeta{Model: gorm.Model{ID: 1}}}))
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

func TestStoreArchivedManagedExperiments(t *testing.T) {
	store, _ := newTestStore(t)
	records := []*core.Experiment{
		{ExperimentMeta: core.ExperimentMeta{UID: "managed", Namespace: "team-a", Name: "chaos", Archived: true}, Experiment: `{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"PodChaos","metadata":{"labels":{"managed-by":"workflow"}}}`},
		{ExperimentMeta: core.ExperimentMeta{UID: "unmanaged", Namespace: "team-a", Name: "chaos", Archived: true}, Experiment: `{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"PodChaos"}`},
		{ExperimentMeta: core.ExperimentMeta{UID: "empty", Namespace: "team-a", Name: "chaos", Archived: true}, Experiment: `{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"PodChaos","metadata":{"labels":{"managed-by":""}}}`},
		{ExperimentMeta: core.ExperimentMeta{UID: "active", Namespace: "team-a", Name: "chaos"}, Experiment: `broken JSON`},
		{ExperimentMeta: core.ExperimentMeta{UID: "namespace", Namespace: "team-b", Name: "chaos", Archived: true}, Experiment: `broken JSON`},
		{ExperimentMeta: core.ExperimentMeta{UID: "name", Namespace: "team-a", Name: "other", Archived: true}, Experiment: `broken JSON`},
	}
	saveRecords(t, store, records...)
	found, err := store.ListArchivedManagedByNamespaceName(context.Background(), "team-a", "chaos")
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.Equal(t, "managed", found[0].UID)
	records[0].Experiment = `broken JSON`
	saveRecords(t, store, records[0])
	found, err = store.ListArchivedManagedByNamespaceName(context.Background(), "team-a", "chaos")
	require.Error(t, err)
	require.Nil(t, found)
}

func TestStoreRetentionDeleteError(t *testing.T) {
	store, db := newTestStore(t)
	expired := time.Now().Add(-48 * time.Hour)
	record := &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: "expired", Archived: true, FinishTime: &expired}}
	saveRecords(t, store, record)
	require.NoError(t, db.Exec("CREATE TRIGGER reject_delete BEFORE DELETE ON experiments BEGIN SELECT RAISE(ABORT, 'delete rejected'); END").Error)
	require.Error(t, store.DeleteByFinishTime(context.Background(), time.Hour))
	found, err := store.FindByUID(context.Background(), record.UID)
	require.NoError(t, err)
	require.Equal(t, record.ID, found.ID)
}

func TestStartupDeletesIncompleteRecords(t *testing.T) {
	store, _ := newTestStore(t)
	finished := time.Now().Add(-time.Hour)
	incomplete := &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: "incomplete"}}
	complete := &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: "complete", FinishTime: &finished}}
	saveRecords(t, store, incomplete, complete)
	DeleteIncompleteExperiments(store, nil)
	_, err := store.FindByUID(context.Background(), incomplete.UID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	found, err := store.FindByUID(context.Background(), complete.UID)
	require.NoError(t, err)
	require.Equal(t, complete.ID, found.ID)
}
