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

package event

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/store/dbstore"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("store/event")

// NewStore return a new EventStore.
func NewStore(db *dbstore.DB) core.EventStore {
	db.AutoMigrate(&core.Event{})
	db.AutoMigrate(&core.PodRecord{})

	es := &eventStore{db}

	if err := es.DeleteIncompleteEvents(context.Background()); err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Error(err, "failed to delete all incomplete events")
	}

	return es
}

type eventStore struct {
	db *dbstore.DB
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
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
		var event core.Event = et
		event.Pods = pods
		eventList = append(eventList, &event)
	}

	return eventList, nil
}

// ListByUID returns an event list by the uid of the experiment.
func (e *eventStore) ListByUID(_ context.Context, uid string) ([]*core.Event, error) {
	var resList []core.Event
	eventList := make([]*core.Event, 0)

	if err := e.db.Where(
		"experiment_id = ?", uid).
		Find(&resList).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	for _, et := range resList {
		pods, err := e.findPodRecordsByEventID(context.Background(), et.ID)
		if err != nil {
			return nil, err
		}
		var event core.Event = et
		event.Pods = pods
		eventList = append(eventList, &event)
	}

	return eventList, nil
}

// ListByExperiment returns an event list by the name and namespace of the experiment.
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
		var event core.Event = et
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

	eventList := make([]*core.Event, 0, len(podRecords))
	for _, pr := range podRecords {
		et := new(core.Event)
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

	eventList := make([]*core.Event, 0, len(podRecords))
	for _, pr := range podRecords {
		et := new(core.Event)
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

// Find returns an event from the datastore by ID.
func (e *eventStore) Find(_ context.Context, id uint) (*core.Event, error) {
	et := new(core.Event)
	if err := e.db.Where(
		"id = ?", id).
		First(et).Error; err != nil {
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
	return e.db.Where("finish_time IS NULL").Unscoped().
		Delete(core.Event{}).Error
}

// ListByFilter returns an event list by podName, podNamespace, experimentName, experimentNamespace, uid, kind, startTime and finishTime.
func (e *eventStore) ListByFilter(_ context.Context, filter core.Filter) ([]*core.Event, error) {
	var (
		resList               []*core.Event
		err                   error
		startTime, finishTime time.Time
		limit                 int
	)

	if filter.LimitStr != "" {
		limit, err = strconv.Atoi(filter.LimitStr)
		if err != nil {
			return nil, fmt.Errorf("the format of the limitStr is wrong")
		}
	}
	if filter.StartTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, strings.Replace(filter.StartTimeStr, " ", "+", -1))
		if err != nil {
			return nil, fmt.Errorf("the format of the startTime is wrong")
		}
	}
	if filter.FinishTimeStr != "" {
		finishTime, err = time.Parse(time.RFC3339, strings.Replace(filter.FinishTimeStr, " ", "+", -1))
		if err != nil {
			return nil, fmt.Errorf("the format of the finishTime is wrong")
		}
	}

	if filter.PodName != "" {
		resList, err = e.ListByPod(context.Background(), filter.PodNamespace, filter.PodName)
		if err == nil && filter.LimitStr != "" {
			sort.Slice(resList, func(i, j int) bool {
				return resList[i].CreatedAt.After(resList[j].CreatedAt)
			})
			resList = resList[:min(limit, len(resList))]
		}
	} else if filter.PodNamespace != "" {
		resList, err = e.ListByNamespace(context.Background(), filter.PodNamespace)
		if err == nil && filter.LimitStr != "" {
			sort.Slice(resList, func(i, j int) bool {
				return resList[i].CreatedAt.After(resList[j].CreatedAt)
			})
			resList = resList[:min(limit, len(resList))]
		}
	} else {
		resList, err = e.DryListByFilter(context.Background(), filter)
	}
	if err != nil {
		return resList, err
	}

	eventList := make([]*core.Event, 0)
	for _, event := range resList {
		if filter.PodName != "" || filter.PodNamespace != "" {
			if filter.ExperimentName != "" && event.Experiment != filter.ExperimentName {
				continue
			}
			if filter.ExperimentNamespace != "" && event.Namespace != filter.ExperimentNamespace {
				continue
			}
			if filter.UID != "" && event.ExperimentID != filter.UID {
				continue
			}
			if filter.Kind != "" && event.Kind != filter.Kind {
				continue
			}
			if filter.StartTimeStr != "" && event.StartTime.Before(startTime) && !event.StartTime.Equal(startTime) {
				continue
			}
			if filter.FinishTimeStr != "" && event.FinishTime.After(finishTime) && !event.FinishTime.Equal(finishTime) {
				continue
			}
		}
		if filter.PodNamespace == "" {
			pods, err := e.findPodRecordsByEventID(context.Background(), event.ID)
			if err != nil {
				return nil, err
			}
			event.Pods = pods
		}
		eventList = append(eventList, event)
	}
	return eventList, nil
}

// DryListByFilter returns an event list by experimentName, experimentNamespace, uid, kind, startTime and finishTime.
func (e *eventStore) DryListByFilter(_ context.Context, filter core.Filter) ([]*core.Event, error) {
	var (
		resList []*core.Event
		err     error
		db      *dbstore.DB
		limit   int
	)

	if filter.LimitStr != "" {
		limit, err = strconv.Atoi(filter.LimitStr)
		if err != nil {
			return nil, fmt.Errorf("the format of the limitStr is wrong")
		}
	}
	if filter.StartTimeStr != "" {
		_, err = time.Parse(time.RFC3339, strings.Replace(filter.StartTimeStr, " ", "+", -1))
		if err != nil {
			return nil, fmt.Errorf("the format of the startTime is wrong")
		}
	}
	if filter.FinishTimeStr != "" {
		_, err = time.Parse(time.RFC3339, strings.Replace(filter.FinishTimeStr, " ", "+", -1))
		if err != nil {
			return nil, fmt.Errorf("the format of the finishTime is wrong")
		}
	}

	query, args := constructQueryArgs(filter.ExperimentName, filter.ExperimentNamespace, filter.UID, filter.Kind, filter.StartTimeStr, filter.FinishTimeStr)
	// List all events
	if len(args) == 0 {
		db = e.db
	} else {
		db = &dbstore.DB{DB: e.db.Where(query, args...)}
	}
	if filter.LimitStr != "" {
		db = &dbstore.DB{DB: db.Order("created_at desc").Limit(limit)}
	}
	if err := db.Find(&resList).Error; err != nil &&
		!gorm.IsRecordNotFoundError(err) {
		return resList, err
	}

	return resList, err
}

// DeleteByFinishTime deletes events and podrecords whose time difference is greater than the given time from FinishTime.
func (e *eventStore) DeleteByFinishTime(_ context.Context, ttl time.Duration) error {
	eventList, err := e.List(context.Background())
	if err != nil {
		return err
	}
	nowTime := time.Now()
	for _, et := range eventList {
		if et.FinishTime == nil {
			continue
		}
		if et.FinishTime.Add(ttl).Before(nowTime) {
			if err := e.db.Model(core.Event{}).Unscoped().Delete(*et).Error; err != nil {
				return err
			}

			if err := e.db.Model(core.PodRecord{}).
				Where(
					"event_id = ? ",
					et.ID).Unscoped().Delete(core.PodRecord{}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteByUID deletes events by the uid of the experiment.
func (e *eventStore) DeleteByUID(_ context.Context, uid string) error {
	eventList, err := e.ListByUID(context.Background(), uid)
	if err != nil {
		return err
	}
	for _, et := range eventList {
		if err := e.db.Model(core.PodRecord{}).
			Where(
				"event_id = ? ",
				et.ID).Unscoped().Delete(core.PodRecord{}).Error; err != nil {
			return err
		}
	}
	return e.db.Where("experiment_id = ?", uid).Unscoped().
		Delete(core.Event{}).Error
}

func (e *eventStore) getUID(_ context.Context, ns, name string) (string, error) {
	events := make([]*core.Event, 0)

	if err := e.db.Where(
		&core.Event{Experiment: name, Namespace: ns}).
		Find(&events).Error; err != nil {
		return "", err
	}

	if len(events) == 0 {
		return "", fmt.Errorf("get UID failure, maybe name or namespace is wrong")
	}

	UID := events[0].ExperimentID
	st := events[0].StartTime

	for _, et := range events {
		if st.Before(*et.StartTime) {
			st = et.StartTime
			UID = et.ExperimentID
		}
	}
	return UID, nil
}

// UpdateIncompleteEvents updates the incomplete event by the namespace and name
func (e *eventStore) UpdateIncompleteEvents(_ context.Context, ns, name string) error {
	return e.db.Model(core.Event{}).
		Where(
			"namespace = ? and experiment = ? and finish_time IS NULL",
			ns, name).
		Update("finish_time", time.Now()).
		Error
}

func constructQueryArgs(experimentName, experimentNamespace, uid, kind, startTime, finishTime string) (string, []interface{}) {
	args := make([]interface{}, 0)
	query := ""
	if experimentName != "" {
		query += "experiment = ?"
		args = append(args, experimentName)
	}
	if experimentNamespace != "" {
		if len(args) > 0 {
			query += " AND namespace = ?"
		} else {
			query += "namespace = ?"
		}
		args = append(args, experimentNamespace)
	}
	if uid != "" {
		if len(args) > 0 {
			query += " AND experiment_id = ?"
		} else {
			query += "experiment_id = ?"
		}
		args = append(args, uid)
	}
	if kind != "" {
		if len(args) > 0 {
			query += " AND kind = ?"
		} else {
			query += "kind = ?"
		}
		args = append(args, kind)
	}
	if startTime != "" {
		if len(args) > 0 {
			query += " AND start_time >= ?"
		} else {
			query += "start_time >= ?"
		}
		args = append(args, strings.Replace(startTime, "T", " ", -1))
	}
	if finishTime != "" {
		if len(args) > 0 {
			query += " AND finish_time <= ?"
		} else {
			query += "finish_time <= ?"
		}
		args = append(args, strings.Replace(finishTime, "T", " ", -1))
	}

	return query, args
}
