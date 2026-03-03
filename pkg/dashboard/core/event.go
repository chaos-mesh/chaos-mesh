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

package core

import (
	"context"
	"time"
)

// EventStore defines operations for working with events.
type EventStore interface {
	// List returns an event list from the datastore.
	List(context.Context) ([]*Event, error)

	// ListByUID returns an event list by the UID.
	ListByUID(context.Context, string) ([]*Event, error)

	// ListByUIDs returns an event list by the UID list.
	ListByUIDs(context.Context, []string) ([]*Event, error)

	// ListByExperiment returns an event list by the namespace, name, or kind.
	ListByExperiment(context context.Context, namespace string, name string, kind string) ([]*Event, error)

	ListByFilter(context.Context, Filter) ([]*Event, error)

	// Find returns an event by ID.
	Find(context.Context, uint) (*Event, error)

	// Create persists a new event to the datastore.
	Create(context.Context, *Event) error

	// DeleteByUID deletes events by the UID.
	DeleteByUID(context.Context, string) error

	// DeleteByUIDs deletes events by the UID list.
	DeleteByUIDs(context.Context, []string) error

	// DeleteByTime deletes events within the specified time interval.
	DeleteByTime(context.Context, string, string) error

	// DeleteByDuration selete events that exceed duration.
	DeleteByDuration(context.Context, time.Duration) error
}

type Event struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	ObjectID  string    `gorm:"index:object_id" json:"object_id"`
	CreatedAt time.Time `json:"created_at"`
	Namespace string    `json:"namespace"`
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `gorm:"type:text;size:32768" json:"message"`
}
