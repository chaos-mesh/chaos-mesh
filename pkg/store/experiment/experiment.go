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

package experiment

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/pingcap/chaos-mesh/pkg/core"
	"github.com/pingcap/chaos-mesh/pkg/store/dbstore"
)

// NewStore returns a new ExperimentStore.
func NewStore(db *dbstore.DB) core.ExperimentStore {
	db.AutoMigrate(&core.ArchiveExperiment{})

	return &experimentStore{db}
}

type experimentStore struct {
	db *dbstore.DB
}

func (e *experimentStore) List(_ context.Context, kind, ns, name string) ([]*core.ArchiveExperiment, error) {
	archives := make([]*core.ArchiveExperiment, 0)
	query, args := constructQueryArgs(kind, ns, name)

	db := e.db.Model(core.ArchiveExperiment{})
	if len(args) > 0 {
		db = db.Where(query, args)
	}

	// List all experiments
	if err := db.Where("archived = ?", true).Find(&archives).Error; err != nil &&
		!gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return archives, nil
}

func (e *experimentStore) ListMeta(_ context.Context, kind, ns, name string) ([]*core.ArchiveExperimentMeta, error) {
	archives := make([]*core.ArchiveExperimentMeta, 0)
	query, args := constructQueryArgs(kind, ns, name)

	db := e.db.Table("archive_experiments")
	if len(args) > 0 {
		db = db.Where(query, args)
	}

	if err := db.Where("archived = ?", true).
		Find(&archives).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return archives, nil
}

func (e *experimentStore) Archive(_ context.Context, ns, name string) error {
	if err := e.db.Model(core.ArchiveExperiment{}).
		Where("namespace = ? AND name = ? AND archived = ?", ns, name, false).
		Updates(map[string]interface{}{"archived": true, "finish_time": time.Now()}).Error; err != nil &&
		!gorm.IsRecordNotFoundError(err) {
		return err
	}
	return nil
}

func (e *experimentStore) FindByUID(_ context.Context, UID string) (*core.ArchiveExperiment, error) {
	expr := new(core.ArchiveExperiment)

	if err := e.db.Model(core.ArchiveExperiment{}).
		Where("uid = ?", UID).First(expr).Error; err != nil {
		return nil, err
	}

	return expr, nil
}

func (e *experimentStore) Set(_ context.Context, archive *core.ArchiveExperiment) error {
	return e.db.Model(core.ArchiveExperiment{}).Save(archive).Error
}

// TODO: implement the left core.EventStore interfaces
func (e *experimentStore) Find(context.Context, int64) (*core.ArchiveExperiment, error) {
	return nil, nil
}

func (e *experimentStore) Delete(context.Context, *core.ArchiveExperiment) error { return nil }

// DeleteByFinishTime deletes experiments whose time difference is greater than the given time from FinishTime.
func (e *experimentStore) DeleteByFinishTime(_ context.Context, ttl time.Duration) error {
	expList, err := e.List(context.Background(), "", "", "")
	if err != nil {
		return err
	}
	nowTime := time.Now()
	for _, exp := range expList {
		if exp.FinishTime.Add(ttl).Before(nowTime) {
			if err := e.db.Table("archive_experiments").Unscoped().Delete(*exp).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func constructQueryArgs(kind, ns, name string) (string, []string) {
	args := make([]string, 0)
	query := ""
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
	return query, args
}
