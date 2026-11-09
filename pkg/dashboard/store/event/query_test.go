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
	"errors"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/glebarez/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

var _ = Describe("Event queries", func() {
	var store core.EventStore
	var db *gorm.DB
	var statements []string
	ctx := context.Background()
	BeforeEach(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		Expect(err).NotTo(HaveOccurred())
		sqlDB, err := db.DB()
		Expect(err).NotTo(HaveOccurred())
		sqlDB.SetMaxOpenConns(1)
		DeferCleanup(sqlDB.Close)
		store, err = NewStore(db)
		Expect(err).NotTo(HaveOccurred())
		statements = nil
		Expect(db.Callback().Query().After("gorm:query").Register("capture_event_queries", func(tx *gorm.DB) {
			statements = append(statements, tx.Statement.SQL.String())
		})).To(Succeed())
	})

	It("filters before paging and counts the same repeated-field predicate", func() {
		instant := time.Date(2026, 11, 1, 0, 0, 0, 1, time.UTC)
		rows := []*core.Event{
			{ID: 1, Namespace: "a", Name: "x", Kind: "PodChaos", ObjectID: "first", CreatedAt: instant},
			{ID: 2, Namespace: "b", Name: "y", Kind: "NetworkChaos", ObjectID: "second", CreatedAt: instant},
			{ID: 3, Namespace: "b", Name: "x", Kind: "PodChaos", ObjectID: "third", CreatedAt: instant},
			{ID: 4, Namespace: "c", Name: "x", Kind: "PodChaos", ObjectID: "first", CreatedAt: instant},
			{ID: 5, Namespace: "a", Name: "z", Kind: "PodChaos", ObjectID: "first", CreatedAt: instant},
			{ID: 6, Namespace: "a", Name: "x", Kind: "StressChaos", ObjectID: "first", CreatedAt: instant},
			{ID: 7, Namespace: "a", Name: "x", Kind: "PodChaos", ObjectID: "outside", CreatedAt: instant},
			{ID: 8, Namespace: "a", Name: "x", Kind: "PodChaos", ObjectID: "first", CreatedAt: instant.Add(-time.Nanosecond)},
			{ID: 9, Namespace: "b", Name: "y", Kind: "NetworkChaos", ObjectID: "second", CreatedAt: instant.Add(time.Nanosecond)},
		}
		for _, row := range rows {
			Expect(store.Create(ctx, row)).To(Succeed())
		}
		query := core.EventQuery{
			Namespaces: []string{"a", "b"}, Names: []string{"x", "y"},
			Kinds: []string{"PodChaos", "NetworkChaos"}, ObjectIDs: []string{"first", "second", "third"},
			Start: &instant, End: &instant,
			Window: &core.EventWindow{Offset: 1, Limit: 1}, IncludeTotal: true,
		}
		result, err := store.Query(ctx, query)
		Expect(err).NotTo(HaveOccurred())
		Expect(eventIDs(result.Events)).To(Equal([]uint{2}))
		Expect(result.Total).To(HaveValue(Equal(int64(3))))
		Expect(result.HasMore).To(BeTrue())
		Expect(statements).To(HaveLen(2))
		Expect(strings.ToLower(statements[0])).To(ContainSubstring("count(*)"))
		Expect(statements[1]).To(ContainSubstring("LIMIT 2 OFFSET 1"))
		statements = nil
		query.IncludeTotal = false
		query.Window.Offset = 2
		result, err = store.Query(ctx, query)
		Expect(err).NotTo(HaveOccurred())
		Expect(eventIDs(result.Events)).To(Equal([]uint{1}))
		Expect(result.Total).To(BeNil())
		Expect(result.HasMore).To(BeFalse())
		Expect(statements).To(HaveLen(1))
		query.Window.Offset = 99
		query.IncludeTotal = true
		result, err = store.Query(ctx, query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Events).To(BeEmpty())
		Expect(result.Events).NotTo(BeNil())
		Expect(result.HasMore).To(BeFalse())
		Expect(result.Total).To(HaveValue(Equal(int64(3))))
		query.Namespaces = []string{"missing"}
		result, err = store.Query(ctx, query)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Total).To(HaveValue(Equal(int64(0))))
	})

	It("sorts all supported fields globally and breaks ties by ID in the same direction", func() {
		early := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		for _, row := range []*core.Event{
			{ID: 1, Namespace: "z", Name: "z", Kind: "z", ObjectID: "z", Message: "z", CreatedAt: early.Add(time.Second)},
			{ID: 2, Namespace: "a", Name: "a", Kind: "a", ObjectID: "a", Message: "a", CreatedAt: early},
			{ID: 3, Namespace: "a", Name: "a", Kind: "a", ObjectID: "a", Message: "a", CreatedAt: early},
		} {
			Expect(store.Create(ctx, row)).To(Succeed())
		}
		for _, sort := range []core.EventSort{core.EventSortID, core.EventSortObjectID, core.EventSortNamespace,
			core.EventSortName, core.EventSortKind, core.EventSortCreatedAt, core.EventSortMessage} {
			for _, order := range []core.EventOrder{core.EventOrderAsc, core.EventOrderDesc} {
				query := core.EventQuery{Sort: sort, Order: order, Window: &core.EventWindow{Offset: 1, Limit: 1}}
				result, err := store.Query(ctx, query)
				Expect(err).NotTo(HaveOccurred())
				expected := uint(3)
				if sort == core.EventSortID {
					expected = 2
				}
				Expect(eventIDs(result.Events)).To(Equal([]uint{expected}), "%s %s", sort, order)
				Expect(result.HasMore).To(BeTrue())
			}
		}
		result, err := store.Query(ctx, core.EventQuery{Window: &core.EventWindow{Limit: 0}})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Events).To(BeEmpty())
	})

	It("sorts mixed-offset history as exact instants and distinguishes NULL from year one", func() {
		values := []any{
			nil, "0001-01-01T00:00:00Z", "1960-01-01T00:00:00.000000001Z",
			"2026-11-01T08:00:00.000000001+08:00", "2026-11-01T00:00:00.000000001Z",
			"2026-11-01 00:00:00.000000002 +0000 UTC m=+1", "2400-01-01T00:00:00Z",
		}
		for id, value := range values {
			Expect(db.Exec("INSERT INTO events (id, created_at) VALUES (?, ?)", id, value).Error).To(Succeed())
		}
		query := core.EventQuery{Sort: core.EventSortCreatedAt, Order: core.EventOrderAsc, Window: &core.EventWindow{Limit: 3}}
		result, err := store.Query(ctx, query)
		Expect(err).NotTo(HaveOccurred())
		Expect(eventIDs(result.Events)).To(Equal([]uint{0, 1, 2}))
		query.Window.Offset = 3
		result, err = store.Query(ctx, query)
		Expect(err).NotTo(HaveOccurred())
		Expect(eventIDs(result.Events)).To(Equal([]uint{3, 4, 5}))
		Expect(result.HasMore).To(BeTrue())
		yearOne := time.Time{}
		query.Start, query.End = &yearOne, &yearOne
		query.Window.Offset = 0
		query.IncludeTotal = true
		result, err = store.Query(ctx, query)
		Expect(err).NotTo(HaveOccurred())
		Expect(eventIDs(result.Events)).To(Equal([]uint{1}))
		Expect(result.Total).To(HaveValue(Equal(int64(1))))
	})

	It("rejects malformed timestamps in scope but does not evaluate excluded metadata", func() {
		for _, malformed := range []any{"broken", "2026-11-01T00:00:00Z\x00junk", "", []byte{}, int64(42)} {
			Expect(db.Exec("DELETE FROM events").Error).To(Succeed())
			Expect(db.Exec("INSERT INTO events (id, namespace, name, kind, object_id, created_at) VALUES (1, 'keep', 'keep', 'keep', 'keep', '2026-11-01T00:00:00Z')").Error).To(Succeed())
			Expect(db.Exec("INSERT INTO events (id, namespace, name, kind, object_id, created_at) VALUES (2, 'omit', 'omit', 'omit', 'omit', ?)", malformed).Error).To(Succeed())
			if raw, ok := malformed.([]byte); ok && len(raw) == 0 {
				Expect(db.Exec("UPDATE events SET created_at = X'' WHERE id = 2").Error).To(Succeed())
			}
			bound := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
			for _, query := range []core.EventQuery{
				{Namespaces: []string{"keep"}}, {Names: []string{"keep"}},
				{Kinds: []string{"keep"}}, {ObjectIDs: []string{"keep"}},
			} {
				query.Sort, query.Start, query.IncludeTotal = core.EventSortCreatedAt, &bound, true
				query.Window = &core.EventWindow{Limit: 1}
				result, err := store.Query(ctx, query)
				Expect(err).NotTo(HaveOccurred(), "malformed value: %#v", malformed)
				Expect(eventIDs(result.Events)).To(Equal([]uint{1}))
				Expect(result.Total).To(HaveValue(Equal(int64(1))))
			}
			result, err := store.Query(ctx, core.EventQuery{Sort: core.EventSortCreatedAt, Start: &bound, Window: &core.EventWindow{Limit: 1}, IncludeTotal: true})
			Expect(err).To(HaveOccurred(), "malformed value: %#v", malformed)
			Expect(result.Events).To(BeNil())
		}
	})

	It("propagates cancellation without a partial page", func() {
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		for _, query := range []core.EventQuery{
			{Window: &core.EventWindow{Limit: 12}, IncludeTotal: true},
			{Window: &core.EventWindow{Limit: 0}},
		} {
			result, err := store.Query(canceled, query)
			Expect(errors.Is(err, context.Canceled)).To(BeTrue())
			Expect(result.Events).To(BeNil())
		}
	})
})

func TestEventQueryCountSnapshot(t *testing.T) {
	g := NewWithT(t)
	store, db, writer := newFileEventTestStore(t)
	g.Expect(db.Exec("INSERT INTO events (id, created_at) VALUES (1, '2026-11-01T00:00:00Z')").Error).To(Succeed())
	inserted := false
	g.Expect(db.Callback().Query().After("gorm:query").Register("insert_after_event_count", func(tx *gorm.DB) {
		if !inserted && strings.Contains(tx.Statement.SQL.String(), "count(*)") {
			inserted = true
			g.Expect(writer.Exec("INSERT INTO events (id, created_at) VALUES (2, '2026-11-01T00:00:00Z')").Error).To(Succeed())
		}
	})).To(Succeed())
	result, err := store.Query(context.Background(), core.EventQuery{Window: &core.EventWindow{Limit: 12}, IncludeTotal: true})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(inserted).To(BeTrue())
	g.Expect(result.Total).To(HaveValue(Equal(int64(1))))
	g.Expect(eventIDs(result.Events)).To(Equal([]uint{1}))
	var currentCount int64
	g.Expect(writer.Model(&core.Event{}).Count(&currentCount).Error).To(Succeed())
	g.Expect(currentCount).To(Equal(int64(2)))
}

func TestEventSQLServerMessageSort(t *testing.T) {
	for _, order := range []core.EventOrder{core.EventOrderAsc, core.EventOrderDesc} {
		t.Run(string(order), func(t *testing.T) {
			g := NewWithT(t)
			sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			g.Expect(err).NotTo(HaveOccurred())
			defer sqlDB.Close()
			db, err := gorm.Open(sqlserver.New(sqlserver.Config{Conn: sqlDB}), &gorm.Config{DisableAutomaticPing: true})
			g.Expect(err).NotTo(HaveOccurred())
			direction := strings.ToUpper(string(order))
			idDirection := ""
			if order == core.EventOrderDesc {
				idDirection = " DESC"
			}
			mock.ExpectQuery(`SELECT * FROM "events" ORDER BY CAST(message AS varchar(max)) ` + direction + `,"id"` + idDirection + ` OFFSET 0 ROW FETCH NEXT 2 ROWS ONLY`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "message"}).AddRow(1, "message"))
			result, err := (&eventStore{db: db}).Query(context.Background(), core.EventQuery{Sort: core.EventSortMessage, Order: order, Window: &core.EventWindow{Limit: 1}})
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(eventIDs(result.Events)).To(Equal([]uint{1}))
			g.Expect(mock.ExpectationsWereMet()).To(Succeed())
		})
	}
}
