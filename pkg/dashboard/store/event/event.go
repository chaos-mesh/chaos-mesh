// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package event

import (
	"context"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func NewStore(db *gorm.DB) core.EventStore {
	db.AutoMigrate(&core.Event{})

	return &eventStore{db}
}

type eventStore struct {
	db *gorm.DB
}

func (e *eventStore) List(_ context.Context) ([]*core.Event, error) {
	var events []*core.Event

	if err := e.db.Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (e *eventStore) ListBy(_ context.Context, by string, args ...interface{}) ([]*core.Event, error) {
	var events []*core.Event

	if err := e.db.Where(by, args...).Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (e *eventStore) ListByUID(c context.Context, uid string) ([]*core.Event, error) {
	return e.ListBy(c, "object_id = ?", uid)
}

func (e *eventStore) ListByUIDs(c context.Context, uids []string) ([]*core.Event, error) {
	return e.ListBy(c, "object_id IN (?)", uids)
}

func (e *eventStore) ListByExperiment(c context.Context, namespace string, name string, kind string) ([]*core.Event, error) {
	return e.ListBy(c, "namespace = ? AND name = ? AND kind = ?", namespace, name, kind)
}

func (e *eventStore) ListByFilter(_ context.Context, filter core.Filter) ([]*core.Event, error) {
	var (
		events []*core.Event
		limit  int
		err    error
	)

	query, args := filter.ConstructQueryArgs()
	statement := e.db.Where(query, args...).Order("id desc")

	if filter.Limit != "" {
		limit, err = strconv.Atoi(filter.Limit)
		if err != nil {
			return nil, err
		}

		statement = statement.Limit(limit)
	}

	if err := statement.Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (e *eventStore) Find(_ context.Context, id uint) (*core.Event, error) {
	event := new(core.Event)

	if err := e.db.First(event, id).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (e *eventStore) Create(_ context.Context, event *core.Event) error {
	return e.db.Create(event).Error
}

func (e *eventStore) DeleteByUID(_ context.Context, uid string) error {
	return e.db.Where("object_id = ?", uid).Delete(&core.Event{}).Error
}

func (e *eventStore) DeleteByUIDs(_ context.Context, uids []string) error {
	return e.db.Where("object_id IN (?)", uids).Delete(&core.Event{}).Error
}

func (e *eventStore) DeleteByTime(_ context.Context, start string, end string) error {
	return e.db.Where("created_at BETWEEN ? AND ?", start, end).Delete(&core.Event{}).Error
}

func (e *eventStore) DeleteByDuration(_ context.Context, duration time.Duration) error {
	now := time.Now().UTC().Add(-duration).Format("2006-01-02 15:04:05")

	return e.db.Where("created_at <= ?", now).Delete(&core.Event{}).Error
}
