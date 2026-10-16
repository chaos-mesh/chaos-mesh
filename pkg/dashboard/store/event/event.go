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
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func NewStore(db *gorm.DB) (core.EventStore, error) {
	if err := db.AutoMigrate(&core.Event{}); err != nil {
		return nil, fmt.Errorf("migrate event table: %w", err)
	}

	return &eventStore{db}, nil
}

type eventStore struct {
	db *gorm.DB
}

func (e *eventStore) List(ctx context.Context) ([]*core.Event, error) {
	var events []*core.Event

	if err := e.db.WithContext(ctx).Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (e *eventStore) ListBy(ctx context.Context, by string, args ...any) ([]*core.Event, error) {
	var events []*core.Event

	if err := e.db.WithContext(ctx).Where(by, args...).Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (e *eventStore) ListByUID(c context.Context, uid string) ([]*core.Event, error) {
	return e.ListBy(c, "object_id = ?", uid)
}

func (e *eventStore) ListByUIDList(c context.Context, uids []string) ([]*core.Event, error) {
	return e.ListBy(c, "object_id IN (?)", uids)
}

func (e *eventStore) ListByExperiment(c context.Context, namespace string, name string, kind string) ([]*core.Event, error) {
	return e.ListBy(c, "namespace = ? AND name = ? AND kind = ?", namespace, name, kind)
}

func (e *eventStore) ListByUIDListWithFilter(ctx context.Context, uids []string, filter core.Filter) ([]*core.Event, error) {
	return listByFilter(e.db.WithContext(ctx).Where("object_id IN (?)", uids), filter)
}

func (e *eventStore) ListByFilter(ctx context.Context, filter core.Filter) ([]*core.Event, error) {
	return listByFilter(e.db.WithContext(ctx), filter)
}

func listByFilter(statement *gorm.DB, filter core.Filter) ([]*core.Event, error) {
	if statement.Dialector.Name() == "sqlite" {
		start, err := sqliteTimeBound(filter.Start)
		if err != nil {
			return nil, err
		}
		end, err := sqliteTimeBound(filter.End)
		if err != nil {
			return nil, err
		}
		if start != nil || end != nil {
			return listSQLiteByTime(statement, filter, start, end)
		}
	}

	var (
		events []*core.Event
		limit  int
		err    error
	)

	query, args := filter.ConstructQueryArgs()
	if query != "" {
		statement = statement.Where(query, args...)
	}

	statement = statement.Order("id desc")

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

func (e *eventStore) Find(ctx context.Context, id uint) (*core.Event, error) {
	event := new(core.Event)

	if err := e.db.WithContext(ctx).First(event, id).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (e *eventStore) Create(ctx context.Context, event *core.Event) error {
	return e.db.WithContext(ctx).Create(event).Error
}

func (e *eventStore) DeleteByUID(ctx context.Context, uid string) error {
	return e.db.WithContext(ctx).Where("object_id = ?", uid).Delete(&core.Event{}).Error
}

func (e *eventStore) DeleteByUIDList(ctx context.Context, uids []string) error {
	return e.db.WithContext(ctx).Where("object_id IN (?)", uids).Delete(&core.Event{}).Error
}

func (e *eventStore) DeleteByDuration(ctx context.Context, duration time.Duration) error {
	now := time.Now().UTC().Add(-duration)
	if e.db.Dialector.Name() == "sqlite" {
		return deleteSQLiteBefore(e.db.WithContext(ctx), now)
	}

	return e.db.WithContext(ctx).Where("created_at <= ?", now).Delete(&core.Event{}).Error
}
