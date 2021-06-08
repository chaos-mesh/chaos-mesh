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

package core

import (
	"context"
	"time"
)

// EventStore defines operations for working with event.
type EventStore interface {
	// List returns an event list from the datastore.
	List(context.Context) ([]*Event, error)

	// ListByFilter returns an event list by podName, podNamespace, experimentName, experimentNamespace, uid, kind, startTime and finishTime.
	ListByFilter(context.Context, Filter) ([]*Event, error)

	// ListByExperiment returns an event list by the name and namespace of the experiment.
	ListByExperiment(context.Context, string, string, string) ([]*Event, error)

	// ListByUID returns an event list by the UID.
	ListByUID(context.Context, string) ([]*Event, error)

	// ListByUIDs returns an event list by the UID list.
	ListByUIDs(context.Context, []string) ([]*Event, error)

	// Find returns an event from the datastore by ID.
	Find(context.Context, uint) (*Event, error)

	// Create persists a new event to the datastore.
	Create(context.Context, *Event) error

	// DeleteByCreateTime deletes events whose time difference is greater than the given time from CreateTime.
	DeleteByCreateTime(context.Context, time.Duration) error

	// DeleteByUID deletes events list by the UID.
	DeleteByUID(context.Context, string) error

	// DeleteByUIDs deletes events list by the UID list.
	DeleteByUIDs(context.Context, []string) error
}

// Event represents an event instance.
type Event struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Kind      string    `json:"kind"`
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	ObjectID  string    `gorm:"index:object_id" json:"object_id"`
}

// Filter represents the filter to list events
type Filter struct {
	CreateTimeStr string
	Name          string
	Namespace     string
	ObjectID      string
	Kind          string
	LimitStr      string
}
