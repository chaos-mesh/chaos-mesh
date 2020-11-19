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
	"github.com/chaos-mesh/chaos-mesh/pkg/store/dbstore"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/store/event"

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
	var (
		s          *Service
		mock       sqlmock.Sqlmock

	)

	BeforeEach(func() {
		var db *sql.DB
		var err error
		db, mock, err = sqlmock.New()
		Expect(err).ShouldNot(HaveOccurred())

		gdb, err := gorm.Open("sqlite3", db)
		Expect(err).ShouldNot(HaveOccurred())

		s = &Service{
			conf:    nil,
			kubeCli: nil,
			archive: nil,
			event:   &event.eventStore{db: &dbstore.DB{DB: gdb},
		}
		//s = &Service{db: &dbstore.DB{DB: gdb}}

	})

	AfterEach(func() {
		err := mock.ExpectationsWereMet()
		Expect(err).ShouldNot(HaveOccurred())
	})

})

