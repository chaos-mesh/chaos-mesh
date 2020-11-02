// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/store/dbstore"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEvent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Event Suite")
}

var _ = Describe("event", func() {
	var es *eventStore
	var mock sqlmock.Sqlmock

	BeforeEach(func() {
		var db *sql.DB
		var err error
		db, mock, err = sqlmock.New()
		Expect(err).ShouldNot(HaveOccurred())

		gdb, err := gorm.Open("sqlite3", db)
		Expect(err).ShouldNot(HaveOccurred())

		es = &eventStore{db: &dbstore.DB{DB: gdb}}
	})

	AfterEach(func() {
		err := mock.ExpectationsWereMet()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("findPodRecordsByEventID", func() {
		It("found", func() {
			timeNow := time.Now()
			podRecord0 := &core.PodRecord{
				ID:        0,
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
				DeletedAt: &timeNow,
				EventID:   0,
				PodIP:     "testIP",
				PodName:   "testName",
				Namespace: "testNamespace",
				Message:   "testMessage",
				Action:    "testAction",
			}
			podRecord1 := &core.PodRecord{
				ID:        1,
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
				DeletedAt: &timeNow,
				EventID:   1,
				PodIP:     "testIP",
				PodName:   "testName",
				Namespace: "testNamespace",
				Message:   "testMessage",
				Action:    "testAction",
			}

			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action).
				AddRow(podRecord1.ID, podRecord1.CreatedAt, podRecord1.UpdatedAt, podRecord1.DeletedAt, podRecord1.EventID,
					podRecord1.PodIP, podRecord1.PodName, podRecord1.Namespace, podRecord1.Message, podRecord1.Action)

			sqlSelectAll := `SELECT * FROM "pod_records"`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelectAll)).WillReturnRows(rows)

			podRecords, err := es.findPodRecordsByEventID(context.TODO(), podRecord0.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(podRecords[0]).Should(Equal(podRecord0))
		})

		It("not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			podRecords, err := es.findPodRecordsByEventID(context.TODO(), 1)
			Expect(len(podRecords)).Should(Equal(0))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("min", func() {
		It("x", func() {
			x := 1
			y := 2
			res := min(x, y)
			Expect(res).Should(Equal(x))
		})

		It("y", func() {
			x := 2
			y := 1
			res := min(x, y)
			Expect(res).Should(Equal(y))
		})
	})

	Context("list", func() {
		It("found", func() {
			timeNow := time.Now()
			podRecord := &core.PodRecord{
				ID:        0,
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
				DeletedAt: &timeNow,
				EventID:   0,
				PodIP:     "testIP",
				PodName:   "testName",
				Namespace: "testNamespace",
				Message:   "testMessage",
				Action:    "testAction",
			}
			event := &core.Event{
				ID:           0,
				CreatedAt:    timeNow,
				UpdatedAt:    timeNow,
				DeletedAt:    &timeNow,
				Experiment:   "testExperiment",
				Namespace:    "testNamespace",
				Kind:         "testKind",
				Message:      "testMessage",
				StartTime:    &timeNow,
				FinishTime:   &timeNow,
				Duration:     "10s",
				Pods:         []*core.PodRecord{podRecord},
				ExperimentID: "testID",
			}
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event.ID, event.CreatedAt, event.UpdatedAt, event.DeletedAt, event.Experiment, event.Namespace,
					event.Kind, event.Message, event.StartTime, event.FinishTime, event.Duration, event.ExperimentID)
			sqlSelectAll := `SELECT * FROM "events"`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelectAll)).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord.ID, podRecord.CreatedAt, podRecord.UpdatedAt, podRecord.DeletedAt, podRecord.EventID,
					podRecord.PodIP, podRecord.PodName, podRecord.Namespace, podRecord.Message, podRecord.Action)

			sqlSelectAll = `SELECT * FROM "pod_records"`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelectAll)).WillReturnRows(rows)

			events, err := es.List(context.TODO())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event))
		})

		It("not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			events, err := es.List(context.TODO())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})
	})

})
