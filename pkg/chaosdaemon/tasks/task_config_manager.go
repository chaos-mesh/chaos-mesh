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

package tasks

import (
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaoserr"
)

type UID = string
type PID = int

var ErrUpdateTaskConfigWithPIDChanges = errors.New("update task config with PID changes")
var ErrTaskConfigMapNotInit = errors.New("TaskConfigMap not init")

// TaskConfig defines a composite of flexible config with an immutable target.
// TaskConfig.main is the ID of task.
// TaskConfig.data is the config provided by developer.
type TaskConfig struct {
	main PID
	data interface{}
}

func NewTaskConfig(main PID, data interface{}) TaskConfig {
	return TaskConfig{
		main,
		data,
	}
}

// TaskConfigManager provides some basic methods on TaskConfig.
// If developers wants to use SumTask, they must implement Addable for the TaskConfig.
type TaskConfigManager struct {
	TaskConfigMap map[UID]TaskConfig
}

func NewTaskConfigManager() TaskConfigManager {
	return TaskConfigManager{make(map[UID]TaskConfig)}
}

func (m TaskConfigManager) AddTask(id UID, task TaskConfig) error {
	if m.TaskConfigMap == nil {
		return ErrTaskConfigMapNotInit
	}
	if _, ok := m.TaskConfigMap[id]; ok {
		return errors.Wrapf(chaoserr.ErrDuplicateEntity, "uid: %s, task: %v", id, task)
	}
	m.TaskConfigMap[id] = task
	return nil
}

func (m TaskConfigManager) UpdateTask(id UID, task TaskConfig) (TaskConfig, error) {
	if m.TaskConfigMap == nil {
		return TaskConfig{}, ErrTaskConfigMapNotInit
	}
	taskOld, ok := m.TaskConfigMap[id]
	if !ok {
		return TaskConfig{}, errors.Wrapf(chaoserr.NotFound("UID"), "uid: %s, task: %v", id, task)
	}
	if taskOld.main != task.main {
		return TaskConfig{}, errors.Wrapf(ErrUpdateTaskConfigWithPIDChanges, "uid: %s, task: %v", id, task)
	}
	m.TaskConfigMap[id] = task
	return taskOld, nil
}

// RecoverTask Delete task inside the TaskConfigManager
func (m TaskConfigManager) RecoverTask(id UID) error {
	if m.TaskConfigMap == nil {
		return ErrTaskConfigMapNotInit
	}
	_, ok := m.TaskConfigMap[id]
	if !ok {
		return errors.Wrapf(chaoserr.NotFound("Task Config"), "UID : %v", id)
	}
	delete(m.TaskConfigMap, id)
	return nil
}

func (m TaskConfigManager) GetConfigWithUID(id UID) (interface{}, error) {
	t, ok := m.TaskConfigMap[id]
	if !ok {
		return TaskConfig{}, chaoserr.NotFound("UID")
	}
	return t.data, nil
}

func (m TaskConfigManager) GetUIDsWithPID(id PID) []UID {
	uIds := make([]UID, 0)
	for uid, task := range m.TaskConfigMap {
		if task.main == id {
			uIds = append(uIds, uid)
		}
	}
	return uIds
}

// Addable introduces the data gathering ability.
type Addable interface {
	Add(a Addable) error
}

// SumTask will sum the TaskConfig with a same TaskConfig.main.
// If developers want to use it with type T, they must implement Addable for *T.
// IMPORTANT: Just here , we do not assume A.Add(B) == B.Add(A).
// What SumTask do : A.Add(B).Add(C).Add(D)... , A marked as uid.
func (m TaskConfigManager) SumTask(uid UID) (TaskConfig, error) {
	if m.TaskConfigMap == nil {
		return TaskConfig{}, ErrTaskConfigMapNotInit
	}
	task, ok := m.TaskConfigMap[uid]
	if !ok {
		return TaskConfig{}, chaoserr.NotFound("UID")
	}
	uids := m.GetUIDsWithPID(task.main)

	for _, uidTemp := range uids {
		if uid == uidTemp {
			continue
		}
		taskTemp, ok := m.TaskConfigMap[uidTemp]
		if !ok {
			return TaskConfig{}, chaoserr.NotFound("Task Config")
		}
		AddableData, ok := task.data.(Addable)
		if !ok {
			return TaskConfig{}, errors.Wrapf(chaoserr.NotImplemented("Addable"), "task.Data")
		}
		AddableTempData, ok := taskTemp.data.(Addable)
		if !ok {
			return TaskConfig{}, errors.Wrapf(chaoserr.NotImplemented("Addable"), "taskTemp.Data")
		}
		err := AddableData.Add(AddableTempData)
		if err != nil {
			return TaskConfig{}, err
		}
	}
	return task, nil
}
