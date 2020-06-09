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

	// List all experiments
	if len(args) == 0 {
		if err := e.db.Find(&archives).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	} else {
		if err := e.db.Where(query, args).Find(&archives).Error; err != nil &&
			!gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	}

	return archives, nil
}

func (e *experimentStore) ListMeta(_ context.Context, kind, ns, name string) ([]*core.ArchiveExperimentMeta, error) {
	archives := make([]*core.ArchiveExperimentMeta, 0)
	query, args := constructQueryArgs(kind, ns, name)

	// List all experiment metas
	if len(args) == 0 {
		if err := e.db.Table("archive_experiments").Find(&archives).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	} else {
		if err := e.db.Table("archive_experiments").Where(query, args).
			Find(&archives).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	}

	return archives, nil
}

func (e *experimentStore) Report(ctx context.Context, kind, namespace, name string) (*core.ArchiveExperiment, error) {
	archive := new(core.ArchiveExperiment)

	if err := e.db.Where(
		"namespace = ? and name = ? and kind = ?", namespace, name, kind).
		First(archive).Error; err != nil {
		return nil, err
	}

	return archive, nil
}

// TODO: implement the left core.EventStore interfaces
func (e *experimentStore) Find(context.Context, int64) (*core.ArchiveExperiment, error) {
	return nil, nil
}

func (e *experimentStore) Create(_ context.Context, archive *core.ArchiveExperiment) error {
	return e.db.Create(archive).Error
}

func (e *experimentStore) Delete(context.Context, *core.ArchiveExperiment) error { return nil }

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
