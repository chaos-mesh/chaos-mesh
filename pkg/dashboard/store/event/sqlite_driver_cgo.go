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

//go:build cgo

package event

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
)

func init() {
	const name = "chaos-mesh-sqlite3"
	sql.Register(name, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			return conn.RegisterFunc("chaos_mesh_event_instant_v1", sqliteInstant, true)
		},
	})
	sqliteDriverName = name
}

func sqliteInstant(value any, storageClass string) ([]byte, error) {
	switch storageClass {
	case "null":
		return nil, nil
	case "integer":
		// Match the SQLite driver's interpretation of integer DATETIME values.
		n := value.(int64)
		if n > 1e12 || n < -1e12 {
			return instantKey(time.Unix(0, n*int64(time.Millisecond))), nil
		}
		return instantKey(time.Unix(n, 0)), nil
	case "text":
		// The SQL expression passes text as a BLOB so embedded NULs survive the callback.
		text := strings.TrimSuffix(string(value.([]byte)), "Z")
		for _, layout := range sqlite3.SQLiteTimestampFormats {
			if instant, err := time.ParseInLocation(layout, text, time.UTC); err == nil {
				return instantKey(instant), nil
			}
		}
		return nil, fmt.Errorf("invalid event created_at timestamp")
	default:
		return nil, fmt.Errorf("invalid event created_at storage class: %s", storageClass)
	}
}
