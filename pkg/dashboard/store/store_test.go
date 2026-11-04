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

package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"go.uber.org/fx/fxtest"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
)

func TestSQLiteLifecycle(t *testing.T) {
	for _, tc := range []struct{ name, prefix, query string }{
		{"plain path", "", ""}, {"URI", "file:", ""},
		{"existing query", "", "?mode=rwc"}, {"existing cache", "file:", "?cache=private"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lc := fxtest.NewLifecycle(t)
			conf := &config.ChaosDashboardConfig{Database: &config.DatabaseConfig{Driver: "sqlite3", Datasource: tc.prefix + filepath.Join(t.TempDir(), "dashboard.sqlite") + tc.query}}
			db, err := Bootstrap(lc, conf, logr.Discard())
			if err != nil {
				t.Fatal(err)
			}
			lc.RequireStart()
			t.Cleanup(lc.RequireStop)
			if err := db.Exec("CREATE TABLE probe (value TEXT)").Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Exec("INSERT INTO probe (value) VALUES (?)", "saved").Error; err != nil {
				t.Fatal(err)
			}
			var got string
			if err := db.Raw("SELECT value FROM probe").Scan(&got).Error; err != nil || got != "saved" {
				t.Fatalf("read %q, error %v", got, err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			if sqlDB.Stats().MaxOpenConnections != 1 {
				t.Fatalf("SQLite max connections %d", sqlDB.Stats().MaxOpenConnections)
			}
			if err := lc.Stop(context.Background()); err != nil {
				t.Fatal(err)
			}
			if err := sqlDB.Ping(); err == nil {
				t.Fatal("database remains open after lifecycle stop")
			}
		})
	}
}

func TestDatabaseInitializationErrors(t *testing.T) {
	for _, tc := range []struct{ driver, source string }{
		{"unsupported", "ignored"}, {"sqlite3", filepath.Join(t.TempDir(), "missing", "dashboard.sqlite")},
	} {
		lc := fxtest.NewLifecycle(t)
		conf := &config.ChaosDashboardConfig{Database: &config.DatabaseConfig{Driver: tc.driver, Datasource: tc.source}}
		if db, err := NewDBStore(lc, conf, logr.Discard()); err == nil || db != nil {
			t.Fatalf("driver %q: db %v, error %v", tc.driver, db, err)
		}
	}
}
