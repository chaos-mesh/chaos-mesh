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

package workflow

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/store/dbstore"
	"github.com/jinzhu/gorm"
	"time"
)

type WorkflowStore struct {
	db *dbstore.DB
}

func NewStore(db *dbstore.DB) *WorkflowStore {
	db.AutoMigrate(&core.WorkflowEntity{})
	return &WorkflowStore{db: db}
}

func (it *WorkflowStore) List(ctx context.Context, namespace, name string, archived bool) ([]*core.WorkflowEntity, error) {
	var entities []core.WorkflowEntity

	err := it.db.Find(&entities).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	var result []*core.WorkflowEntity
	for _, item := range entities {
		item := item
		result = append(result, &item)
	}
	return result, nil
}

func (it *WorkflowStore) ListMeta(ctx context.Context, namespace, name string, archived bool) ([]*core.WorkflowMeta, error) {
	entities, err := it.List(ctx, namespace, name, archived)
	if err != nil {
		return nil, err
	}
	var result []*core.WorkflowMeta
	for _, item := range entities {
		item := item
		result = append(result, &item.WorkflowMeta)
	}
	return result, nil
}

func (it *WorkflowStore) FindByID(ctx context.Context, id uint) (*core.WorkflowEntity, error) {
	result := new(core.WorkflowEntity)
	if err := it.db.Where(
		"id = ?", id).
		First(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (it *WorkflowStore) FindByUID(ctx context.Context, uid string) (*core.WorkflowEntity, error) {
	result := new(core.WorkflowEntity)
	if err := it.db.Where(
		"uid = ?", uid).
		First(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (it *WorkflowStore) FindMetaByUID(ctx context.Context, UID string) (*core.WorkflowMeta, error) {
	entity, err := it.FindByUID(ctx, UID)
	if err != nil {
		return nil, err
	}
	return &entity.WorkflowMeta, nil
}

func (it *WorkflowStore) Save(ctx context.Context, entity core.WorkflowEntity) error {
	return it.db.Model(core.WorkflowEntity{}).Save(entity).Error
}

func (it *WorkflowStore) DeleteByUID(ctx context.Context, uid string) error {
	return it.db.Where("object_id = ?", uid).Unscoped().
		Delete(core.Event{}).Error
}

func (it *WorkflowStore) DeleteByUIDs(ctx context.Context, uids []string) error {
	return it.db.Where("object_id IN (?)", uids).Unscoped().Delete(core.Event{}).Error
}

func (it *WorkflowStore) MarkAsArchived(ctx context.Context, namespace, name string) error {
	if err := it.db.Model(core.Schedule{}).
		Where("namespace = ? AND name = ? AND archived = ?", namespace, name, false).
		Updates(map[string]interface{}{"archived": true, "finish_time": time.Now()}).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}
	return nil
}

func (it *WorkflowStore) MarkAsArchivedWithUID(ctx context.Context, uid string) error {
	if err := it.db.Model(core.Schedule{}).
		Where("uid = ? AND archived = ?", uid, false).
		Updates(map[string]interface{}{"archived": true, "finish_time": time.Now()}).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}
	return nil
}
