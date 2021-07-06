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
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/store/dbstore"
)

// NewStore return a new EventStore.
func NewStore(db *dbstore.DB) core.EventStore {
	db.AutoMigrate(&core.Event{})

	return &eventStore{db}
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

// List returns the list of events
func (e *eventStore) List(_ context.Context) ([]*core.Event, error) {
	var resList []core.Event
	eventList := make([]*core.Event, 0)

	if err := e.db.Find(&resList).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	for _, et := range resList {
		var event core.Event = et
		eventList = append(eventList, &event)
	}

	return eventList, nil
}

// ListByUID returns an event list by the uid of the experiment.
func (e *eventStore) ListByUID(_ context.Context, uid string) ([]*core.Event, error) {
	var resList []core.Event
	eventList := make([]*core.Event, 0)

	if err := e.db.Where(
		"object_id = ?", uid).
		Find(&resList).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	for _, et := range resList {
		var event core.Event = et
		eventList = append(eventList, &event)
	}

	return eventList, nil
}

// ListByUIDs returns an event list by the uids of the experiments.
func (e *eventStore) ListByUIDs(_ context.Context, uids []string) ([]*core.Event, error) {
	var resList []core.Event
	eventList := make([]*core.Event, 0)

	if err := e.db.Table("events").Where(
		"object_id IN (?)", uids).
		Find(&resList).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	for _, et := range resList {
		var event core.Event = et
		eventList = append(eventList, &event)
	}

	return eventList, nil
}

// ListByExperiment returns an event list by the name and namespace of the experiment.
func (e *eventStore) ListByExperiment(_ context.Context, namespace string, experiment string, kind string) ([]*core.Event, error) {
	var resList []core.Event

	if err := e.db.Where(
		"namespace = ? and name = ? and kind = ?",
		namespace, experiment, kind).
		Find(&resList).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	eventList := make([]*core.Event, 0, len(resList))
	for _, et := range resList {
		var event core.Event = et
		eventList = append(eventList, &event)
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

	return et, nil
}

// Create persists a new event to the datastore.
func (e *eventStore) Create(_ context.Context, et *core.Event) error {
	return e.db.Create(et).Error
}

// ListByFilter returns an event list by experimentName, experimentNamespace, uid, kind, creatTime.
func (e *eventStore) ListByFilter(_ context.Context, filter core.Filter) ([]*core.Event, error) {
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
	if filter.CreateTimeStr != "" {
		_, err = time.Parse(time.RFC3339, strings.Replace(filter.CreateTimeStr, " ", "+", -1))
		if err != nil {
			return nil, fmt.Errorf("the format of the createTime is wrong")
		}
	}

	query, args := constructQueryArgs(filter.Name, filter.Namespace, filter.ObjectID, filter.Kind, filter.CreateTimeStr)
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

// DeleteByCreateTime deletes events whose time difference is greater than the given time from CreateTime.
func (e *eventStore) DeleteByCreateTime(_ context.Context, ttl time.Duration) error {
	eventList, err := e.List(context.Background())
	if err != nil {
		return err
	}
	nowTime := time.Now()
	for _, et := range eventList {
		if et.CreatedAt.Add(ttl).Before(nowTime) {
			if err := e.db.Model(core.Event{}).Unscoped().Delete(*et).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteByUID deletes events by the uid of the experiment.
func (e *eventStore) DeleteByUID(_ context.Context, uid string) error {
	return e.db.Where("object_id = ?", uid).Unscoped().
		Delete(core.Event{}).Error
}

// DeleteByUIDs deletes events by the uid list of the experiment.
func (e *eventStore) DeleteByUIDs(_ context.Context, uids []string) error {
	return e.db.Where("object_id IN (?)", uids).Unscoped().Delete(core.Event{}).Error
}

func constructQueryArgs(experimentName, experimentNamespace, uid, kind, createTime string) (string, []interface{}) {
	args := make([]interface{}, 0)
	query := ""
	if experimentName != "" {
		query += "name = ?"
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
			query += " AND object_id = ?"
		} else {
			query += "object_id = ?"
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
	if createTime != "" {
		if len(args) > 0 {
			query += " AND created_at >= ?"
		} else {
			query += "created_at >= ?"
		}
		args = append(args, strings.Replace(createTime, "T", " ", -1))
	}

	return query, args
}
