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

package experiment

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestExperiment(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Experiment Suite")
}

func genRows() *sqlmock.Rows {
	return sqlmock.NewRows(
		[]string{"id", "created_at", "delete_at", "uid", "namespace", "name", "kind", "action"},
	)
}

func addRow(rows *sqlmock.Rows, exp *core.Experiment) {
	rows.AddRow(exp.ID, exp.CreatedAt, exp.DeletedAt, exp.UID, exp.Namespace, exp.Name,
		exp.Kind, exp.Action)
}

var _ = Describe("Experiment", func() {
	var (
		err         error
		db          *sql.DB
		mock        sqlmock.Sqlmock
		es          *experimentStore
		experiment0 *core.Experiment
	)

	BeforeEach(func() {
		db, mock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		Expect(err).ShouldNot(HaveOccurred())

		gdb, err := gorm.Open("sqlite3", db)
		Expect(err).ShouldNot(HaveOccurred())

		es = &experimentStore{db: gdb}

		now := time.Now()
		experiment0 = &core.Experiment{
			ExperimentMeta: core.ExperimentMeta{
				Model: gorm.Model{
					ID:        0,
					CreatedAt: now,
					DeletedAt: nil,
				},
				UID:       "UID0",
				Namespace: "default",
				Name:      "experiment0",
				Kind:      "PodChaos",
				Action:    "pod-failure",
			},
		}
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
	})

	Context("FindByUID", func() {
		sql := "SELECT * FROM \"experiments\" WHERE (uid = ?) ORDER BY \"experiments\".\"id\" ASC LIMIT 1"

		It("experiment0 should be found", func() {
			rows := genRows()
			addRow(rows, experiment0)

			mock.ExpectQuery(sql).WithArgs(experiment0.UID).WillReturnRows(rows)

			exp, err := es.FindByUID(context.TODO(), "UID0")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(exp).Should(Equal(experiment0))
		})
	})
})
