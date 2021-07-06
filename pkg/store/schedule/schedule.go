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

package schedule

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/store/dbstore"
)

var log = ctrl.Log.WithName("store/schedule")

// NewStore returns a new ScheduleStore.
func NewStore(db *dbstore.DB) core.ScheduleStore {
	db.AutoMigrate(&core.Schedule{})

	return &ScheduleStore{db}
}

// DeleteIncompleteSchedules call core.ScheduleStore.DeleteIncompleteSchedules to deletes all incomplete schedules.
func DeleteIncompleteSchedules(es core.ScheduleStore, _ core.EventStore) {
	if err := es.DeleteIncompleteSchedules(context.Background()); err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Error(err, "failed to delete all incomplete schedules")
	}
}

type ScheduleStore struct {
	db *dbstore.DB
}

// ListMeta implements the core.ScheduleStore.ListMeta method.
func (e *ScheduleStore) ListMeta(_ context.Context, namespace, name string, archived bool) ([]*core.ScheduleMeta, error) {
	db := e.db.Table("schedules")
	sches := make([]*core.ScheduleMeta, 0)
	query, args := constructQueryArgs("", namespace, name, "")

	if err := db.Where(query, args).Where(query, args).Where("archived = ?", archived).Find(&sches).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return sches, nil
}

// FindByUID implements the core.ScheduleStore.FindByUID method.
func (e *ScheduleStore) FindByUID(_ context.Context, uid string) (*core.Schedule, error) {
	sch := new(core.Schedule)

	if err := e.db.Where("uid = ?", uid).First(sch).Error; err != nil {
		return nil, err
	}

	return sch, nil
}

// FindMetaByUID implements the core.ScheduleStore.FindMetaByUID method.
func (e *ScheduleStore) FindMetaByUID(_ context.Context, uid string) (*core.ScheduleMeta, error) {
	db := e.db.Table("schedules")
	sch := new(core.ScheduleMeta)

	if err := db.Where("uid = ?", uid).First(sch).Error; err != nil {
		return nil, err
	}

	return sch, nil
}

// Set implements the core.ScheduleStore.Set method.
func (e *ScheduleStore) Set(_ context.Context, schedule *core.Schedule) error {
	return e.db.Model(core.Schedule{}).Save(schedule).Error
}

// Archive implements the core.ScheduleStore.Archive method.
func (e *ScheduleStore) Archive(_ context.Context, ns, name string) error {
	if err := e.db.Model(core.Schedule{}).
		Where("namespace = ? AND name = ? AND archived = ?", ns, name, false).
		Updates(map[string]interface{}{"archived": true, "finish_time": time.Now()}).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	return nil
}

// Delete deletes the experiment from the datastore.
func (e *ScheduleStore) Delete(_ context.Context, exp *core.Schedule) error {
	err := e.db.Table("schedules").Unscoped().Delete(*exp).Error
	return err
}

// DeleteByFinishTime deletes schedules whose time difference is greater than the given time from FinishTime.
func (e *ScheduleStore) DeleteByFinishTime(_ context.Context, ttl time.Duration) error {
	sches, err := e.ListMeta(context.Background(), "", "", true)
	if err != nil {
		return err
	}

	nowTime := time.Now()
	for _, sch := range sches {
		if sch.FinishTime.Add(ttl).Before(nowTime) {
			if err := e.db.Table("schedules").Unscoped().Delete(*sch).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteByUIDs deletes schedules by the uid list.
func (e *ScheduleStore) DeleteByUIDs(_ context.Context, uids []string) error {
	return e.db.Table("schedules").Where("uid IN (?)", uids).Unscoped().Delete(core.Schedule{}).Error
}

// DeleteIncompleteSchedules implements the core.ScheduleStore.DeleteIncompleteSchedules method.
func (e *ScheduleStore) DeleteIncompleteSchedules(_ context.Context) error {
	return e.db.Where("finish_time IS NULL").Unscoped().Delete(core.Schedule{}).Error
}

func constructQueryArgs(kind, ns, name, uid string) (string, []string) {
	query := ""
	args := make([]string, 0)

	if kind != "" {
		query += "kind = ?"
		args = append(args, kind)
	}

	if ns != "" {
		if len(args) > 0 {
			query += " AND namespace = ?"
		} else {
			query += "namespace = ?"
		}
		args = append(args, ns)
	}

	if name != "" {
		if len(args) > 0 {
			query += " AND name = ?"
		} else {
			query += "name = ?"
		}
		args = append(args, name)
	}

	if uid != "" {
		if len(args) > 0 {
			query += " AND uid = ?"
		} else {
			query += "uid = ?"
		}
		args = append(args, uid)
	}

	return query, args
}
