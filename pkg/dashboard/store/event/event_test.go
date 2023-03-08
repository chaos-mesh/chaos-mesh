// Copyright 2021 Chaos Mesh Authors.
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
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func TestEvent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Event Suite")
}

func genRows() *sqlmock.Rows {
	return sqlmock.NewRows(
		[]string{"id", "object_id", "created_at", "namespace", "name", "kind", "type", "reason", "message"},
	)
}

func addRow(rows *sqlmock.Rows, event *core.Event) {
	rows.AddRow(event.ID, event.ObjectID, event.CreatedAt, event.Namespace, event.Name,
		event.Kind, event.Type, event.Reason, event.Message)
}

var _ = Describe("Event", func() {
	var (
		err    error
		db     *sql.DB
		mock   sqlmock.Sqlmock
		es     *eventStore
		event0 *core.Event
		event1 *core.Event
	)

	BeforeEach(func() {
		db, mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		Expect(err).ShouldNot(HaveOccurred())

		gdb, err := gorm.Open("sqlite3", db)
		Expect(err).ShouldNot(HaveOccurred())

		es = &eventStore{db: gdb}

		now := time.Now()
		event0 = &core.Event{
			ID:        0,
			ObjectID:  "UID0",
			CreatedAt: now,
			Namespace: "default",
			Name:      "event0",
			Kind:      "PodChaos",
			Type:      "type",
			Reason:    "reason",
			Message:   "message",
		}
		event1 = &core.Event{
			ID:        1,
			ObjectID:  "UID1",
			CreatedAt: now.Add(time.Hour * 24),
			Namespace: "chaos-mesh",
			Name:      "event1",
			Kind:      "NetworkChaos",
			Type:      "type",
			Reason:    "reason",
			Message:   "message",
		}
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	Context("List", func() {
		It("event0 should be found", func() {
			rows := genRows()
			addRow(rows, event0)

			mock.ExpectQuery("SELECT * FROM \"events\"").WillReturnRows(rows)

			events, err := es.List(context.TODO())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})
	})

	Context("ListByUID", func() {
		sql := "SELECT * FROM \"events\" WHERE (object_id = ?)"

		It("event0 should be found", func() {
			rows := genRows()
			addRow(rows, event0)

			mock.ExpectQuery(sql).WithArgs(event0.ObjectID).WillReturnRows(rows)

			events, err := es.ListByUID(context.TODO(), "UID0")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("event0 shoud not be found", func() {
			rows := genRows()
			addRow(rows, event0)
			mock.ExpectQuery(sql).WithArgs(event1.ObjectID).WillReturnRows(sqlmock.NewRows(nil))

			events, err := es.ListByUID(context.TODO(), "UID1")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})
	})

	Context("ListByExperiment", func() {
		sql := "SELECT * FROM \"events\" WHERE (namespace = ? AND name = ? AND kind = ?)"

		It("event0 should be found", func() {
			rows := genRows()
			addRow(rows, event0)

			mock.ExpectQuery(sql).WithArgs(event0.Namespace, event0.Name, event0.Kind).WillReturnRows(rows)

			events, err := es.ListByExperiment(context.TODO(), "default", "event0", "PodChaos")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("event1 should be found", func() {
			rows := genRows()
			addRow(rows, event1)

			mock.ExpectQuery(sql).WithArgs(event1.Namespace, event1.Name, event1.Kind).WillReturnRows(rows)

			events, err := es.ListByExperiment(context.TODO(), "chaos-mesh", "event1", "NetworkChaos")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event1))
		})
	})

	Context("Find", func() {
		sql := "SELECT * FROM \"events\" WHERE (\"events\".\"id\" = 0) ORDER BY \"events\".\"id\" ASC LIMIT 1"

		It("event0 should be found", func() {
			rows := genRows()
			addRow(rows, event0)

			mock.ExpectQuery(sql).WillReturnRows(rows)

			event, err := es.Find(context.TODO(), 0)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(event).Should(Equal(event0))
		})
	})
})
