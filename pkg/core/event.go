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

package core

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
)

// EventStore defines operations for working with event.
type EventStore interface {
	// List returns a event list from the datastore.
	List(context.Context) ([]*Event, error)

	// ListByExperiment returns a event list by the name and namespace of the experiment.
	ListByExperiment(context.Context, string, string) ([]*Event, error)

	// Find returns a event from the datastore by ID.
	Find(context.Context, int64) (*Event, error)

	// Create persists a new event to the datastore.
	Create(context.Context, *Event) error

	// Update persists an updated event to the datastore.
	Update(context.Context, *Event) error
}

// Event represents a event instance.
type Event struct {
	gorm.Model
	Experiment string
	Namespace  string
	Kind       string
	Message    string
	StartTime  *time.Time
	FinishTime *time.Time
}
