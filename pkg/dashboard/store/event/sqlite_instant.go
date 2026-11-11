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
	"encoding/binary"
	"time"

	"gorm.io/driver/sqlite"
)

const sqliteInstantExpression = "chaos_mesh_event_instant_v1(CASE WHEN typeof(created_at) = 'text' THEN CAST(created_at AS BLOB) ELSE created_at END, typeof(created_at))"

var sqliteDriverName = sqlite.DriverName

// SQLiteDialector opens connections with the event timestamp function when CGO is available.
func SQLiteDialector(dsn string) sqlite.Dialector {
	return sqlite.Dialector{DriverName: sqliteDriverName, DSN: dsn}
}

func instantKey(instant time.Time) []byte {
	key := make([]byte, 12)
	binary.BigEndian.PutUint64(key, uint64(instant.Unix())^(uint64(1)<<63))
	binary.BigEndian.PutUint32(key[8:], uint32(instant.Nanosecond()))
	return key
}
