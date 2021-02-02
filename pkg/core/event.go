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
	ListByExperiment(context.Context, string, string) ([]*Event, error)

	// ListByNamespace returns an event list by the namespace of the pod.
	ListByNamespace(context.Context, string) ([]*Event, error)

	// ListByPod returns an event list by the name and namespace of the pod.
	ListByPod(context.Context, string, string) ([]*Event, error)

	// ListByUID returns an event list by the UID.
	ListByUID(context.Context, string) ([]*Event, error)

	// DryListByFilter returns an event list by experimentName, experimentNamespace, uid, kind, startTime and finishTime.
	DryListByFilter(context.Context, Filter) ([]*Event, error)

	// Find returns an event from the datastore by ID.
	Find(context.Context, uint) (*Event, error)

	// FindByExperimentAndStartTime returns an event by the experiment and start time.
	FindByExperimentAndStartTime(context.Context, string, string, *time.Time) (*Event, error)

	// Create persists a new event to the datastore.
	Create(context.Context, *Event) error

	// Update persists an updated event to the datastore.
	Update(context.Context, *Event) error

	// DeleteIncompleteEvent deletes all incomplete events.
	// If the chaos-dashboard was restarted, some incomplete events would be stored in datastore,
	// which means the event would never save the finish_time.
	// DeleteIncompleteEvent can be used to delete all incomplete events to avoid this case.
	DeleteIncompleteEvents(context.Context) error

	// DeleteByFinishTime deletes events and podrecords whose time difference is greater than the given time from FinishTime.
	DeleteByFinishTime(context.Context, time.Duration) error

	// DeleteByUID deletes events list by the UID.
	DeleteByUID(context.Context, string) error

	// UpdateIncompleteEvents updates the incomplete event by the namespace and name
	// If chaos is deleted before an event is over, then the incomplete event would be stored in datastore,
	// which means the event would never save the finish_time.
	// UpdateIncompleteEvents can update the finish_time when the chaos is deleted.
	UpdateIncompleteEvents(context.Context, string, string) error
}

// Event represents an event instance.
type Event struct {
	ID           uint         `gorm:"primary_key" json:"id"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	DeletedAt    *time.Time   `sql:"index" json:"deleted_at"`
	Experiment   string       `gorm:"index:experiment" json:"experiment"`
	Namespace    string       `json:"namespace"`
	Kind         string       `json:"kind"`
	Message      string       `json:"message"`
	StartTime    *time.Time   `gorm:"index:start_time" json:"start_time"`
	FinishTime   *time.Time   `json:"finish_time"`
	Duration     string       `json:"duration"`
	Pods         []*PodRecord `gorm:"-" json:"pods"`
	ExperimentID string       `gorm:"index:experiment_id" json:"experiment_id"`
}

// PodRecord represents a pod record with event ID.
type PodRecord struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"deleted_at"`
	EventID   uint       `gorm:"index:event_id" json:"event_id"`
	PodIP     string     `gorm:"index:pod_id" json:"pod_ip"`
	PodName   string     `json:"pod_name"`
	Namespace string     `json:"namespace"`
	Message   string     `json:"message"`
	Action    string     `json:"action"`
}

// Filter represents the filter to list events
type Filter struct {
	PodName             string
	PodNamespace        string
	StartTimeStr        string
	FinishTimeStr       string
	ExperimentName      string
	ExperimentNamespace string
	UID                 string
	Kind                string
	LimitStr            string
}
