// Copyright 2026 Chaos Mesh Authors.
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

package event

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

const sqliteTimeBatchSize = 256

// SQLite's textual timestamp ordering does not compare offsets as instants.
// Keep the driver's time.Time decoding and its full fractional precision.
type eventTime struct {
	ID        uint
	CreatedAt time.Time
}

func sqliteTimeBound(value string) (*time.Time, error) {
	if value == "" || value == "0001-01-01 00:00:00" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02 15:04:05", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func listSQLiteByTime(statement *gorm.DB, filter core.Filter, start, end *time.Time) ([]*core.Event, error) {
	limit := -1
	if filter.Limit != "" {
		var err error
		limit, err = strconv.Atoi(filter.Limit)
		if err != nil {
			return nil, err
		}
	}
	if err := statement.Statement.Context.Err(); err != nil {
		return nil, err
	}
	events := make([]*core.Event, 0)
	if limit == 0 || (start != nil && end != nil && start.After(*end)) {
		return events, nil
	}

	filter.Start, filter.End = "", ""
	query, args := filter.ConstructQueryArgs()
	if query != "" {
		statement = statement.Where(query, args...)
	}
	err := statement.Transaction(func(tx *gorm.DB) error {
		base := tx.Session(&gorm.Session{}).Model(&core.Event{}).
			Select("id", "created_at").Where("created_at IS NOT NULL").Order("id desc")
		var cursor *uint
		for {
			page := base.Session(&gorm.Session{}).Limit(sqliteTimeBatchSize)
			if cursor != nil {
				page = page.Where("id < ?", *cursor)
			}
			var candidates []eventTime
			if err := page.Find(&candidates).Error; err != nil {
				return err
			}
			ids := make([]uint, 0, len(candidates))
			for _, candidate := range candidates {
				if (start != nil && candidate.CreatedAt.Before(*start)) || (end != nil && candidate.CreatedAt.After(*end)) {
					continue
				}
				ids = append(ids, candidate.ID)
				if limit > 0 && len(events)+len(ids) == limit {
					break
				}
			}
			if len(ids) > 0 {
				var matches []*core.Event
				if err := tx.Session(&gorm.Session{NewDB: true}).Where("id IN (?)", ids).Order("id desc").Find(&matches).Error; err != nil {
					return err
				}
				events = append(events, matches...)
			}
			if len(candidates) < sqliteTimeBatchSize || (limit > 0 && len(events) >= limit) {
				return nil
			}
			cursor = &candidates[len(candidates)-1].ID
		}
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}

func deleteSQLiteBefore(statement *gorm.DB, cutoff time.Time) error {
	return statement.Transaction(func(tx *gorm.DB) error {
		base := tx.Session(&gorm.Session{}).Model(&core.Event{}).
			Select("id", "created_at").Where("created_at IS NOT NULL").Order("id desc")
		var cursor *uint
		for {
			page := base.Session(&gorm.Session{}).Limit(sqliteTimeBatchSize)
			if cursor != nil {
				page = page.Where("id < ?", *cursor)
			}
			var candidates []eventTime
			if err := page.Find(&candidates).Error; err != nil {
				return err
			}
			ids := make([]uint, 0, len(candidates))
			for _, candidate := range candidates {
				if !candidate.CreatedAt.After(cutoff) {
					ids = append(ids, candidate.ID)
				}
			}
			if len(ids) > 0 {
				if err := tx.Session(&gorm.Session{NewDB: true}).Where("id IN (?)", ids).Delete(&core.Event{}).Error; err != nil {
					return err
				}
			}
			if len(candidates) < sqliteTimeBatchSize {
				return nil
			}
			cursor = &candidates[len(candidates)-1].ID
		}
	})
}
