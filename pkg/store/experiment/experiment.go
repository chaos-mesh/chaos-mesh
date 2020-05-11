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

// TODO: implement core.EventStore interface
func (e *experimentStore) List(context.Context) ([]*core.ArchiveExperiment, error) { return nil, nil }
func (e *experimentStore) ListByKind(context.Context, string) ([]*core.ArchiveExperiment, error) {
	return nil, nil
}
func (e *experimentStore) Find(context.Context, int64) (*core.ArchiveExperiment, error) {
	return nil, nil
}
func (e *experimentStore) FindByName(context.Context, string, string) (*core.ArchiveExperiment, error) {
	return nil, nil
}
func (e *experimentStore) Create(context.Context, *core.ArchiveExperiment) error { return nil }
func (e *experimentStore) Update(context.Context, *core.ArchiveExperiment) error { return nil }
func (e *experimentStore) Delete(context.Context, *core.ArchiveExperiment) error { return nil }
