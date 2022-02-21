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
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaoserr"
)

var ErrCanNotAdd = errors.New("can not add")
var ErrCanNotAssign = errors.New("can not assign")

// Injectable stand for the base behavior of task : inject a process with PID.
type Injectable interface {
	Inject(pid PID) error
}

// Recoverable introduce the task recovering ability.
// Used in Recover.
type Recoverable interface {
	Recover(pid PID) error
}

// Creator init an Injectable with values.
// We use it in a case that TaskConfig.data init an Injectable task here.
type Creator interface {
	New(values interface{}) (Injectable, error)
}

// Assign change some of an Injectable task with its own values.
// We use it in a case that we use TaskConfig.data
// to update an Injectable task.
type Assign interface {
	Assign(Injectable) error
}

// TaskExecutor indicate that the type can be used
// for execute task here as a task config.
// Addable means we can sum many task config in to one for apply.
// Creator means we can use the task config to create a running task
// which can Inject on PID.
// Assign means we can use the task config to update an existing running task.
type TaskExecutor interface {
	Object
	Addable
	Creator
	Assign
}

// TaskManager is a Manager for chaos tasks.
// A task base on a target marked with its PID.
// We assume task should implement Injectable.
// We use TaskConfig.data which implement TaskExecutor to:
//	Sum task configs on same PID into one.
// 	Create task.
// 	Assign or update task.
// SO if developers wants to use functions in TaskManager ,
// their imported TaskConfig need to implement interface TaskExecutor.
// If developers wants to recover task successfully,
// the task must implement Recoverable.
// If not implement ,
// the Recover function will return a ErrNotImplement("Recoverable") error.
// IMPORTANT: We assume task config obey that TaskConfig.data A,B. A.Add(B)
// is approximately equal to B.Add(A)
type TaskManager struct {
	taskConfigManager TaskConfigManager
	taskMap           map[PID]Injectable

	logger logr.Logger
}

func NewTaskManager(logger logr.Logger) TaskManager {
	return TaskManager{
		NewTaskConfigManager(),
		make(map[PID]Injectable),
		logger,
	}
}

func (cm TaskManager) CopyTaskConfigManager() TaskConfigManager {
	tm := NewTaskConfigManager()
	for uid, task := range cm.taskConfigManager.TaskConfigMap {
		tm.TaskConfigMap[uid] = task
	}
	return tm
}

func (cm TaskManager) CopyTaskMap() map[PID]Injectable {
	pm := make(map[PID]Injectable)
	for pid, chaosOnProcess := range cm.taskMap {
		cm.taskMap[pid] = chaosOnProcess
	}
	return pm
}

func (cm TaskManager) GetConfigWithUID(id UID) (interface{}, error) {
	return cm.taskConfigManager.GetConfigWithUID(id)
}

func (cm TaskManager) GetTaskWithPID(pid PID) (Injectable, error) {
	p, ok := cm.taskMap[pid]
	if !ok {
		return nil, ErrPIDNotFound
	}
	return p, nil
}

func (cm TaskManager) GetUIDsWithPID(pid PID) []UID {
	return cm.taskConfigManager.GetUIDsWithPID(pid)
}

// Update the task with a same UID, PID and new task config.
// If it comes a UID not injected , Update will return ChaosErr.NotFound("UID").
// If it comes the import PID of task do not equal to the last one,
// Update will return ErrUpdateTaskConfigWithPIDChanges.
func (cm TaskManager) Update(uid UID, pid PID, config TaskExecutor) error {
	oldTask, err := cm.taskConfigManager.UpdateTaskConfig(uid, NewTaskConfig(pid, config))
	if err != nil {
		return err
	}
	err = cm.commit(uid, pid)
	if err != nil {
		_, _ = cm.taskConfigManager.UpdateTaskConfig(uid, oldTask)
		return err
	}
	return nil
}

// Create the first task,
// the New function of TaskExecutor:Creator will only be used here.
// values is only the import parameter of New function in TaskExecutor:Creator.
// If it comes a task are already be injected on the PID,
// Create will return ChaosErr.ErrDuplicateEntity.
func (cm TaskManager) Create(uid UID, pid PID, config TaskExecutor, values interface{}) error {
	if _, ok := cm.taskMap[pid]; ok {
		return errors.Wrapf(chaoserr.ErrDuplicateEntity, "create")
	}

	err := cm.taskConfigManager.AddTaskConfig(uid, NewTaskConfig(pid, config))
	if err != nil {
		return err
	}

	processTask, err := config.New(values)
	if err != nil {
		_ = cm.taskConfigManager.DeleteTaskConfig(uid)
		return errors.Wrapf(err, "New task: %v", config)
	}

	cm.taskMap[pid] = processTask
	err = cm.commit(uid, pid)
	if err != nil {
		_ = cm.taskConfigManager.DeleteTaskConfig(uid)
		delete(cm.taskMap, pid)
		return errors.Wrapf(err, "update new task")
	}
	return nil
}

// Apply the task when the target pid of task is already be Created.
// If it comes a UID injected , Apply will return ChaosErr.ErrDuplicateEntity.
// If the Process has not been Created , Apply will return ChaosErr.NotFound("PID").
func (cm TaskManager) Apply(uid UID, pid PID, config TaskExecutor) error {
	err := cm.taskConfigManager.AddTaskConfig(uid, NewTaskConfig(pid, config))
	if err != nil {
		return err
	}
	err = cm.commit(uid, pid)
	if err != nil {
		_ = cm.taskConfigManager.DeleteTaskConfig(uid)
		return err
	}
	return nil
}

// Recover the task when there is no task config on PID or
// recovering the task with last task config on PID.
// Recover in Recoverable will be used here,
// if it runs failed it will just return the error.
// If Recover is failed but developer wants to clear it,
// just run : TaskManager.ClearTask(pid, true).
// If PID is already recovered successfully, Recover will return ChaosErr.NotFound("PID").
// If UID is not Applied or Created or the target PID of UID is not the import pid,
// Recover will return ChaosErr.NotFound("UID").
func (cm TaskManager) Recover(uid UID, pid PID) error {
	uIDs := cm.taskConfigManager.GetUIDsWithPID(pid)
	if len(uIDs) == 0 {
		return ErrPIDNotFound
	}
	if len(uIDs) == 1 {
		if uIDs[0] != uid {
			return ErrUIDNotFound
		}
		err := cm.ClearTask(pid, false)
		if err != nil {
			return err
		}
		err = cm.taskConfigManager.DeleteTaskConfig(uid)
		if err != nil {
			cm.logger.Error(err, "recover task with error")
		}
		return nil
	}

	err := cm.taskConfigManager.DeleteTaskConfig(uid)
	if err != nil {
		cm.logger.Error(err, "recover task with error")
		return nil
	}

	err = cm.commit(uIDs[0], pid)
	if err != nil {
		return errors.Wrapf(err, "update new task")
	}
	return nil
}

func (cm TaskManager) commit(uid UID, pid PID) error {
	task, err := cm.taskConfigManager.SumTaskConfig(uid)
	if err != nil {
		return errors.Wrapf(err, "unknown recovering in the taskConfigManager, UID: %v", uid)
	}
	process, ok := cm.taskMap[pid]
	if !ok {
		return errors.Wrapf(ErrPIDNotFound, "PID : %d", pid)
	}
	tasker, ok := task.data.(TaskExecutor)
	if !ok {
		return errors.New("task.Data here must implement TaskExecutor")
	}
	err = tasker.Assign(process)
	if err != nil {
		return err
	}
	err = process.Inject(pid)
	if err != nil {
		return errors.Wrapf(err, "inject existing process PID : %d", pid)
	}
	return nil
}

// ClearTask clear the task totally.
// IMPORTANT: Developer should only use this function
// when want to force clear task with ignoreRecoverErr==true.
func (cm TaskManager) ClearTask(pid PID, ignoreRecoverErr bool) error {
	if process, ok := cm.taskMap[pid]; ok {
		pRecover, ok := process.(Recoverable)
		if !ok {
			return errors.Wrapf(chaoserr.NotImplemented("Recoverable"), "process")
		}
		err := pRecover.Recover(pid)
		if err != nil {
			if ignoreRecoverErr {
				cm.logger.Error(errors.Wrapf(err, "recover chaos"), "ERR IGNORED")
			} else {
				return errors.Wrapf(err, "recover chaos")
			}

		}
		delete(cm.taskMap, pid)
		return nil
	}
	return errors.Wrapf(ErrPIDNotFound, "recovering task")
}
