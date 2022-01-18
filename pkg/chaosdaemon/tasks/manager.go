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
	TaskIDMap map[UID]PID
}

func Get() TaskManager {
	return TaskManager{
		make(map[UID]PID),
	}
}

func (m TaskManager) AddTask(id UID, task PID) error {
	if m.TaskIDMap == nil {
		return errors.New("map not init")
	}
	if _, ok := m.TaskIDMap[id]; ok {
		return errors.Wrapf(ChaosErr.ErrDuplicateEntity, "uid: %s, task: %v", id, task)
	}
	m.TaskIDMap[id] = task
	return nil
}

func (m TaskManager) RecoverTask(id UID) error {
	if _, ok := m.TaskIDMap[id]; !ok {
		return errors.Wrapf(ChaosErr.NotFound("UID"), "UID : %v", id)
	}
	delete(m.TaskIDMap, id)
	return nil
}

func (m TaskManager) GetTasksUIDByPID(id PID) []UID {
	uIds := make([]UID, 0)
	for uid, pid := range m.TaskIDMap {
		if pid == id {
			uIds = append(uIds, uid)
		}
	}
	return uIds
}
