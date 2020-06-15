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
	// List returns an event list from the datastore.
	List(context.Context) ([]*Event, error)

	// ListByFilter returns an event list by the podName, podNamespace, experimentName, experimentNamespace and the startTime.
	ListByFilter(context.Context, string, string, string, string, string) ([]*Event, error)

	// ListByExperiment returns an event list by the name and namespace of the experiment.
	ListByExperiment(context.Context, string, string) ([]*Event, error)

	// ListByNamespace returns an event list by the namespace of the pod.
	ListByNamespace(context.Context, string) ([]*Event, error)

	// ListByPod returns an event list by the name and namespace of the pod.
	ListByPod(context.Context, string, string) ([]*Event, error)

	// Find returns an event from the datastore by ID.
	Find(context.Context, uint) (*Event, error)

	// FindByExperimentAndStartTime returns an event by the experiment and start time.
	FindByExperimentAndStartTime(context.Context, string, string, *time.Time) (*Event, error)

	// Create persists a new event to the datastore.
	Create(context.Context, *Event) error

	// Update persists an updated event to the datastore.
	Update(context.Context, *Event) error

	// DeleteIncompleteEvent deletes all incomplete events.
	// If the chaos-dashboard was restarted, some incomplete events would be stored in dbtastore,
	// which means the event would never save the finish_time.
	// DeleteIncompleteEvent can be used to delete all incomplete events to avoid this case.
	DeleteIncompleteEvents(context.Context) error

	// DeleteByFinishTime deletes events and podrecords whose time difference is greater than the given time from FinishTime.
	DeleteByFinishTime(context.Context, time.Duration) error
}

// Event represents an event instance.
type Event struct {
	gorm.Model
	Experiment   string `gorm:"index:experiment"`
	Namespace    string
	Kind         string
	Message      string
	ExperimentID string     `gorm:"index:experiment_id"`
	StartTime    *time.Time `gorm:"index:start_time"`
	FinishTime   *time.Time
	Pods         []*PodRecord `gorm:"-"`
}

// PodRecord represents a pod record with event ID.
type PodRecord struct {
	gorm.Model
	EventID   uint   `gorm:"index:event_id"`
	PodIP     string `gorm:"index:pod_id"`
	PodName   string
	Namespace string
	Message   string
	Action    string
}
