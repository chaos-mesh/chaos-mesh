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

	"github.com/chaos-mesh/chaos-mesh/pkg/cerr"
)

// Injectable stand for the base behavior of task : inject a process with IsID.
type Injectable interface {
	Inject(pid IsID) error
}

// Recoverable introduce the task recovering ability.
// Used in Recover.
type Recoverable interface {
	Recover(pid IsID) error
}

// Creator init an Injectable with values.
// We use it in a case that TaskConfig.Data init an Injectable task here.
type Creator interface {
	New(values interface{}) (Injectable, error)
}

// Assign change some of an Injectable task with its own values.
// We use it in a case that we use TaskConfig.Data
// to update an Injectable task.
type Assign interface {
	Assign(Injectable) error
}

// TaskExecutor indicate that the type can be used
// for execute task here as a task config.
// Mergeable means we can sum many task config in to one for apply.
// Creator means we can use the task config to create a running task
// which can Inject on IsID.
// Assign means we can use the task config to update an existing running task.
type TaskExecutor interface {
	Object
	Mergeable
	Creator
	Assign
}

// TaskManager is a Manager for chaos tasks.
// A task base on a target marked with its IsID.
// We assume task should implement Injectable.
// We use TaskConfig.Data which implement TaskExecutor to:
//
//	Sum task configs on same IsID into one.
//	Create task.
//	Assign or update task.
//
// SO if developers wants to use functions in TaskManager ,
// their imported TaskConfig need to implement interface TaskExecutor.
// If developers wants to recover task successfully,
// the task must implement Recoverable.
// If not implement ,
// the Recover function will return a ErrNotImplement("Recoverable") error.
// IMPORTANT: We assume task config obey that TaskConfig.Data A,B. A.Merge(B)
// is approximately equal to B.Merge(A)
type TaskManager struct {
	taskConfigManager TaskConfigManager
	taskMap           map[IsID]Injectable

	logger logr.Logger
}

func NewTaskManager(logger logr.Logger) TaskManager {
	return TaskManager{
		NewTaskConfigManager(),
		make(map[IsID]Injectable),
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

func (cm TaskManager) CopyTaskMap() map[IsID]Injectable {
	pm := make(map[IsID]Injectable)
	for pid, chaosOnProcess := range cm.taskMap {
		cm.taskMap[pid] = chaosOnProcess
	}
	return pm
}

func (cm TaskManager) GetConfigWithUID(id TaskID) (TaskConfig, error) {
	return cm.taskConfigManager.GetConfigWithUID(id)
}

func (cm TaskManager) GetTaskWithPID(pid IsID) (Injectable, error) {
	p, ok := cm.taskMap[pid]
	if !ok {
		return nil, ErrNotFoundID.WrapInput(pid).Err()
	}
	return p, nil
}

func (cm TaskManager) GetUIDsWithPID(pid IsID) []TaskID {
	return cm.taskConfigManager.GetUIDsWithPID(pid)
}

func (cm TaskManager) CheckTasks(uid TaskID, pid IsID) error {
	config, err := cm.GetConfigWithUID(uid)
	if err != nil {
		return err
	}
	if config.Id != pid {
		return ErrDiffID.Wrapf("expected: %v, input: %v", config.Id, pid).Err()
	}
	return nil
}

// Create the first task,
// the New function of TaskExecutor:Creator will only be used here.
// values is only the import parameter of New function in TaskExecutor:Creator.
// If it comes a task are already be injected on the IsID,
// Create will return ChaosErr.ErrDuplicateEntity.
func (cm TaskManager) Create(uid TaskID, pid IsID, config TaskExecutor, values interface{}) error {
	if _, ok := cm.taskMap[pid]; ok {
		return errors.Wrapf(cerr.ErrDuplicateEntity, "create")
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
// If it comes a TaskID injected , Apply will return ChaosErr.ErrDuplicateEntity.
// If the Process has not been Created , Apply will return ChaosErr.NotFound("IsID").
func (cm TaskManager) Apply(uid TaskID, pid IsID, config TaskExecutor) error {
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

// Update the task with a same TaskID, IsID and new task config.
// If it comes a TaskID not injected , Update will return ChaosErr.NotFound("TaskID").
// If it comes the import IsID of task do not equal to the last one,
// Update will return ErrDiffID.
func (cm TaskManager) Update(uid TaskID, pid IsID, config TaskExecutor) error {
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

// Recover the task when there is no task config on IsID or
// recovering the task with last task config on IsID.
// Recover in Recoverable will be used here,
// if it runs failed it will just return the error.
// If Recover is failed but developer wants to clear it,
// just run : TaskManager.ClearTask(pid, true).
// If IsID is already recovered successfully, Recover will return ChaosErr.NotFound("IsID").
// If TaskID is not Applied or Created or the target IsID of TaskID is not the import pid,
// Recover will return ChaosErr.NotFound("TaskID").
func (cm TaskManager) Recover(uid TaskID, pid IsID) error {
	uIDs := cm.taskConfigManager.GetUIDsWithPID(pid)
	if len(uIDs) == 0 {
		return ErrNotFoundTaskID.WrapInput(pid).Err()
	}
	if len(uIDs) == 1 {
		if uIDs[0] != uid {
			return ErrNotFoundTaskID.WrapInput(uid).Err()
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

	uIDs = cm.taskConfigManager.GetUIDsWithPID(pid)

	err = cm.commit(uIDs[0], pid)
	if err != nil {
		return errors.Wrapf(err, "update new task")
	}
	return nil
}

func (cm TaskManager) commit(uid TaskID, pid IsID) error {
	task, err := cm.taskConfigManager.MergeTaskConfig(uid)
	if err != nil {
		return errors.Wrapf(err, "unknown recovering in the taskConfigManager, TaskID: %v", uid)
	}
	process, ok := cm.taskMap[pid]
	if !ok {
		return ErrNotFoundID.WrapInput(pid).Err()
	}
	tasker, ok := task.Data.(TaskExecutor)
	if !ok {
		return errors.New("task.Data here must implement TaskExecutor")
	}
	err = tasker.Assign(process)
	if err != nil {
		return err
	}
	err = process.Inject(pid)
	if err != nil {
		return errors.Wrapf(err, "inject existing process IsID : %v", pid)
	}
	return nil
}

// ClearTask clear the task totally.
// IMPORTANT: Developer should only use this function
// when want to force clear task with ignoreRecoverErr==true.
func (cm TaskManager) ClearTask(pid IsID, ignoreRecoverErr bool) error {
	if process, ok := cm.taskMap[pid]; ok {
		pRecover, ok := process.(Recoverable)
		if !ok {
			return cerr.NotImpl[Recoverable]().WrapInput(process).Err()
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
	return ErrNotFoundID.WrapInput(pid).Err()
}
