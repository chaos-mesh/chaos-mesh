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
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func storedEventTimes(t *testing.T, db *gorm.DB) map[uint]string {
	t.Helper()
	var rows []struct {
		ID    uint
		Value string
	}
	require.NoError(t, db.Raw("SELECT id, CAST(created_at AS TEXT) AS value FROM events ORDER BY id").Scan(&rows).Error)
	result := make(map[uint]string, len(rows))
	for _, row := range rows {
		result[row.ID] = row.Value
	}
	return result
}

func TestEventMixedTimezoneFiltering(t *testing.T) {
	store, db := newEventTestStore(t)
	for i, at := range []string{
		"2026-10-16 08:00:00+08:00",
		"2026-10-15 19:00:00-05:00",
		"2026-10-16T00:00:00Z",
		"2026-10-16 07:59:59.999999999+08:00",
		"2026-10-15 19:00:00.000000001-05:00",
		"2026-10-16 00:00:01+00:00",
		"2026-10-16 00:00:00",
		"2026-10-16 05:45:00+05:45",
		"2026-10-16T10:00:00Z",
		"2026-10-16T00:00:00Z",
		"2026-10-16 08:00:00 +0800 CST",
		"2026-10-15 19:00:00 -0500 EST m=+12.345",
	} {
		objectID := "target"
		if i == 9 {
			objectID = "other"
		}
		require.NoError(t, db.Exec("INSERT INTO events (id, object_id, namespace, name, kind, created_at) VALUES (?, ?, ?, ?, ?, ?)", i+1, objectID, "team-a", "chaos", "PodChaos", at).Error)
	}
	for i, row := range []struct{ namespace, name, kind string }{
		{"team-b", "chaos", "PodChaos"}, {"team-a", "other", "PodChaos"}, {"team-a", "chaos", "NetworkChaos"},
	} {
		require.NoError(t, db.Exec("INSERT INTO events (id, object_id, namespace, name, kind, created_at) VALUES (?, ?, ?, ?, ?, ?)", i+13, "target", row.namespace, row.name, row.kind, "2026-10-16 08:00:00+08:00").Error)
	}
	before := storedEventTimes(t, db)
	_, err := NewStore(db)
	require.NoError(t, err)
	require.Equal(t, before, storedEventTimes(t, db), "opening the store must not rewrite history")
	for _, scenario := range []struct {
		name, start, end, limit string
		want                    []uint
	}{
		{"same instant", "2026-10-16 00:00:00", "2026-10-16 00:00:00", "", []uint{12, 11, 8, 7, 3, 2, 1}},
		{"inclusive start", "2026-10-16 00:00:00", "", "", []uint{12, 11, 9, 8, 7, 6, 5, 3, 2, 1}},
		{"inclusive end", "", "2026-10-16 00:00:00", "", []uint{12, 11, 8, 7, 4, 3, 2, 1}},
		{"limit after time filter", "2026-10-16 00:00:00", "2026-10-16 00:00:00", "2", []uint{12, 11}},
		{"zero limit", "2026-10-16 00:00:00", "2026-10-16 00:00:00", "0", []uint{}},
		{"unlimited", "2026-10-16 00:00:00", "2026-10-16 00:00:00", "-1", []uint{12, 11, 8, 7, 3, 2, 1}},
		{"one nanosecond later", "2026-10-16 00:00:00.000000001", "2026-10-16 00:00:00.000000001", "", []uint{5}},
		{"reversed range", "2026-10-16 00:00:01", "2026-10-16 00:00:00", "", []uint{}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			filter := core.Filter{ObjectID: "target", Namespace: "team-a", Name: "chaos", Kind: "PodChaos", Start: scenario.start, End: scenario.end, Limit: scenario.limit}
			got, err := store.ListByFilter(context.Background(), filter)
			require.NoError(t, err)
			require.Equal(t, scenario.want, eventIDs(got))
			got, err = store.ListByUIDListWithFilter(context.Background(), []string{"target"}, filter)
			require.NoError(t, err)
			require.Equal(t, scenario.want, eventIDs(got))
			got, err = store.ListByUIDListWithFilter(context.Background(), []string{"other"}, filter)
			require.NoError(t, err)
			require.Empty(t, got)
		})
	}
	got, err := store.ListByUIDListWithFilter(context.Background(), []string{"target"}, core.Filter{Namespace: "team-a", Name: "chaos", Kind: "PodChaos", Start: "2026-10-16 00:00:00", End: "2026-10-16 00:00:00"})
	require.NoError(t, err)
	require.Equal(t, []uint{12, 11, 8, 7, 3, 2, 1}, eventIDs(got), "UID and metadata restrictions must all survive time filtering")
	require.Equal(t, before, storedEventTimes(t, db), "filtering must not rewrite timestamps")
}

func TestEventMixedTimezoneRetention(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store, db := newEventTestStore(t)
		cutoff := time.Now().UTC().Add(-time.Hour)
		for i, row := range []struct {
			delta  time.Duration
			zone   int
			layout string
		}{
			{0, 0, "2006-01-02 15:04:05.999999999-07:00"},
			{0, 8 * 3600, "2006-01-02 15:04:05.999999999-07:00"},
			{0, -5 * 3600, "2006-01-02 15:04:05.999999999-07:00"},
			{-time.Nanosecond, 8 * 3600, "2006-01-02 15:04:05.999999999-07:00"},
			{time.Nanosecond, -5 * 3600, "2006-01-02 15:04:05.999999999-07:00"},
			{time.Second, 8 * 3600, "2006-01-02 15:04:05.999999999-07:00"},
			{-time.Nanosecond, 0, time.RFC3339Nano},
			{time.Nanosecond, 0, time.RFC3339Nano},
			{0, 0, "2006-01-02 15:04:05"},
		} {
			at := cutoff.Add(row.delta).In(time.FixedZone("stored", row.zone)).Format(row.layout)
			require.NoError(t, db.Exec("INSERT INTO events (id, object_id, created_at) VALUES (?, ?, ?)", i+1, "target", at).Error)
		}
		before := storedEventTimes(t, db)
		require.NoError(t, store.DeleteByDuration(context.Background(), time.Hour))
		got, err := store.List(context.Background())
		require.NoError(t, err)
		require.ElementsMatch(t, []uint{5, 6, 8}, eventIDs(got), "only events newer than the inclusive cutoff survive")
		require.Equal(t, map[uint]string{5: before[5], 6: before[6], 8: before[8]}, storedEventTimes(t, db), "surviving timestamp bytes must remain unchanged")
		require.NoError(t, store.DeleteByDuration(context.Background(), time.Hour))
		got, err = store.List(context.Background())
		require.NoError(t, err)
		require.ElementsMatch(t, []uint{5, 6, 8}, eventIDs(got))
	})
}

func TestEventTimeFilteringLargeHistory(t *testing.T) {
	store, db := newEventTestStore(t)
	base := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	for id := 1; id <= 1200; id++ {
		at := base.Add(time.Hour)
		if id == 3 || id == 260 || id == 530 {
			at = base
		}
		at = at.In(time.FixedZone("stored", 8*3600))
		require.NoError(t, db.Exec("INSERT INTO events (id, object_id, created_at) VALUES (?, ?, ?)", id, "target", at.Format(time.RFC3339Nano)).Error)
	}
	filter := core.Filter{Start: "2026-10-16 00:00:00", End: "2026-10-16 00:00:00", Limit: "2"}
	got, err := store.ListByFilter(context.Background(), filter)
	require.NoError(t, err)
	require.Equal(t, []uint{530, 260}, eventIDs(got), "newer nonmatching history must not consume the limit")
	filter.Limit = "-1"
	got, err = store.ListByUIDListWithFilter(context.Background(), []string{"target"}, filter)
	require.NoError(t, err)
	require.Equal(t, []uint{530, 260, 3}, eventIDs(got))
}

func TestEventTimeRetentionFailureAndRetry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store, db := newEventTestStore(t)
		cutoff := time.Now().UTC().Add(-time.Hour)
		for id := 1; id <= 1200; id++ {
			at := cutoff.Add(-time.Nanosecond)
			if id == 2 {
				at = cutoff.Add(time.Nanosecond)
			}
			at = at.In(time.FixedZone("stored", -5*3600))
			require.NoError(t, db.Exec("INSERT INTO events (id, object_id, created_at) VALUES (?, ?, ?)", id, "target", at.Format(time.RFC3339Nano)).Error)
		}
		require.NoError(t, db.Exec("INSERT INTO events (id, object_id, created_at) VALUES (1201, 'unknown-time', NULL)").Error)
		before := storedEventTimes(t, db)
		require.NoError(t, db.Exec("CREATE TRIGGER reject_expiry BEFORE DELETE ON events WHEN OLD.id = 1 BEGIN SELECT RAISE(ABORT, 'expiry rejected'); END").Error)
		require.ErrorContains(t, store.DeleteByDuration(context.Background(), time.Hour), "expiry rejected")
		require.Equal(t, before, storedEventTimes(t, db), "failed cleanup must roll back earlier deletions")
		require.NoError(t, db.Exec("DROP TRIGGER reject_expiry").Error)
		require.NoError(t, store.DeleteByDuration(context.Background(), time.Hour))
		got, err := store.List(context.Background())
		require.NoError(t, err)
		require.ElementsMatch(t, []uint{2, 1201}, eventIDs(got), "unexpired and unknown times must survive")
	})
}

func TestEventTimeComparisonsIncludeZeroID(t *testing.T) {
	store, db := newEventTestStore(t)
	at := time.Now().UTC().Add(-time.Hour)
	for id := range 1024 {
		require.NoError(t, db.Exec("INSERT INTO events (id, object_id, created_at) VALUES (?, ?, ?)", id, "target", at.In(time.FixedZone("stored", 8*3600)).Format(time.RFC3339Nano)).Error)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	bound := at.Format("2006-01-02 15:04:05.999999999")
	got, err := store.ListByFilter(ctx, core.Filter{Start: bound, End: bound})
	require.NoError(t, err)
	require.Len(t, got, 1024)
	require.EqualValues(t, 1023, got[0].ID)
	require.Zero(t, got[len(got)-1].ID)
	require.NoError(t, store.DeleteByDuration(ctx, 0))
	got, err = store.List(context.Background())
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestEventTimeCleanupRejectsMalformedHistory(t *testing.T) {
	store, db := newEventTestStore(t)
	at := time.Now().UTC().Add(-time.Hour)
	for id := 1; id <= 600; id++ {
		value := at.In(time.FixedZone("stored", 8*3600)).Format(time.RFC3339Nano)
		if id == 1 {
			value = "0000-not-a-date"
		}
		require.NoError(t, db.Exec("INSERT INTO events (id, object_id, created_at) VALUES (?, ?, ?)", id, "target", value).Error)
	}
	before := storedEventTimes(t, db)
	require.Error(t, store.DeleteByDuration(context.Background(), 0))
	require.Equal(t, before, storedEventTimes(t, db), "unreadable timestamps must not cause partial deletion")
	got, err := store.ListByFilter(context.Background(), core.Filter{End: at.Format("2006-01-02 15:04:05.999999999")})
	require.Error(t, err)
	require.Nil(t, got, "a failed filtered read must not return partial results")
}
