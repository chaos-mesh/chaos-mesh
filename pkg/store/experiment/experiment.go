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

package experiment

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/store/dbstore"
)

var log = ctrl.Log.WithName("store/experiment")

// NewStore returns a new ExperimentStore.
func NewStore(db *dbstore.DB) core.ExperimentStore {
	db.AutoMigrate(&core.Experiment{})

	return &experimentStore{db}
}

// DeleteIncompleteExperiments call core.ExperimentStore.DeleteIncompleteExperiments to deletes all incomplete experiments.
func DeleteIncompleteExperiments(es core.ExperimentStore, _ core.EventStore) {
	if err := es.DeleteIncompleteExperiments(context.Background()); err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Error(err, "failed to delete all incomplete experiments")
	}
}

type experimentStore struct {
	db *dbstore.DB
}

// ListMeta implements the core.ExperimentStore.ListMeta method.
func (e *experimentStore) ListMeta(_ context.Context, kind, namespace, name string, archived bool) ([]*core.ExperimentMeta, error) {
	db := e.db.Table("experiments")
	experiments := make([]*core.ExperimentMeta, 0)
	query, args := constructQueryArgs(kind, namespace, name, "")

	if err := db.Where(query, args).Where(query, args).Where("archived = ?", archived).Find(&experiments).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return experiments, nil
}

// FindByUID implements the core.ExperimentStore.FindByUID method.
func (e *experimentStore) FindByUID(_ context.Context, uid string) (*core.Experiment, error) {
	experiment := new(core.Experiment)

	if err := e.db.Where("uid = ?", uid).First(experiment).Error; err != nil {
		return nil, err
	}

	return experiment, nil
}

// FindMetaByUID implements the core.ExperimentStore.FindMetaByUID method.
func (e *experimentStore) FindMetaByUID(_ context.Context, uid string) (*core.ExperimentMeta, error) {
	db := e.db.Table("experiments")
	experiment := new(core.ExperimentMeta)

	if err := db.Where("uid = ?", uid).First(experiment).Error; err != nil {
		return nil, err
	}

	return experiment, nil
}

// Set implements the core.ExperimentStore.Set method.
func (e *experimentStore) Set(_ context.Context, experiment *core.Experiment) error {
	return e.db.Model(core.Experiment{}).Save(experiment).Error
}

// Archive implements the core.ExperimentStore.Archive method.
func (e *experimentStore) Archive(_ context.Context, ns, name string) error {
	if err := e.db.Model(core.Experiment{}).
		Where("namespace = ? AND name = ? AND archived = ?", ns, name, false).
		Updates(map[string]interface{}{"archived": true, "finish_time": time.Now()}).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	return nil
}

// Delete deletes the experiment from the datastore.
func (e *experimentStore) Delete(_ context.Context, exp *core.Experiment) error {
	err := e.db.Table("experiments").Unscoped().Delete(*exp).Error
	return err
}

// DeleteByFinishTime deletes experiments whose time difference is greater than the given time from FinishTime.
func (e *experimentStore) DeleteByFinishTime(_ context.Context, ttl time.Duration) error {
	experiments, err := e.ListMeta(context.Background(), "", "", "", true)
	if err != nil {
		return err
	}

	nowTime := time.Now()
	for _, exp := range experiments {
		if exp.FinishTime.Add(ttl).Before(nowTime) {
			if err := e.db.Table("experiments").Unscoped().Delete(*exp).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteByUIDs deletes archives by the uid list.
func (e *experimentStore) DeleteByUIDs(_ context.Context, uids []string) error {
	return e.db.Table("experiments").Where("uid IN (?)", uids).Unscoped().Delete(core.Experiment{}).Error
}

// DeleteIncompleteExperiments implements the core.ExperimentStore.DeleteIncompleteExperiments method.
func (e *experimentStore) DeleteIncompleteExperiments(_ context.Context) error {
	return e.db.Where("finish_time IS NULL").Unscoped().Delete(core.Experiment{}).Error
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
