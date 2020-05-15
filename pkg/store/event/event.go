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
	"context"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/pingcap/chaos-mesh/pkg/core"
	"github.com/pingcap/chaos-mesh/pkg/store/dbstore"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("eventStore")

// NewStore return a new EventStore.
func NewStore(db *dbstore.DB) core.EventStore {
	db.AutoMigrate(&core.Event{})
	db.AutoMigrate(&core.PodRecord{})

	es := &eventStore{db}
	if err := es.DeleteIncompleteEvents(context.Background()); err != nil && gorm.IsRecordNotFoundError(err) {
		log.Error(err, "failed to delete all incomplete events")
	}

	return es
}

type eventStore struct {
	db *dbstore.DB
}

// TODO: implement core.EventStore interface
func (e *eventStore) List(context.Context) ([]*core.Event, error) { return nil, nil }
func (e *eventStore) ListByExperiment(context.Context, string, string) ([]*core.Event, error) {
	return nil, nil
}
func (e *eventStore) ListByPod(context.Context, string, string) ([]*core.Event, error) {
	return nil, nil
}
func (e *eventStore) Find(context.Context, int64) (*core.Event, error) { return nil, nil }

func (e *eventStore) FindByExperimentAndStartTime(
	_ context.Context,
	name, namespace string,
	startTime *time.Time,
) (*core.Event, error) {
	et := new(core.Event)
	if err := e.db.Where(
		"namespace = ? and experiment = ? and start_time = ?",
		namespace, name, startTime).
		First(et).Error; err != nil {
		return nil, err
	}

	var pods []*core.PodRecord

	if err := e.db.Where(
		"event_id = ?", et.ID).
		Find(&pods).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return et, nil
}

// Create persists a new event to the datastore.
func (e *eventStore) Create(_ context.Context, et *core.Event) error {
	if err := e.db.Create(et).Error; err != nil {
		return err
	}

	for _, pod := range et.Pods {
		pod.EventID = et.ID
		if err := e.db.Create(pod).Error; err != nil {
			return err
		}
	}

	return nil
}

// Update persists an updated event to the datastore.
func (e *eventStore) Update(_ context.Context, et *core.Event) error {
	return e.db.Model(core.Event{}).
		Where(
			"namespace = ? and experiment = ? and start_time = ?",
			et.Namespace, et.Experiment, et.StartTime).
		Update("finish_time", et.FinishTime).
		Error
}

// DeleteIncompleteEvents implement core.EventStore interface.
func (e *eventStore) DeleteIncompleteEvents(_ context.Context) error {
	return e.db.Where("finish_time IS NULL").
		Delete(core.Event{}).Error
}
