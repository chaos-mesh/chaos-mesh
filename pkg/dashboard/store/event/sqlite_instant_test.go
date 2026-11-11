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
	"database/sql"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SQLite event instants", func() {
	It("matches driver timestamp decoding without changing stored bytes", func() {
		sqlDB, err := sql.Open("sqlite", ":memory:")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(sqlDB.Close)
		_, err = sqlDB.Exec("CREATE TABLE events (id INTEGER PRIMARY KEY, created_at datetime)")
		Expect(err).NotTo(HaveOccurred())
		values := []any{
			"2026-11-01 01:02:03.123456789+08:00",
			"2026-11-01T01:02:03.123456789+08:00",
			"2026-11-01 01:02:03.123456789",
			"2026-11-01T01:02:03.123456789Z",
			"2026-11-01 01:02:03", "2026-11-01T01:02:03",
			"2026-11-01 01:02", "2026-11-01T01:02", "2026-11-01",
			"2026-11-01 01:02:03.123456789 +0800 CST",
			" 2026-11-01 01:02:03.123456789 +0800 CST m=+3.123 ",
			"2026-10-31T17:02:03.123456789Z",
			"2026-10-31T17:02:03.123456790Z",
			"1960-01-01T00:00:00.000000001Z",
			"0001-01-01T00:00:00Z", "2400-01-01T00:00:00.000000001Z",
			nil, "malformed", "2026-11-01T00:00:00Z\x00junk",
			"2026-11-01T00:00:00ZZ", " 2026-11-01T00:00:00Z ", "",
			int64(42), 1.5, []byte("2026-11-01T00:00:00Z"), []byte{},
		}
		for _, value := range values {
			_, err := sqlDB.Exec("DELETE FROM events")
			Expect(err).NotTo(HaveOccurred())
			_, err = sqlDB.Exec("INSERT INTO events (id, created_at) VALUES (1, ?)", value)
			Expect(err).NotTo(HaveOccurred())
			var before, after []byte
			Expect(sqlDB.QueryRow("SELECT CAST(created_at AS BLOB) FROM events").Scan(&before)).To(Succeed())
			var decoded *time.Time
			driverErr := sqlDB.QueryRow("SELECT created_at FROM events").Scan(&decoded)
			var key []byte
			keyErr := sqlDB.QueryRow("SELECT " + sqliteInstantExpression + " FROM events").Scan(&key)
			if driverErr != nil {
				Expect(keyErr).To(HaveOccurred(), "value: %#v", value)
			} else {
				Expect(keyErr).NotTo(HaveOccurred(), "value: %#v", value)
				if decoded == nil {
					Expect(key).To(BeNil())
				} else {
					Expect(key).To(Equal(instantKey(*decoded)), "value: %#v", value)
				}
			}
			Expect(sqlDB.QueryRow("SELECT CAST(created_at AS BLOB) FROM events").Scan(&after)).To(Succeed())
			Expect(after).To(Equal(before))
		}
	})

	It("orders exact instants beyond UnixNano's range", func() {
		values := []time.Time{
			time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Unix(-1, 999999999), time.Unix(0, 0), time.Unix(0, 1),
			time.Date(2400, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		for index := 1; index < len(values); index++ {
			Expect(string(instantKey(values[index-1])) < string(instantKey(values[index]))).To(BeTrue())
		}
		Expect(instantKey(values[4].In(time.FixedZone("offset", 8*3600)))).To(Equal(instantKey(values[4])))
	})

	It("registers on fresh pooled connections before any store opens", func() {
		db, err := sql.Open("sqlite", ":memory:")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(db.Close)
		first, err := db.Conn(context.Background())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(first.Close)
		second, err := db.Conn(context.Background())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(second.Close)
		for _, conn := range []*sql.Conn{first, second} {
			var key []byte
			Expect(conn.QueryRowContext(context.Background(), "SELECT chaos_mesh_event_instant_v1(CAST(? AS BLOB), 'text')", "2400-01-01T00:00:00.000000001Z").Scan(&key)).To(Succeed())
			Expect(key).To(Equal(instantKey(time.Date(2400, 1, 1, 0, 0, 0, 1, time.UTC))))
		}
	})
})
