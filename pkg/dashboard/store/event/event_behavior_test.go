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

package event

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func newEventTestStore(t *testing.T) (core.EventStore, *gorm.DB) {
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

func eventIDs(events []*core.Event) []uint {
	ids := make([]uint, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}
	return ids
}

func TestEventStoreFilters(t *testing.T) {
	store, _ := newEventTestStore(t)
	ctx := context.Background()
	base := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	events := []*core.Event{
		{ID: 1, ObjectID: "target", Namespace: "team-a", Name: "chaos", Kind: "PodChaos", CreatedAt: base.Add(-time.Hour)},
		{ID: 2, ObjectID: "target", Namespace: "team-a", Name: "chaos", Kind: "PodChaos", CreatedAt: base},
		{ID: 3, ObjectID: "target", Namespace: "team-a", Name: "chaos", Kind: "PodChaos", CreatedAt: base.Add(time.Hour)},
		{ID: 4, ObjectID: "other", Namespace: "team-a", Name: "chaos", Kind: "PodChaos", CreatedAt: base},
		{ID: 5, ObjectID: "target", Namespace: "team-b", Name: "chaos", Kind: "PodChaos", CreatedAt: base},
		{ID: 6, ObjectID: "target", Namespace: "team-a", Name: "other", Kind: "PodChaos", CreatedAt: base},
		{ID: 7, ObjectID: "target", Namespace: "team-a", Name: "chaos", Kind: "NetworkChaos", CreatedAt: base},
	}
	for _, event := range events {
		require.NoError(t, store.Create(ctx, event))
	}
	format := func(value time.Time) string { return value.Format("2006-01-02 15:04:05") }
	cases := []struct {
		name   string
		filter core.Filter
		want   []uint
	}{
		{"empty", core.Filter{}, []uint{7, 6, 5, 4, 3, 2, 1}},
		{"object", core.Filter{ObjectID: "target"}, []uint{7, 6, 5, 3, 2, 1}},
		{"namespace", core.Filter{Namespace: "team-a"}, []uint{7, 6, 4, 3, 2, 1}},
		{"name", core.Filter{Name: "chaos"}, []uint{7, 5, 4, 3, 2, 1}},
		{"kind", core.Filter{Kind: "PodChaos"}, []uint{6, 5, 4, 3, 2, 1}},
		{"combined", core.Filter{ObjectID: "target", Namespace: "team-a", Name: "chaos", Kind: "PodChaos"}, []uint{3, 2, 1}},
		{"limit after filter", core.Filter{ObjectID: "target", Namespace: "team-a", Name: "chaos", Kind: "PodChaos", Limit: "2"}, []uint{3, 2}},
		{"limit only", core.Filter{Limit: "2"}, []uint{7, 6}},
		{"zero limit", core.Filter{Limit: "0"}, []uint{}},
		{"no matches", core.Filter{ObjectID: "missing"}, []uint{}},
		{"quoted value", core.Filter{Name: "chaos' OR 1=1 --"}, []uint{}},
		{"start", core.Filter{Start: format(base.Add(-time.Minute))}, []uint{7, 6, 5, 4, 3, 2}},
		{"end", core.Filter{End: format(base.Add(time.Minute))}, []uint{7, 6, 5, 4, 2, 1}},
		{"inclusive start", core.Filter{Start: format(base)}, []uint{7, 6, 5, 4, 3, 2}},
		{"inclusive end", core.Filter{End: format(base)}, []uint{7, 6, 5, 4, 2, 1}},
		{"exact time", core.Filter{Start: format(base), End: format(base)}, []uint{7, 6, 5, 4, 2}},
		{"time range", core.Filter{Start: format(base.Add(-time.Minute)), End: format(base.Add(time.Minute))}, []uint{7, 6, 5, 4, 2}},
		{"reversed range", core.Filter{Start: format(base.Add(time.Minute)), End: format(base.Add(-time.Minute))}, []uint{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := store.ListByFilter(ctx, tc.filter)
			require.NoError(t, err)
			require.Equal(t, tc.want, eventIDs(actual))
		})
	}
	actual, err := store.ListByUIDListWithFilter(ctx, []string{"target"}, core.Filter{Namespace: "team-a", Name: "chaos", Kind: "PodChaos", Limit: "2"})
	require.NoError(t, err)
	require.Equal(t, []uint{3, 2}, eventIDs(actual))
	actual, err = store.ListByUIDListWithFilter(ctx, nil, core.Filter{})
	require.NoError(t, err)
	require.Empty(t, actual)
	actual, err = store.ListByUIDList(ctx, nil)
	require.NoError(t, err)
	require.Empty(t, actual)
	actual, err = store.ListByUID(ctx, "other")
	require.NoError(t, err)
	require.Equal(t, []uint{4}, eventIDs(actual))
	actual, err = store.ListByExperiment(ctx, "team-a", "chaos", "PodChaos")
	require.NoError(t, err)
	require.ElementsMatch(t, []uint{1, 2, 3, 4}, eventIDs(actual))
	actual, err = store.ListByUIDList(ctx, []string{"target", "other"})
	require.NoError(t, err)
	require.Len(t, actual, len(events))
	for _, limit := range []string{"invalid", "1.5", "9999999999999999999999999999999"} {
		actual, err := store.ListByFilter(ctx, core.Filter{Limit: limit})
		require.Error(t, err)
		require.Nil(t, actual)
		actual, err = store.ListByUIDListWithFilter(ctx, []string{"target"}, core.Filter{Limit: limit})
		require.Error(t, err)
		require.Nil(t, actual)
	}
}

func TestEventStoreCreateAndFind(t *testing.T) {
	store, _ := newEventTestStore(t)
	ctx := context.Background()
	event := &core.Event{ObjectID: "target", Namespace: "team-a", Name: "chaos", Kind: "PodChaos", Type: "Normal", Reason: "Applied", Message: "fault injected"}
	require.NoError(t, store.Create(ctx, event))
	require.NotZero(t, event.ID)
	require.False(t, event.CreatedAt.IsZero())
	found, err := store.Find(ctx, event.ID)
	require.NoError(t, err)
	require.Equal(t, event.ObjectID, found.ObjectID)
	require.Equal(t, event.Namespace, found.Namespace)
	require.Equal(t, event.Name, found.Name)
	require.Equal(t, event.Kind, found.Kind)
	require.Equal(t, event.Type, found.Type)
	require.Equal(t, event.Reason, found.Reason)
	require.Equal(t, event.Message, found.Message)
	require.True(t, event.CreatedAt.Equal(found.CreatedAt))
	duplicate := *event
	duplicate.Message = "replacement"
	require.Error(t, store.Create(ctx, &duplicate))
	found, err = store.Find(ctx, event.ID)
	require.NoError(t, err)
	require.Equal(t, "fault injected", found.Message)
	found, err = store.Find(ctx, event.ID+1)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, found)
}

func TestEventStoreDeletion(t *testing.T) {
	for _, operation := range []string{"uid", "uid list", "empty uid list", "duration"} {
		t.Run(operation, func(t *testing.T) {
			store, _ := newEventTestStore(t)
			ctx := context.Background()
			now := time.Now().UTC()
			events := []*core.Event{
				{ID: 1, ObjectID: "target", CreatedAt: now.Add(-48 * time.Hour)},
				{ID: 2, ObjectID: "target", CreatedAt: now},
				{ID: 3, ObjectID: "other", CreatedAt: now.Add(-time.Hour)},
				{ID: 4, ObjectID: "keep", CreatedAt: now.Add(time.Hour)},
			}
			for _, event := range events {
				require.NoError(t, store.Create(ctx, event))
			}
			var want []uint
			switch operation {
			case "uid":
				require.NoError(t, store.DeleteByUID(ctx, "target"))
				require.NoError(t, store.DeleteByUID(ctx, "target"))
				want = []uint{3, 4}
			case "uid list":
				require.NoError(t, store.DeleteByUIDList(ctx, []string{"target", "other", "missing"}))
				want = []uint{4}
			case "empty uid list":
				require.NoError(t, store.DeleteByUIDList(ctx, nil))
				require.NoError(t, store.DeleteByUIDList(ctx, []string{}))
				want = []uint{1, 2, 3, 4}
			case "duration":
				require.NoError(t, store.DeleteByDuration(ctx, 24*time.Hour))
				want = []uint{2, 3, 4}
			}
			remaining, err := store.List(ctx)
			require.NoError(t, err)
			require.ElementsMatch(t, want, eventIDs(remaining))
		})
	}
}

func TestEventStoreDatabaseErrors(t *testing.T) {
	store, db := newEventTestStore(t)
	require.NoError(t, db.Migrator().DropTable(&core.Event{}))
	ctx := context.Background()
	events, err := store.List(ctx)
	require.Error(t, err)
	require.Nil(t, events)
	_, err = store.ListByUID(ctx, "target")
	require.Error(t, err)
	_, err = store.ListByUIDList(ctx, []string{"target"})
	require.Error(t, err)
	_, err = store.ListByFilter(ctx, core.Filter{})
	require.Error(t, err)
	_, err = store.ListByUIDListWithFilter(ctx, []string{"target"}, core.Filter{})
	require.Error(t, err)
	_, err = store.ListByExperiment(ctx, "team-a", "chaos", "PodChaos")
	require.Error(t, err)
	_, err = store.Find(ctx, 1)
	require.Error(t, err)
	require.Error(t, store.Create(ctx, &core.Event{}))
	require.Error(t, store.DeleteByUID(ctx, "target"))
	require.Error(t, store.DeleteByUIDList(ctx, []string{"target"}))
	require.Error(t, store.DeleteByDuration(ctx, time.Hour))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	store, err = NewStore(db)
	require.Error(t, err)
	require.Nil(t, store)
}
