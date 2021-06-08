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
		es      *eventStore
		mock    sqlmock.Sqlmock
		event0  *core.Event
		event1  *core.Event
		timeNow time.Time
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

		event0 = &core.Event{
			ID:        0,
			CreatedAt: timeNow,
			Kind:      "testKind",
			Type:      "testType",
			Reason:    "testReason",
			Message:   "testMessage",
			Name:      "testName",
			Namespace: "testNamespace",
			ObjectID:  "testID0",
		}
		event1 = &core.Event{
			ID:        1,
			CreatedAt: timeNow,
			Kind:      "testKind",
			Type:      "testType",
			Reason:    "testReason",
			Message:   "testMessage",
			Name:      "testName",
			Namespace: "testNamespace",
			ObjectID:  "testID1",
		}
	})

	AfterEach(func() {
		err := mock.ExpectationsWereMet()
		Expect(err).ShouldNot(HaveOccurred())
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
				NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
					"namespace", "object_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.Kind, event0.Type, event0.Reason,
					event0.Message, event0.Name, event0.Namespace, event0.ObjectID)

			sqlSelect := `SELECT * FROM "events"`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WillReturnRows(rows)

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
				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
					"namespace", "object_id"}).
					AddRow(event0.ID, event0.CreatedAt, event0.Kind, event0.Type, event0.Reason,
						event0.Message, event0.Name, event0.Namespace, event0.ObjectID),
				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
					"namespace", "object_id"}).
					AddRow(event1.ID, event1.CreatedAt, event1.Kind, event1.Type, event1.Reason,
						event1.Message, event1.Name, event1.Namespace, event1.ObjectID),
			}

			sqlSelect := `SELECT * FROM "events" WHERE (object_id = ?)`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ObjectID).WillReturnRows(mockedRow[0])

			events, err := es.ListByUID(context.TODO(), event0.ObjectID)
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
				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
					"namespace", "object_id"}).
					AddRow(event0.ID, event0.CreatedAt, event0.Kind, event0.Type, event0.Reason,
						event0.Message, event0.Name, event0.Namespace, event0.ObjectID),
				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
					"namespace", "object_id"}).
					AddRow(event1.ID, event1.CreatedAt, event1.Kind, event1.Type, event1.Reason,
						event1.Message, event1.Name, event1.Namespace, event1.ObjectID),
			}

			sqlSelect := `SELECT * FROM "events" WHERE (namespace = ? and name = ? and kind = ?)`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Namespace, event0.Name, event0.Kind).WillReturnRows(mockedRow[0])

			events, err := es.ListByExperiment(context.TODO(), event0.Namespace, event0.Name, event0.Kind)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})

		It("not found", func() {
			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
			events, err := es.ListByExperiment(context.TODO(), "testNamespaceNotFound", "testNameNotFound", "testKindNotFound")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(events)).Should(Equal(0))
		})
	})

	Context("find", func() {
		It("found", func() {
			mockedRow := []*sqlmock.Rows{
				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
					"namespace", "object_id"}).
					AddRow(event0.ID, event0.CreatedAt, event0.Kind, event0.Type, event0.Reason,
						event0.Message, event0.Name, event0.Namespace, event0.ObjectID),
				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
					"namespace", "object_id"}).
					AddRow(event1.ID, event1.CreatedAt, event1.Kind, event1.Type, event1.Reason,
						event1.Message, event1.Name, event1.Namespace, event1.ObjectID),
			}

			sqlSelect := `SELECT * FROM "events" WHERE (id = ?)`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(mockedRow[0])

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
				CreateTimeStr: "testWrong",
			}
			_, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).Should(HaveOccurred())
			Expect(strings.Contains(err.Error(), "the format of the createTime is wrong")).To(Equal(true))
		})

		It("empty args", func() {
			rows := sqlmock.
				NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
					"namespace", "object_id"}).
				AddRow(event0.ID, event0.CreatedAt, event0.Kind, event0.Type, event0.Reason,
					event0.Message, event0.Name, event0.Namespace, event0.ObjectID)
			sqlSelect := `SELECT * FROM "events"`
			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WillReturnRows(rows)

			filter := core.Filter{}
			events, err := es.ListByFilter(context.TODO(), filter)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(events[0]).Should(Equal(event0))
		})
	})
})

func TestConstructQueryArgs(t *testing.T) {
	cases := []struct {
		kind          string
		ns            string
		name          string
		uid           string
		createTime    string
		expectedQuery string
		expectedArgs  []string
	}{
		{
			name:          "",
			ns:            "",
			uid:           "",
			kind:          "",
			createTime:    "",
			expectedQuery: "",
			expectedArgs:  []string{},
		},
		{
			name:          "testName",
			ns:            "",
			uid:           "",
			kind:          "",
			createTime:    "",
			expectedQuery: "name = ?",
			expectedArgs:  []string{"testName"},
		},
		{
			name:          "",
			ns:            "testNamespace",
			uid:           "",
			kind:          "",
			createTime:    "",
			expectedQuery: "namespace = ?",
			expectedArgs:  []string{"testNamespace"},
		},
		{
			name:          "",
			ns:            "",
			uid:           "testUID",
			kind:          "",
			createTime:    "",
			expectedQuery: "object_id = ?",
			expectedArgs:  []string{"testUID"},
		},
		{
			name:          "",
			ns:            "",
			uid:           "",
			kind:          "testKind",
			createTime:    "",
			expectedQuery: "kind = ?",
			expectedArgs:  []string{"testKind"},
		},
		{
			name:          "",
			ns:            "",
			uid:           "",
			kind:          "",
			createTime:    "20200101",
			expectedQuery: "created_at >= ?",
			expectedArgs:  []string{"20200101"},
		},
		{
			name:          "testName",
			ns:            "testNamespace",
			uid:           "",
			kind:          "",
			createTime:    "",
			expectedQuery: "name = ? AND namespace = ?",
			expectedArgs:  []string{"testName", "testNamespace"},
		},
		{
			name:          "testName",
			ns:            "testNamespace",
			uid:           "testUID",
			kind:          "",
			createTime:    "",
			expectedQuery: "name = ? AND namespace = ? AND object_id = ?",
			expectedArgs:  []string{"testName", "testNamespace", "testUID"},
		},
		{
			name:          "testName",
			ns:            "testNamespace",
			uid:           "testUID",
			kind:          "testKind",
			createTime:    "",
			expectedQuery: "name = ? AND namespace = ? AND object_id = ? AND kind = ?",
			expectedArgs:  []string{"testName", "testNamespace", "testUID", "testKind"},
		},
		{
			name:          "testName",
			ns:            "testNamespace",
			uid:           "testUID",
			kind:          "testKind",
			createTime:    "20200101",
			expectedQuery: "name = ? AND namespace = ? AND object_id = ? AND kind = ? AND created_at >= ?",
			expectedArgs:  []string{"testName", "testNamespace", "testUID", "testKind", "20200101"},
		},
		{
			name:          "testName",
			ns:            "testNamespace",
			uid:           "testUID",
			kind:          "testKind",
			createTime:    "20200101",
			expectedQuery: "name = ? AND namespace = ? AND object_id = ? AND kind = ? AND created_at >= ?",
			expectedArgs:  []string{"testName", "testNamespace", "testUID", "testKind", "20200101"},
		},
	}

	for _, c := range cases {
		query, args := constructQueryArgs(c.name, c.ns, c.uid, c.kind, c.createTime)
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
