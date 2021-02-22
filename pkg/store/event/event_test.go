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
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/store/dbstore"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEvent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Event Suite")
}

var _ = Describe("event", func() {
	var (
		es         *eventStore
		mock       sqlmock.Sqlmock
		podRecord0 *core.PodRecord
		podRecord1 *core.PodRecord
		event0     *core.Event
		event1     *core.Event
		timeNow    time.Time
	)

	BeforeEach(func() {
		var db *sql.DB
		var err error
		db, mock, err = sqlmock.New()
		Expect(err).ShouldNot(HaveOccurred())

		gdb, err := gorm.Open("sqlite3", db)
		Expect(err).ShouldNot(HaveOccurred())

		es = &eventStore{db: &dbstore.DB{DB: gdb}}

		timeNow = time.Now()
		oneMinute, _ := time.ParseDuration("1m")
		timeAfter := timeNow.Add(oneMinute)
		podRecord0 = &core.PodRecord{
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
		podRecord1 = &core.PodRecord{
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
		event0 = &core.Event{
			ID:           0,
			CreatedAt:    timeNow,
			UpdatedAt:    timeNow,
			DeletedAt:    &timeNow,
			Experiment:   "testExperiment0",
			Namespace:    "testNamespace",
			Kind:         "testKind",
			Message:      "testMessage",
			StartTime:    &timeNow,
			FinishTime:   &timeNow,
			Duration:     "10s",
			Pods:         []*core.PodRecord{podRecord0},
			ExperimentID: "testID0",
		}
		event1 = &core.Event{
			ID:           1,
			CreatedAt:    timeAfter,
			UpdatedAt:    timeAfter,
			DeletedAt:    &timeAfter,
			Experiment:   "testExperiment0",
			Namespace:    "testNamespace",
			Kind:         "testKind",
			Message:      "testMessage",
			StartTime:    &timeAfter,
			FinishTime:   &timeAfter,
			Duration:     "10s",
			Pods:         []*core.PodRecord{podRecord1},
			ExperimentID: "testID1",
		}
	})

	AfterEach(func() {
		err := mock.ExpectationsWereMet()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("findPodRecordsByEventID", func() {
		It("found", func() {
			mockedRow := []*sqlmock.Rows{
				sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
					AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
						podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action),
				sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
					AddRow(podRecord1.ID, podRecord1.CreatedAt, podRecord1.UpdatedAt, podRecord1.DeletedAt, podRecord1.EventID,
						podRecord1.PodIP, podRecord1.PodName, podRecord1.Namespace, podRecord1.Message, podRecord1.Action),
			}
			sqlSelect := `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(mockedRow[0])

			podRecords, err := es.findPodRecordsByEventID(context.TODO(), podRecord0.EventID)
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
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)
			sqlSelect := `SELECT * FROM "events"`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.List(context.TODO())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			events, err := es.List(context.TODO())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})
	})

	Context("listByUID", func() {
		It("found", func() {
			mockedRow := []*sqlmock.Rows{
				sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
					AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
						event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID),
				sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
					AddRow(event1.ID, event1.CreatedAt, event1.UpdatedAt, event1.DeletedAt, event1.Experiment, event1.Namespace,
						event1.Kind, event1.Message, event1.StartTime, event1.FinishTime, event1.Duration, event1.ExperimentID),
			}

			sqlSelect := `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((experiment_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ExperimentID).WillReturnRows(mockedRow[0])

			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.ListByUID(context.TODO(), event0.ExperimentID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			events, err := es.ListByUID(context.TODO(), "testIDNotFound")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})
	})

	Context("listByExperiment", func() {
		It("found", func() {
			mockedRow := []*sqlmock.Rows{
				sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
					AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
						event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID),
				sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
					AddRow(event1.ID, event1.CreatedAt, event1.UpdatedAt, event1.DeletedAt, event1.Experiment, event1.Namespace,
						event1.Kind, event1.Message, event1.StartTime, event1.FinishTime, event1.Duration, event1.ExperimentID),
			}

			sqlSelect := `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((namespace = ? and experiment = ? ))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Namespace, event0.Experiment).WillReturnRows(mockedRow[0])

			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.ListByExperiment(context.TODO(), event0.Namespace, event0.Experiment)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			events, err := es.ListByExperiment(context.TODO(), "testNamespaceNotFound", "testNameNotFound")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})
	})

	Context("listByNamesoace", func() {
		It("found", func() {
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect := `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND (("pod_records"."namespace" = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Namespace).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)

			sqlSelect = `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((id = ?)) ORDER BY "events"."id" ASC LIMIT 1`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.ListByNamespace(context.TODO(), event0.Namespace)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			events, err := es.ListByNamespace(context.TODO(), "testNamespaceNotFound")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})
	})

	Context("listByPod", func() {
		It("found", func() {
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect := `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND (("pod_records"."pod_name" = ?) AND ("pod_records"."namespace" = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.PodName, podRecord0.Namespace).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)

			sqlSelect = `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((id = ?)) ORDER BY "events"."id" ASC LIMIT 1`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.ListByPod(context.TODO(), podRecord0.Namespace, podRecord0.PodName)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			events, err := es.ListByPod(context.TODO(), "testNamespaceNotFound", "testnameNotFound")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})
	})

	Context("find", func() {
		It("found", func() {
			mockedRow := []*sqlmock.Rows{
				sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
					AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
						event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID),
				sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
					AddRow(event1.ID, event1.CreatedAt, event1.UpdatedAt, event1.DeletedAt, event1.Experiment, event1.Namespace,
						event1.Kind, event1.Message, event1.StartTime, event1.FinishTime, event1.Duration, event1.ExperimentID),
			}

			sqlSelect := `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(mockedRow[0])

			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			event, err := es.Find(context.TODO(), event0.ID)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(event).Should(Equal(event0))
		})

		It("not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			_, err := es.Find(context.TODO(), 30)
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("findByExperimentAndStartTime", func() {
		It("found", func() {
			mockedRow := []*sqlmock.Rows{
				sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
					AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
						event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID),
				sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
					AddRow(event1.ID, event1.CreatedAt, event1.UpdatedAt, event1.DeletedAt, event1.Experiment, event1.Namespace,
						event1.Kind, event1.Message, event1.StartTime, event1.FinishTime, event1.Duration, event1.ExperimentID),
			}

			sqlSelect := `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((namespace = ? and experiment = ? and start_time = ?)) ORDER BY "events"."id" ASC LIMIT 1`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Namespace, event0.Experiment, event0.StartTime).WillReturnRows(mockedRow[0])

			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			event, err := es.FindByExperimentAndStartTime(context.TODO(), event0.Experiment, event0.Namespace, event0.StartTime)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(event.ID).Should(Equal(event0.ID))
		})

		It("event not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			_, err := es.FindByExperimentAndStartTime(context.TODO(), "expNotFound", "namespaceNotFound", event0.StartTime)
			Expect(err).Should(HaveOccurred())
			Expect(gorm.IsRecordNotFoundError(err)).Should(Equal(true))
		})
	})

	Context("listByFilter", func() {
		It("limitStr wrong", func() {
			filter := core.Filter{
				LimitStr: "testWrong",
			}
			_, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).Should(HaveOccurred())
			Expect(strings.Contains(err.Error(), "the format of the limitStr is wrong")).To(Equal(true))
		})

		It("startTimeStr wrong", func() {
			filter := core.Filter{
				StartTimeStr: "testWrong",
			}
			_, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).Should(HaveOccurred())
			Expect(strings.Contains(err.Error(), "the format of the startTime is wrong")).To(Equal(true))
		})

		It("finishTimeStr wrong", func() {
			filter := core.Filter{
				FinishTimeStr: "testWrong",
			}
			_, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).Should(HaveOccurred())
			Expect(strings.Contains(err.Error(), "the format of the finishTime is wrong")).To(Equal(true))
		})

		It("empty args", func() {
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)
			sqlSelect := `SELECT * FROM "events"`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			filter := core.Filter{}
			events, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("podName", func() {
			filter := core.Filter{
				PodName:      "testName",
				PodNamespace: "testNamespace",
				LimitStr:     "1",
			}
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect := `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND (("pod_records"."pod_name" = ?) AND ("pod_records"."namespace" = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.PodName, podRecord0.Namespace).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)

			sqlSelect = `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((id = ?)) ORDER BY "events"."id" ASC LIMIT 1`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("podNamespace", func() {
			filter := core.Filter{
				PodNamespace: "testNamespace",
				LimitStr:     "1",
			}
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect := `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND (("pod_records"."namespace" = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Namespace).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)

			sqlSelect = `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((id = ?)) ORDER BY "events"."id" ASC LIMIT 1`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("experimentName continue", func() {
			filter := core.Filter{
				PodNamespace:   "testNamespace",
				ExperimentName: "experimentNameNotFound",
				LimitStr:       "1",
			}
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect := `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND (("pod_records"."namespace" = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Namespace).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)

			sqlSelect = `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((id = ?)) ORDER BY "events"."id" ASC LIMIT 1`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})

		It("experimentNamespace continue", func() {
			filter := core.Filter{
				PodNamespace:        "testNamespace",
				ExperimentNamespace: "experimentNamespaceNotFound",
				LimitStr:            "1",
			}
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect := `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND (("pod_records"."namespace" = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Namespace).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)

			sqlSelect = `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((id = ?)) ORDER BY "events"."id" ASC LIMIT 1`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})

		It("uid continue", func() {
			filter := core.Filter{
				PodNamespace: "testNamespace",
				UID:          "UIDNotFound",
				LimitStr:     "1",
			}
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect := `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND (("pod_records"."namespace" = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Namespace).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)

			sqlSelect = `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((id = ?)) ORDER BY "events"."id" ASC LIMIT 1`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})

		It("kind continue", func() {
			filter := core.Filter{
				PodNamespace: "testNamespace",
				Kind:         "KindNotFound",
				LimitStr:     "1",
			}
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect := `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND (("pod_records"."namespace" = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Namespace).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)

			sqlSelect = `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((id = ?)) ORDER BY "events"."id" ASC LIMIT 1`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(rows)

			rows = sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "event_id", "pod_ip", "pod_name",
					"namespace", "message", "action"}).
				AddRow(podRecord0.ID, podRecord0.CreatedAt, podRecord0.UpdatedAt, podRecord0.DeletedAt, podRecord0.EventID,
					podRecord0.PodIP, podRecord0.PodName, podRecord0.Namespace, podRecord0.Message, podRecord0.Action)

			sqlSelect = `SELECT * FROM "pod_records" WHERE "pod_records"."deleted_at" IS NULL AND ((event_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(podRecord0.EventID).WillReturnRows(rows)

			events, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})
	})

	Context("dryListByFilter", func() {
		It("limitStr wrong", func() {
			filter := core.Filter{
				LimitStr: "testWrong",
			}
			_, err := es.DryListByFilter(context.TODO(), filter)
			Expect(err).Should(HaveOccurred())
			Expect(strings.Contains(err.Error(), "the format of the limitStr is wrong")).To(Equal(true))
		})

		It("startTimeStr wrong", func() {
			filter := core.Filter{
				StartTimeStr: "testWrong",
			}
			_, err := es.DryListByFilter(context.TODO(), filter)
			Expect(err).Should(HaveOccurred())
			Expect(strings.Contains(err.Error(), "the format of the startTime is wrong")).To(Equal(true))
		})

		It("finishTimeStr wrong", func() {
			filter := core.Filter{
				FinishTimeStr: "testWrong",
			}
			_, err := es.DryListByFilter(context.TODO(), filter)
			Expect(err).Should(HaveOccurred())
			Expect(strings.Contains(err.Error(), "the format of the finishTime is wrong")).To(Equal(true))
		})

		It("empty args", func() {
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)
			sqlSelect := `SELECT * FROM "events"`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WillReturnRows(rows)

			filter := core.Filter{}
			events, err := es.DryListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0].ID).Should(Equal(event0.ID))
		})

		It("test args", func() {
			filter := core.Filter{
				ExperimentName: "testExperiment0",
				UID:            "testID0",
			}
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
					"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID)
			sqlSelect := `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND ((experiment = ? AND experiment_id = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(filter.ExperimentName, filter.UID).WillReturnRows(rows)

			events, err := es.DryListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0].ID).Should(Equal(event0.ID))
		})

		It("not found", func() {
			filter := core.Filter{}
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			events, err := es.DryListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})
	})

	Context("getUID", func() {
		It("found", func() {
			mockedRow := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "experiment", "namespace", "kind",
				"message", "start_time", "finish_time", "duration", "experiment_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.UpdatedAt, event0.DeletedAt, event0.Experiment, event0.Namespace,
					event0.Kind, event0.Message, event0.StartTime, event0.FinishTime, event0.Duration, event0.ExperimentID).
				AddRow(event1.ID, event1.CreatedAt, event1.UpdatedAt, event1.DeletedAt, event1.Experiment, event1.Namespace,
					event1.Kind, event1.Message, event1.StartTime, event1.FinishTime, event1.Duration, event1.ExperimentID)

			sqlSelect := `SELECT * FROM "events" WHERE "events"."deleted_at" IS NULL AND (("events"."experiment" = ?) AND ("events"."namespace" = ?))`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Experiment, event0.Namespace).WillReturnRows(mockedRow)
			uid, err := es.getUID(context.TODO(), event0.Namespace, event0.Experiment)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(uid).Should(Equal(event1.ExperimentID))
		})

		It("not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			uid, err := es.getUID(context.TODO(), "NamespaceNotFound", "NameNotFound")
			Expect(strings.Contains(err.Error(), "get UID failure")).To(Equal(true))
			Expect(uid).Should(Equal(""))
		})
	})
})

func TestConstructQueryArgs(t *testing.T) {
	cases := []struct {
		kind          string
		ns            string
		name          string
		uid           string
		startTime     string
		finishTime    string
		expectedQuery string
		expectedArgs  []string
	}{
		{
			name:          "",
			ns:            "",
			uid:           "",
			kind:          "",
			startTime:     "",
			finishTime:    "",
			expectedQuery: "",
			expectedArgs:  []string{},
		},
		{
			name:          "testName",
			ns:            "",
			uid:           "",
			kind:          "",
			startTime:     "",
			finishTime:    "",
			expectedQuery: "experiment = ?",
			expectedArgs:  []string{"testName"},
		},
		{
			name:          "",
			ns:            "testNamespace",
			uid:           "",
			kind:          "",
			startTime:     "",
			finishTime:    "",
			expectedQuery: "namespace = ?",
			expectedArgs:  []string{"testNamespace"},
		},
		{
			name:          "",
			ns:            "",
			uid:           "testUID",
			kind:          "",
			startTime:     "",
			finishTime:    "",
			expectedQuery: "experiment_id = ?",
			expectedArgs:  []string{"testUID"},
		},
		{
			name:          "",
			ns:            "",
			uid:           "",
			kind:          "testKind",
			startTime:     "",
			finishTime:    "",
			expectedQuery: "kind = ?",
			expectedArgs:  []string{"testKind"},
		},
		{
			name:          "",
			ns:            "",
			uid:           "",
			kind:          "",
			startTime:     "20200101",
			finishTime:    "",
			expectedQuery: "start_time >= ?",
			expectedArgs:  []string{"20200101"},
		},
		{
			name:          "",
			ns:            "",
			uid:           "",
			kind:          "",
			startTime:     "",
			finishTime:    "20200102",
			expectedQuery: "finish_time <= ?",
			expectedArgs:  []string{"20200102"},
		},
		{
			name:          "testName",
			ns:            "testNamespace",
			uid:           "",
			kind:          "",
			startTime:     "",
			finishTime:    "",
			expectedQuery: "experiment = ? AND namespace = ?",
			expectedArgs:  []string{"testName", "testNamespace"},
		},
		{
			name:          "testName",
			ns:            "testNamespace",
			uid:           "testUID",
			kind:          "",
			startTime:     "",
			finishTime:    "",
			expectedQuery: "experiment = ? AND namespace = ? AND experiment_id = ?",
			expectedArgs:  []string{"testName", "testNamespace", "testUID"},
		},
		{
			name:          "testName",
			ns:            "testNamespace",
			uid:           "testUID",
			kind:          "testKind",
			startTime:     "",
			finishTime:    "",
			expectedQuery: "experiment = ? AND namespace = ? AND experiment_id = ? AND kind = ?",
			expectedArgs:  []string{"testName", "testNamespace", "testUID", "testKind"},
		},
		{
			name:          "testName",
			ns:            "testNamespace",
			uid:           "testUID",
			kind:          "testKind",
			startTime:     "20200101",
			finishTime:    "",
			expectedQuery: "experiment = ? AND namespace = ? AND experiment_id = ? AND kind = ? AND start_time >= ?",
			expectedArgs:  []string{"testName", "testNamespace", "testUID", "testKind", "20200101"},
		},
		{
			name:          "testName",
			ns:            "testNamespace",
			uid:           "testUID",
			kind:          "testKind",
			startTime:     "20200101",
			finishTime:    "20200102",
			expectedQuery: "experiment = ? AND namespace = ? AND experiment_id = ? AND kind = ? AND start_time >= ? AND finish_time <= ?",
			expectedArgs:  []string{"testName", "testNamespace", "testUID", "testKind", "20200101", "20200102"},
		},
	}

	for _, c := range cases {
		query, args := constructQueryArgs(c.name, c.ns, c.uid, c.kind, c.startTime, c.finishTime)
		argString := []string{}
		for _, arg := range args {
			argString = append(argString, arg.(string))
		}
		if query != c.expectedQuery {
			t.Errorf("expected query %s but got %s", c.expectedQuery, query)
		}
		if !reflect.DeepEqual(c.expectedArgs, argString) {
			t.Errorf("expected args %v but got %v", c.expectedArgs, argString)
		}
	}
}
