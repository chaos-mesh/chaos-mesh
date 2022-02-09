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

var ErrUpdateTaskWithPIDChanges = errors.New("update task with PID changes")
var ErrTaskMapNotInit = errors.New("TaskMap not init")

// TaskManager provides some basic function on Tasks.
// If developers wants to use SumTask , they must implement Addable for the Task.
type TaskManager struct {
	TaskMap map[UID]Task
}

func NewTaskManager() TaskManager {
	return TaskManager{make(map[UID]Task)}
}

// Task defines a composite of flexible config with an immutable target.
// We use PID to stand by the target.
type Task struct {
	main PID
	data interface{}
}

func NewTask(main PID, data interface{}) Task {
	return Task{
		main,
		data,
	}
}

func (m TaskManager) AddTask(id UID, task Task) error {
	if m.TaskMap == nil {
		return ErrTaskMapNotInit
	}
	if _, ok := m.TaskMap[id]; ok {
		return errors.Wrapf(chaoserr.ErrDuplicateEntity, "uid: %s, task: %v", id, task)
	}
	m.TaskMap[id] = task
	return nil
}

func (m TaskManager) UpdateTask(id UID, task Task) (Task, error) {
	if m.TaskMap == nil {
		return Task{}, ErrTaskMapNotInit
	}
	taskOld, ok := m.TaskMap[id]
	if !ok {
		return Task{}, errors.Wrapf(chaoserr.NotFound("UID"), "uid: %s, task: %v", id, task)
	}
	if taskOld.main != task.main {
		return Task{}, errors.Wrapf(ErrUpdateTaskWithPIDChanges, "uid: %s, task: %v", id, task)
	}
	m.TaskMap[id] = task
	return taskOld, nil
}

func (m TaskManager) RecoverTask(id UID) error {
	if m.TaskMap == nil {
		return ErrTaskMapNotInit
	}
	_, ok := m.TaskMap[id]
	if !ok {
		return errors.Wrapf(chaoserr.NotFound("TASK"), "UID : %v", id)
	}
	delete(m.TaskMap, id)
	return nil
}

func (m TaskManager) GetWithUID(id UID) (interface{}, error) {
	t, ok := m.TaskMap[id]
	if !ok {
		return Task{}, chaoserr.NotFound("UID")
	}
	return t.data, nil
}

func (m TaskManager) GetWithPID(id PID) []UID {
	uIds := make([]UID, 0)
	for uid, task := range m.TaskMap {
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

// SumTask will add sum all the tasks that tasks pid is equal to task with input uid.
// If developers want to use it with type T, they must implement Addable for *T.
func (m TaskManager) SumTask(uid UID) (Task, error) {
	if m.TaskMap == nil {
		return Task{}, ErrTaskMapNotInit
	}
	task, ok := m.TaskMap[uid]
	if !ok {
		return Task{}, chaoserr.NotFound("UID")
	}
	uids := m.GetWithPID(task.main)

	for _, uidTemp := range uids {
		if uid == uidTemp {
			continue
		}
		taskTemp, ok := m.TaskMap[uidTemp]
		if !ok {
			return Task{}, chaoserr.NotFound("TASK")
		}
		AddableData, ok := task.data.(Addable)
		if !ok {
			return Task{}, errors.Wrapf(chaoserr.NotImplemented("Addable"), "task.Data")
		}
		AddableTempData, ok := taskTemp.data.(Addable)
		if !ok {
			return Task{}, errors.Wrapf(chaoserr.NotImplemented("Addable"), "taskTemp.Data")
		}
		err := AddableData.Add(AddableTempData)
		if err != nil {
			return Task{}, err
		}
	}
	return task, nil
}
