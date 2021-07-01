// Copyright 2021 Chaos Mesh Authors.
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

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// ScheduleStore defines operations for working with schedules.
type ScheduleStore interface {
	// ListMeta returns schedule metadata list from the datastore.
	ListMeta(ctx context.Context, namespace, name string, archived bool) ([]*ScheduleMeta, error)

	// FindByUID returns a schedule by UID.
	FindByUID(ctx context.Context, UID string) (*Schedule, error)

	// FindMetaByUID returns a schedule metadata by UID.
	FindMetaByUID(context.Context, string) (*ScheduleMeta, error)

	// Set saves the schedule to datastore.
	Set(context.Context, *Schedule) error

	// Archive archives schedules which "archived" field is false.
	Archive(ctx context.Context, namespace, name string) error

	// Delete deletes the archive from the datastore.
	Delete(context.Context, *Schedule) error

	// DeleteByFinishTime deletes archives which time difference is greater than the given time from FinishTime.
	DeleteByFinishTime(context.Context, time.Duration) error

	// DeleteByUIDs deletes archives by the uid list.
	DeleteByUIDs(context.Context, []string) error

	// DeleteIncompleteSchedules deletes all incomplete schedules.
	// If the chaos-dashboard was restarted and the schedule is completed during the restart,
	// which means the schedule would never save the finish_time.
	// DeleteIncompleteSchedules can be used to delete all incomplete schedules to avoid this case.
	DeleteIncompleteSchedules(context.Context) error
}

// Schedule represents a schedule instance. Use in db.
type Schedule struct {
	ScheduleMeta
	Schedule string `gorm:"size:2048"` // JSON string
}

// ScheduleMeta defines the metadata of a schedule instance. Use in db.
type ScheduleMeta struct {
	gorm.Model
	UID        string    `gorm:"index:schedule_uid" json:"uid"`
	Kind       string    `json:"kind"`
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Action     string    `json:"action"`
	StartTime  time.Time `json:"start_time"`
	FinishTime time.Time `json:"finish_time"`
	Archived   bool      `json:"archived"`
}

// ScheduleInfo defines a form data of schedule from API.
type ScheduleInfo struct {
	Name                    string                     `json:"name" binding:"required,NameValid"`
	Namespace               string                     `json:"namespace" binding:"required,NameValid"`
	Labels                  map[string]string          `json:"labels" binding:"MapSelectorsValid"`
	Annotations             map[string]string          `json:"annotations" binding:"MapSelectorsValid"`
	Scope                   ScopeInfo                  `json:"scope"`
	Target                  TargetInfo                 `json:"target"`
	Schedule                string                     `json:"schedule"`
	Duration                string                     `json:"duration" binding:"DurationValid"`
	StartingDeadlineSeconds *int64                     `json:"starting_deadline_seconds,omitempty"`
	ConcurrencyPolicy       v1alpha1.ConcurrencyPolicy `json:"concurrency_policy"`
	HistoryLimit            int                        `json:"history_limit,omitempty"`
}
