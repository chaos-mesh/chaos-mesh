// Copyright 2020 PingCAP, Inc.
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
	"github.com/jinzhu/gorm"

	"github.com/pingcap/chaos-mesh/pkg/store/dbstore"
)

// EventStore defines operations for working with event.
type EventStore interface{}

// NewStore return a new EventStore.
func NewStore(db *dbstore.DB) EventStore {
	db.AutoMigrate(&Event{})

	return &eventStore{db}
}

type eventStore struct {
	db *dbstore.DB
}

// Event represents a event instance.
type Event struct {
	gorm.Model
}
