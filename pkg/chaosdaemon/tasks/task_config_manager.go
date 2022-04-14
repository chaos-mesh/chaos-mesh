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

	"github.com/chaos-mesh/chaos-mesh/pkg/cerr"
)

var ErrNotFoundID = cerr.NotFound("ID")
var ErrNotFoundTaskID = cerr.NotFound("TaskID")
var ErrNotFoundTypeTaskConfig = cerr.NotFoundType[TaskConfig]()

var ErrDiffID = cerr.FromErr(errors.New("different IsID"))

var ErrTaskConfigMapNotInit = cerr.NotInit[map[TaskID]TaskConfig]().WrapName("TaskConfigMap").Err()

type IsID interface {
	ToID() string
}

// Object ensure the outer config change will not change
// the data inside the TaskManager.
type Object interface {
	DeepCopy() Object
}

// Mergeable introduces the data gathering ability.
type Mergeable interface {
	Merge(a Mergeable) error
}

// TaskConfig defines a composite of flexible config with an immutable target.
// TaskConfig.Id is the ID of task.
// TaskConfig.Data is the config provided by developer.
type TaskConfig struct {
	Id   IsID
	Data Object
}

func NewTaskConfig(id IsID, data Object) TaskConfig {
	return TaskConfig{
		id,
		data.DeepCopy(),
	}
}

type TaskID = string

// TaskConfigManager provides some basic methods on TaskConfig.
// If developers wants to use MergeTaskConfig, they must implement Mergeable for the TaskConfig.
type TaskConfigManager struct {
	TaskConfigMap map[TaskID]TaskConfig
}

func NewTaskConfigManager() TaskConfigManager {
	return TaskConfigManager{make(map[TaskID]TaskConfig)}
}

func (m TaskConfigManager) AddTaskConfig(id TaskID, task TaskConfig) error {
	if m.TaskConfigMap == nil {
		return ErrTaskConfigMapNotInit
	}
	if _, ok := m.TaskConfigMap[id]; ok {
		return errors.Wrapf(cerr.ErrDuplicateEntity, "uid: %s, task: %v", id, task)
	}
	m.TaskConfigMap[id] = task
	return nil
}

func (m TaskConfigManager) UpdateTaskConfig(id TaskID, task TaskConfig) (TaskConfig, error) {
	if m.TaskConfigMap == nil {
		return TaskConfig{}, ErrTaskConfigMapNotInit
	}
	taskOld, ok := m.TaskConfigMap[id]
	if !ok {
		return TaskConfig{}, ErrNotFoundTaskID.WrapInput(id).WrapInput(task).Err()
	}
	if taskOld.Id != task.Id {
		return TaskConfig{}, ErrDiffID.Wrapf("expect: %v, input: %v", taskOld.Id, task.Id).Err()
	}
	m.TaskConfigMap[id] = task
	return taskOld, nil
}

// DeleteTaskConfig Delete task inside the TaskConfigManager
func (m TaskConfigManager) DeleteTaskConfig(id TaskID) error {
	if m.TaskConfigMap == nil {
		return ErrTaskConfigMapNotInit
	}
	_, ok := m.TaskConfigMap[id]
	if !ok {
		return ErrNotFoundTypeTaskConfig.WrapInput(id).Err()
	}
	delete(m.TaskConfigMap, id)
	return nil
}

func (m TaskConfigManager) GetConfigWithUID(id TaskID) (TaskConfig, error) {
	t, ok := m.TaskConfigMap[id]
	if !ok {
		return TaskConfig{}, ErrNotFoundTaskID.WrapInput(id).Err()
	}
	return t, nil
}

func (m TaskConfigManager) GetUIDsWithPID(id IsID) []TaskID {
	uIds := make([]TaskID, 0)
	for uid, task := range m.TaskConfigMap {
		if task.Id == id {
			uIds = append(uIds, uid)
		}
	}
	return uIds
}

func (m TaskConfigManager) CheckTask(uid TaskID, pid IsID) error {
	t, ok := m.TaskConfigMap[uid]
	if !ok {
		return ErrNotFoundTaskID.WrapInput(uid).Err()
	}
	if t.Id != pid {
		return ErrDiffID.Wrapf("expect: %v, input: %v", t.Id, pid).Err()
	}
	return nil
}

// MergeTaskConfig will sum the TaskConfig with a same TaskConfig.Id.
// If developers want to use it with type T, they must implement Mergeable for *T.
// IMPORTANT: Just here , we do not assume A.Merge(B) == B.Merge(A).
// What MergeTaskConfig do : A := new(TaskConfig), A.Merge(B).Merge(C).Merge(D)... , A marked as uid.
func (m TaskConfigManager) MergeTaskConfig(uid TaskID) (TaskConfig, error) {
	if m.TaskConfigMap == nil {
		return TaskConfig{}, ErrTaskConfigMapNotInit
	}
	taskRaw, ok := m.TaskConfigMap[uid]
	if !ok {
		return TaskConfig{}, ErrNotFoundTaskID.WrapInput(uid).Err()
	}

	task := TaskConfig{
		Id:   taskRaw.Id,
		Data: taskRaw.Data.DeepCopy(),
	}
	uids := m.GetUIDsWithPID(task.Id)

	for _, uidTemp := range uids {
		if uid == uidTemp {
			continue
		}
		taskTemp, ok := m.TaskConfigMap[uidTemp]
		if !ok {
			return TaskConfig{}, ErrNotFoundTypeTaskConfig.WrapInput(uidTemp).Err()
		}
		AddableData, ok := task.Data.(Mergeable)
		if !ok {
			return TaskConfig{}, cerr.NotImpl[Mergeable]().WrapInput(task.Data).Err()
		}
		AddableTempData, ok := taskTemp.Data.(Mergeable)
		if !ok {
			return TaskConfig{}, cerr.NotImpl[Mergeable]().WrapInput(taskTemp.Data).Err()
		}
		err := AddableData.Merge(AddableTempData)
		if err != nil {
			return TaskConfig{}, err
		}
	}
	return task, nil
}
