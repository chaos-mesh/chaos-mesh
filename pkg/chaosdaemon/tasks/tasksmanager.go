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
	"github.com/chaos-mesh/chaos-mesh/pkg/ChaosErr"
	"github.com/pkg/errors"
)

// A Manager for Chaos-Daemon.
// For example , Time-Chaos apply tasks on several processes.
// Every task have its own ID and process have PID.
// TaskManager provide some common function to solve the relationship
// of task and process

type UID = string
type PID = int

type TaskManager struct {
	TaskMap map[UID]Task
}

func NewTaskManager() TaskManager {
	return TaskManager{make(map[UID]Task)}
}

type Task struct {
	Main PID
	Data Addable
}

func GetTask(main PID, data Addable) Task {
	return Task{
		main,
		data,
	}
}

func (m TaskManager) AddTask(id UID, task Task) error {
	if m.TaskMap == nil {
		return errors.New("map not init")
	}
	if _, ok := m.TaskMap[id]; ok {
		return errors.Wrapf(ChaosErr.ErrDuplicateEntity, "uid: %s, task: %v", id, task)
	}
	m.TaskMap[id] = task
	return nil
}

func (m TaskManager) UpdateTask(id UID, task Task) error {
	if m.TaskMap == nil {
		return errors.New("map not init")
	}
	if _, ok := m.TaskMap[id]; !ok {
		return errors.Wrapf(ChaosErr.NotFound("UID"), "uid: %s, task: %v", id, task)
	}
	m.TaskMap[id] = task
	return nil
}

func (m TaskManager) RecoverTask(id UID) (Task, error) {
	task, ok := m.TaskMap[id]
	if !ok {
		return Task{}, errors.Wrapf(ChaosErr.NotFound("TASK"), "UID : %v", id)
	}
	delete(m.TaskMap, id)
	return task, nil
}

func (m TaskManager) SumTask(uid UID) (Task, error) {
	task, ok := m.TaskMap[uid]
	if !ok {
		return Task{}, ChaosErr.NotFound("UID")
	}
	uids := m.GetTasksUIDByPID(task.Main)

	for _, uidTemp := range uids {
		if uid == uidTemp {
			continue
		}
		taskTemp, ok := m.TaskMap[uidTemp]
		if !ok {
			return Task{}, ChaosErr.NotFound("TASK")
		}
		err := task.Data.Add(taskTemp.Data)
		if err != nil {
			return Task{}, err
		}
	}
	return task, nil
}

func (m TaskManager) GetTasksUIDByPID(id PID) []UID {
	uIds := make([]UID, 0)
	for uid, task := range m.TaskMap {
		if task.Main == id {
			uIds = append(uIds, uid)
		}
	}
	return uIds
}
