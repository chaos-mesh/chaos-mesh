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

	"github.com/jinzhu/gorm"
)

// PodRecordStore defines operations for working with pod records
type PodRecordStore interface {
	// List returns a pod record list from the datastore.
	List(context.Context) ([]*PodRecord, error)

	// ListByPod returns a pod record list by the name and namespace of the pod.
	ListByPod(context.Context, string, string) ([]*PodRecord, error)

	// ListByEvent returns a pod record list by the event ID.
	ListByEvent(context.Context, int64) ([]*PodRecord, error)

	// Find returns a pod record from the datastore by ID.
	Find(context.Context, int64) (*PodRecord, error)

	// Create persists a new pod record to the datastore.
	Create(context.Context, *PodRecord) error

	// Update persists an updated pod record to the datastore.
	Update(context.Context, *PodRecord) error
}

// PodRecord represents a pod record with event ID.
type PodRecord struct {
	gorm.Model
	EventID   int64
	PodIP     string
	PodName   string
	Namespace string
}
