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
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	sqlite "github.com/glebarez/go-sqlite"
)

const sqliteInstantExpression = "chaos_mesh_event_instant_v1(NULLIF(CAST(created_at AS BLOB), X''), typeof(created_at))"

func init() {
	// The driver installs registered functions when opening each connection.
	sqlite.MustRegisterDeterministicScalarFunction("chaos_mesh_event_instant_v1", 2,
		func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
			if args[1] == "null" {
				return nil, nil
			}
			// TEXT arguments lose embedded NULs in go-sqlite's UDF bridge.
			// CAST preserves bytes, NULLIF avoids the driver's empty-BLOB panic,
			// and typeof preserves the original storage class.
			raw, ok := args[0].([]byte)
			if !ok || args[1] != "text" {
				return nil, fmt.Errorf("invalid event created_at storage class: %v", args[1])
			}
			instant, err := parseSQLiteInstant(string(raw))
			if err != nil {
				return nil, err
			}
			return instantKey(instant), nil
		})
}

func instantKey(instant time.Time) []byte {
	key := make([]byte, 12)
	binary.BigEndian.PutUint64(key, uint64(instant.Unix())^(uint64(1)<<63))
	binary.BigEndian.PutUint32(key[8:], uint32(instant.Nanosecond()))
	return key
}

// Match github.com/glebarez/go-sqlite v1.22.0's private parseTime and
// parseTimeString. Driver parity tests must pass when upgrading that dependency.
func parseSQLiteInstant(value string) (time.Time, error) {
	stringValue := value
	if index := strings.Index(stringValue, "m="); index > 0 {
		stringValue = stringValue[:index]
	}
	if instant, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", strings.TrimSpace(stringValue)); err == nil {
		return instant, nil
	}
	value = strings.TrimSuffix(value, "Z")
	for _, layout := range []string{
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02T15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	} {
		if instant, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return instant, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid event created_at timestamp")
}
