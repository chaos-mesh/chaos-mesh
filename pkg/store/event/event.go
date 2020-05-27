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

// findPodRecordsByEventID returns the list of PodRecords according to the eventID
func (e *eventStore) findPodRecordsByEventID(_ context.Context, id uint) ([]*core.PodRecord, error) {
	pods := make([]*core.PodRecord, 0)
	if err := e.db.Where(
		"event_id = ?", id).
		Find(&pods).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	return pods, nil
}

// List returns the list of events
func (e *eventStore) List(_ context.Context) ([]*core.Event, error) {
	var resList []core.Event
	eventList := make([]*core.Event, 0)

	if err := e.db.Find(&resList).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	for _, et := range resList {
		pods, err := e.findPodRecordsByEventID(context.Background(), et.ID)
		if err != nil {
			return nil, err
		}
		var event core.Event
		event = et
		event.Pods = pods
		eventList = append(eventList, &event)
	}

	return eventList, nil
}

// ListByExperiment returns a event list by the name and namespace of the experiment.
func (e *eventStore) ListByExperiment(_ context.Context, namespace string, experiment string) ([]*core.Event, error) {
	var resList []core.Event

	if err := e.db.Where(
		"namespace = ? and experiment = ? ",
		namespace, experiment).
		Find(&resList).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	eventList := make([]*core.Event, 0, len(resList))
	for _, et := range resList {
		pods, err := e.findPodRecordsByEventID(context.Background(), et.ID)
		if err != nil {
			return nil, err
		}
		var event core.Event
		event = et
		event.Pods = pods
		eventList = append(eventList, &event)
	}

	return eventList, nil
}

// ListByNamespace returns the list of events according to the namespace
func (e *eventStore) ListByNamespace(_ context.Context, namespace string) ([]*core.Event, error) {
	podRecords := make([]*core.PodRecord, 0)

	if err := e.db.Where(
		&core.PodRecord{Namespace: namespace}).
		Find(&podRecords).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	et := new(core.Event)
	eventList := make([]*core.Event, 0, len(podRecords))
	for _, pr := range podRecords {
		if err := e.db.Where(
			"id = ?", pr.EventID).
			First(et).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}

		pods, err := e.findPodRecordsByEventID(context.Background(), et.ID)
		if err != nil {
			return nil, err
		}
		et.Pods = pods
		eventList = append(eventList, et)
	}
	return eventList, nil
}

// ListByPod returns the list of events according to the pod
func (e *eventStore) ListByPod(_ context.Context, namespace string, name string) ([]*core.Event, error) {
	podRecords := make([]*core.PodRecord, 0)

	if err := e.db.Where(
		&core.PodRecord{PodName: name, Namespace: namespace}).
		Find(&podRecords).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	et := new(core.Event)
	eventList := make([]*core.Event, 0, len(podRecords))
	for _, pr := range podRecords {
		if err := e.db.Where(
			"id = ?", pr.EventID).
			First(et).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}

		pods, err := e.findPodRecordsByEventID(context.Background(), et.ID)
		if err != nil {
			return nil, err
		}
		et.Pods = pods
		eventList = append(eventList, et)
	}
	return eventList, nil
}

// Find returns a event from the datastore by ID.
func (e *eventStore) Find(_ context.Context, id uint) (*core.Event, error) {
	et := new(core.Event)
	if err := e.db.Where(
		"id = ?", id).
		First(et).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	pods, err := e.findPodRecordsByEventID(context.Background(), et.ID)
	if err != nil {
		return nil, err
	}
	et.Pods = pods
	return et, nil
}

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
