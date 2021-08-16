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

// func TestEvent(t *testing.T) {
// 	RegisterFailHandler(Fail)
// 	RunSpecs(t, "Event Suite")
// }

// var _ = Describe("event", func() {
// 	var (
// 		es      *eventStore
// 		mock    sqlmock.Sqlmock
// 		event0  *core.Event
// 		event1  *core.Event
// 		timeNow time.Time
// 	)

// 	BeforeEach(func() {
// 		var db *sql.DB
// 		var err error
// 		db, mock, err = sqlmock.New()
// 		Expect(err).ShouldNot(HaveOccurred())

// 		gdb, err := gorm.Open("sqlite3", db)
// 		Expect(err).ShouldNot(HaveOccurred())

// 		es = &eventStore{db: &dbstore.DB{DB: gdb}}

// 		timeNow = time.Now()

// 		event0 = &core.Event{
// 			ID:        0,
// 			CreatedAt: timeNow,
// 			Kind:      "testKind",
// 			Type:      "testType",
// 			Reason:    "testReason",
// 			Message:   "testMessage",
// 			Name:      "testName",
// 			Namespace: "testNamespace",
// 			ObjectID:  "testID0",
// 		}
// 		event1 = &core.Event{
// 			ID:        1,
// 			CreatedAt: timeNow,
// 			Kind:      "testKind",
// 			Type:      "testType",
// 			Reason:    "testReason",
// 			Message:   "testMessage",
// 			Name:      "testName",
// 			Namespace: "testNamespace",
// 			ObjectID:  "testID1",
// 		}
// 	})

// 	AfterEach(func() {
// 		err := mock.ExpectationsWereMet()
// 		Expect(err).ShouldNot(HaveOccurred())
// 	})

// 	Context("list", func() {
// 		It("found", func() {
// 			rows := sqlmock.
// 				NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
// 					"namespace", "object_id"}).
// 				AddRow(event0.ID, event0.CreatedAt, event0.Kind, event0.Type, event0.Reason,
// 					event0.Message, event0.Name, event0.Namespace, event0.ObjectID)

// 			sqlSelect := `SELECT * FROM "events"`
// 			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WillReturnRows(rows)

// 			events, err := es.List(context.TODO())
// 			Expect(err).ShouldNot(HaveOccurred())
// 			Expect(events[0]).Should(Equal(event0))
// 		})

// 		It("not found", func() {
// 			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
// 			events, err := es.List(context.TODO())
// 			Expect(err).ShouldNot(HaveOccurred())
// 			Expect(len(events)).Should(Equal(0))
// 		})
// 	})

// 	Context("listByUID", func() {
// 		It("found", func() {
// 			mockedRow := []*sqlmock.Rows{
// 				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
// 					"namespace", "object_id"}).
// 					AddRow(event0.ID, event0.CreatedAt, event0.Kind, event0.Type, event0.Reason,
// 						event0.Message, event0.Name, event0.Namespace, event0.ObjectID),
// 				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
// 					"namespace", "object_id"}).
// 					AddRow(event1.ID, event1.CreatedAt, event1.Kind, event1.Type, event1.Reason,
// 						event1.Message, event1.Name, event1.Namespace, event1.ObjectID),
// 			}

// 			sqlSelect := `SELECT * FROM "events" WHERE (object_id = ?)`
// 			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ObjectID).WillReturnRows(mockedRow[0])

// 			events, err := es.ListByUID(context.TODO(), event0.ObjectID)
// 			Expect(err).ShouldNot(HaveOccurred())
// 			Expect(events[0]).Should(Equal(event0))
// 		})

// 		It("not found", func() {
// 			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
// 			events, err := es.ListByUID(context.TODO(), "testIDNotFound")
// 			Expect(err).ShouldNot(HaveOccurred())
// 			Expect(len(events)).Should(Equal(0))
// 		})
// 	})

// 	Context("listByExperiment", func() {
// 		It("found", func() {
// 			mockedRow := []*sqlmock.Rows{
// 				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
// 					"namespace", "object_id"}).
// 					AddRow(event0.ID, event0.CreatedAt, event0.Kind, event0.Type, event0.Reason,
// 						event0.Message, event0.Name, event0.Namespace, event0.ObjectID),
// 				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
// 					"namespace", "object_id"}).
// 					AddRow(event1.ID, event1.CreatedAt, event1.Kind, event1.Type, event1.Reason,
// 						event1.Message, event1.Name, event1.Namespace, event1.ObjectID),
// 			}

// 			sqlSelect := `SELECT * FROM "events" WHERE (namespace = ? and name = ? and kind = ?)`
// 			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.Namespace, event0.Name, event0.Kind).WillReturnRows(mockedRow[0])

// 			events, err := es.ListByExperiment(context.TODO(), event0.Namespace, event0.Name, event0.Kind)
// 			Expect(err).ShouldNot(HaveOccurred())
// 			Expect(events[0]).Should(Equal(event0))
// 		})

// 		It("not found", func() {
// 			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
// 			events, err := es.ListByExperiment(context.TODO(), "testNamespaceNotFound", "testNameNotFound", "testKindNotFound")
// 			Expect(err).ShouldNot(HaveOccurred())
// 			Expect(len(events)).Should(Equal(0))
// 		})
// 	})

// 	Context("find", func() {
// 		It("found", func() {
// 			mockedRow := []*sqlmock.Rows{
// 				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
// 					"namespace", "object_id"}).
// 					AddRow(event0.ID, event0.CreatedAt, event0.Kind, event0.Type, event0.Reason,
// 						event0.Message, event0.Name, event0.Namespace, event0.ObjectID),
// 				sqlmock.NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
// 					"namespace", "object_id"}).
// 					AddRow(event1.ID, event1.CreatedAt, event1.Kind, event1.Type, event1.Reason,
// 						event1.Message, event1.Name, event1.Namespace, event1.ObjectID),
// 			}

// 			sqlSelect := `SELECT * FROM "events" WHERE (id = ?)`
// 			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(event0.ID).WillReturnRows(mockedRow[0])

// 			event, err := es.Find(context.TODO(), event0.ID)
// 			Expect(err).ShouldNot(HaveOccurred())
// 			Expect(event).Should(Equal(event0))
// 		})

// 		It("not found", func() {
// 			mock.ExpectQuery(`.+`).WillReturnRows(sqlmock.NewRows(nil))
// 			_, err := es.Find(context.TODO(), 30)
// 			Expect(err).Should(HaveOccurred())
// 		})
// 	})

// 	Context("listByFilter", func() {
// 		It("limitStr wrong", func() {
// 			filter := core.Filter{
// 				Limit: "testWrong",
// 			}
// 			_, err := es.ListByFilter(context.TODO(), filter)
// 			Expect(err).Should(HaveOccurred())
// 			Expect(strings.Contains(err.Error(), "the format of the limitStr is wrong")).To(Equal(true))
// 		})

// 		It("empty args", func() {
// 			rows := sqlmock.
// 				NewRows([]string{"id", "created_at", "kind", "type", "reason", "message", "name",
// 					"namespace", "object_id"}).
// 				AddRow(event0.ID, event0.CreatedAt, event0.Kind, event0.Type, event0.Reason,
// 					event0.Message, event0.Name, event0.Namespace, event0.ObjectID)
// 			sqlSelect := `SELECT * FROM "events"`
// 			mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WillReturnRows(rows)

// 			filter := core.Filter{}
// 			events, err := es.ListByFilter(context.TODO(), filter)
// 			Expect(err).ShouldNot(HaveOccurred())
// 			Expect(events[0]).Should(Equal(event0))
// 		})
// 	})
// })
