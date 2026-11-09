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
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func (e *eventStore) Query(ctx context.Context, query core.EventQuery) (core.EventResult, error) {
	if err := ctx.Err(); err != nil {
		return core.EventResult{}, err
	}
	result := core.EventResult{Events: make([]*core.Event, 0)}
	run := func(db *gorm.DB) error {
		base := eventPredicates(db.Model(&core.Event{}), query)
		if query.IncludeTotal {
			var total int64
			if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
				return err
			}
			result.Total = &total
		}
		if query.Window != nil && query.Window.Limit == 0 {
			return nil
		}
		statement := base.Session(&gorm.Session{})
		column := eventSortColumn(query.Sort)
		desc := query.Order != core.EventOrderAsc
		expression := ""
		if column == "created_at" && db.Dialector.Name() == "sqlite" {
			expression = sqliteInstantExpression
		} else if column == "message" && db.Dialector.Name() == "sqlserver" {
			// SQL Server cannot sort the legacy text column directly.
			expression = "CAST(message AS varchar(max))"
		}
		if expression != "" {
			direction := " ASC"
			if desc {
				direction = " DESC"
			}
			statement = statement.Order(expression + direction)
		} else {
			statement = statement.Order(clause.OrderByColumn{Column: clause.Column{Name: column}, Desc: desc})
		}
		if column != "id" {
			statement = statement.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: desc})
		}
		if query.Window != nil {
			statement = statement.Offset(query.Window.Offset).Limit(query.Window.Limit + 1)
		}
		if err := statement.Find(&result.Events).Error; err != nil {
			return err
		}
		if query.Window != nil && len(result.Events) > query.Window.Limit {
			result.HasMore = true
			result.Events = result.Events[:query.Window.Limit]
		}
		return nil
	}
	db := e.db.WithContext(ctx)
	var err error
	// SQLite count and page share a read snapshot. Other databases retain their
	// configured isolation, so concurrent writes can change the count between statements.
	if query.IncludeTotal && db.Dialector.Name() == "sqlite" {
		err = db.Transaction(run)
	} else {
		err = run(db)
	}
	if err != nil {
		return core.EventResult{}, err
	}
	return result, nil
}

func eventPredicates(db *gorm.DB, query core.EventQuery) *gorm.DB {
	for _, filter := range []struct {
		column string
		values []string
	}{
		{"namespace", query.Namespaces},
		{"name", query.Names},
		{"kind", query.Kinds},
		{"object_id", query.ObjectIDs},
	} {
		if len(filter.values) > 0 {
			db = db.Where(filter.column+" IN ?", filter.values)
		}
	}
	column := "created_at"
	if db.Dialector.Name() == "sqlite" {
		column = sqliteInstantExpression
	}
	if query.Start != nil {
		var bound any = *query.Start
		if db.Dialector.Name() == "sqlite" {
			bound = instantKey(*query.Start)
		}
		db = db.Where(column+" >= ?", bound)
	}
	if query.End != nil {
		var bound any = *query.End
		if db.Dialector.Name() == "sqlite" {
			bound = instantKey(*query.End)
		}
		db = db.Where(column+" <= ?", bound)
	}
	return db
}

func eventSortColumn(sort core.EventSort) string {
	switch sort {
	case core.EventSortObjectID:
		return "object_id"
	case core.EventSortNamespace:
		return "namespace"
	case core.EventSortName:
		return "name"
	case core.EventSortKind:
		return "kind"
	case core.EventSortCreatedAt:
		return "created_at"
	case core.EventSortMessage:
		return "message"
	default:
		return "id"
	}
}
