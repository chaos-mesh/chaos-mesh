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

package experiment

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

var log = utils.LogStore.WithName("experiments")

// NewStore returns a new ExperimentStore.
func NewStore(db *gorm.DB) core.ExperimentStore {
	db.AutoMigrate(&core.Experiment{})

	return &experimentStore{db}
}

type experimentStore struct {
	db *gorm.DB
}

func (e *experimentStore) ListByFilter(_ context.Context, filter core.Filter, archived bool) ([]*core.ExperimentMeta, error) {
	var (
		exps []*core.ExperimentMeta
	)

	query, args := filter.ConstructQueryArgs()
	statement := e.db.Table("experiments").Unscoped().Where(query, args...).Where("archived = ?", archived).Order("id desc")

	if err := statement.Find(&exps).Error; err != nil {
		return nil, err
	}

	return exps, nil
}

func (e *experimentStore) ListMeta(c context.Context, namespace, name, kind string, archived bool) ([]*core.ExperimentMeta, error) {
	return e.ListByFilter(c, core.Filter{
		Namespace: namespace,
		Name:      name,
		Kind:      kind,
	}, archived)
}

func (e *experimentStore) FindByUID(_ context.Context, uid string) (*core.Experiment, error) {
	experiment := new(core.Experiment)

	if err := e.db.Where("uid = ?", uid).First(experiment).Error; err != nil {
		return nil, err
	}

	return experiment, nil
}

func (e *experimentStore) Save(_ context.Context, experiment *core.Experiment) error {
	return e.db.Save(experiment).Error
}

func (e *experimentStore) Archive(_ context.Context, ns, name string) error {
	return e.db.
		Model(&core.Experiment{}).
		Where("namespace = ? AND name = ? AND archived = ?", ns, name, false).
		Updates(map[string]interface{}{"archived": true, "deleted_at": time.Now()}).Error
}

func (e *experimentStore) Delete(_ context.Context, exp *core.Experiment) error {
	return e.db.Unscoped().Delete(exp).Error
}
func (e *experimentStore) DeleteByUIDs(_ context.Context, uids []string) error {
	return e.db.Unscoped().Where("uid IN (?)", uids).Delete(&core.Experiment{}).Error
}

func (e *experimentStore) DeleteByDuration(_ context.Context, duration time.Duration) error {
	now := time.Now().UTC().Add(-duration).Format("2006-01-02 15:04:05")

	return e.db.Where("deleted_at <= ?", now).Delete(&core.Experiment{}).Error
}

func (e *experimentStore) DeleteIncompleteExperiments(_ context.Context) error {
	return e.db.Unscoped().Where("deleted_at IS NULL").Delete(&core.Experiment{}).Error
}

// DeleteIncompleteExperiments call core.ExperimentStore.DeleteIncompleteExperiments to deletes all incomplete experiments.
func DeleteIncompleteExperiments(es core.ExperimentStore, _ core.EventStore) {
	if err := es.DeleteIncompleteExperiments(context.Background()); err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Error(err, "failed to delete all incomplete experiments")
	}
}
