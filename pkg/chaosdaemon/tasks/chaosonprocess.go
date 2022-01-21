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

	"github.com/chaos-mesh/chaos-mesh/pkg/ChaosErr"
)

var ErrCanNotAdd = errors.New("can not add")
var ErrCanNotAssign = errors.New("can not assign")

// NewChaosOnProcess init ChaosOnProcess with values can not be Assigned.
type NewChaosOnProcess interface {
	New(immutableValues interface{}) (ChaosOnProcess, error)
}

type AssignChaosOnProcess interface {
	Assign(ChaosOnProcess) error
}

type TaskToProcess interface {
	Addable
	NewChaosOnProcess
	AssignChaosOnProcess
}

type ChaosCanRecover interface {
	Recover(pid PID) error
}

type ChaosOnProcess interface {
	Inject(pid PID) error
}

// ChaosOnProcessManager is a Manager for chaos tasks base on their
// target that marked with a PID and implement with ChaosOnProcess.
// Since we need to :
//	Sum tasks on same PID to one task.
// 	Create working chaos on process called ChaosOnProcess by task info.
// 	Assign the summed task info to the ChaosOnProcess.
// If developers wants to use functions in ChaosOnProcessManager ,
// their imported Task need to implement interface TaskToProcess.
// If developers wants to totally recover task successfully
// when no task applied on the ChaosOnProcess.
// This ChaosOnProcess must implement ChaosCanRecover.
// If not implement ,
// the Recover function will return a ErrNotImplement("ChaosCanRecover") error.
type ChaosOnProcessManager struct {
	taskManager TaskManager
	processMap  map[PID]ChaosOnProcess

	logger logr.Logger
}

func NewChaosOnProcessManager(logger logr.Logger) ChaosOnProcessManager {
	return ChaosOnProcessManager{
		NewTaskManager(),
		make(map[PID]ChaosOnProcess),
		logger,
	}
}

func (cm ChaosOnProcessManager) CopyTaskManager() TaskManager {
	tm := NewTaskManager()
	for uid, task := range cm.taskManager.TaskMap {
		tm.TaskMap[uid] = task
	}
	return tm
}

func (cm ChaosOnProcessManager) CopyProcessMap() map[PID]ChaosOnProcess {
	pm := make(map[PID]ChaosOnProcess)
	for pid, chaosOnProcess := range cm.processMap {
		cm.processMap[pid] = chaosOnProcess
	}
	return pm
}

func (cm ChaosOnProcessManager) GetWithUID(id UID) (interface{}, error) {
	return cm.taskManager.GetWithUID(id)
}

func (cm ChaosOnProcessManager) GetWithPID(pid PID) (ChaosOnProcess, error) {
	p, ok := cm.processMap[pid]
	if !ok {
		return nil, ChaosErr.NotFound("PID")
	}
	return p, nil
}

func (cm ChaosOnProcessManager) GetUIDsWithPID(pid PID) []UID {
	return cm.taskManager.GetWithPID(pid)
}

// Update the task config with a same UID and PID .
// If it comes a UID not injected , Update will return ChaosErr.NotFound("UID").
// If it comes the import PID of task do not equal to the last one,
// Update will return ErrUpdateTaskWithPIDChanges.
func (cm ChaosOnProcessManager) Update(uid UID, pid PID, config TaskToProcess) error {
	oldTask, err := cm.taskManager.UpdateTask(uid, NewTask(pid, config))
	if err != nil {
		return err
	}
	err = cm.commit(uid, pid)
	if err != nil {
		_, _ = cm.taskManager.UpdateTask(uid, oldTask)
		return err
	}
	return nil
}

// Create the first task on a process,
// the New function of TaskToProcess:NewChaosOnProcess will only be used here.
// immutableValues is only the import parameter of New function in TaskToProcess:NewChaosOnProcess.
// If it comes a PID are already be injected ,
// Create will return ChaosErr.ErrDuplicateEntity.
func (cm ChaosOnProcessManager) Create(uid UID, pid PID, config TaskToProcess, immutableValues interface{}) error {
	if _, ok := cm.processMap[pid]; ok {
		return errors.Wrapf(ChaosErr.ErrDuplicateEntity, "create")
	}

	err := cm.taskManager.AddTask(uid, NewTask(pid, config))
	if err != nil {
		return err
	}

	processTask, err := config.New(immutableValues)
	if err != nil {
		_ = cm.taskManager.RecoverTask(uid)
		return errors.Wrapf(err, "fork time skew : %v", config)
	}

	cm.processMap[pid] = processTask
	err = cm.commit(uid, pid)
	if err != nil {
		_ = cm.taskManager.RecoverTask(uid)
		delete(cm.processMap, pid)
		return errors.Wrapf(err, "update new process")
	}
	return nil
}

// Apply the task when the target pid of task is already be Created.
// If it comes a UID injected , Apply will return ChaosErr.ErrDuplicateEntity.
// If the Process has not been Created , Apply will return ChaosErr.NotFound("PID").
func (cm ChaosOnProcessManager) Apply(uid UID, pid PID, config TaskToProcess) error {
	err := cm.taskManager.AddTask(uid, NewTask(pid, config))
	if err != nil {
		return err
	}
	err = cm.commit(uid, pid)
	if err != nil {
		_ = cm.taskManager.RecoverTask(uid)
		return err
	}
	return nil
}

// Recover the task, if there is no task on PID or recovering the last task on PID.
// Recover in ChaosCanRecover will run, if runs failed it will just return the error.
// If Recover will fail , but developer wants to clear it run : cm.ClearProcessChaos(pid, true).
// If PID is already recovered successfully, Recover will return ChaosErr.NotFound("PID").
// If UID is not Applied or Created or the target PID of UID is not the import pid,
// Recover will return ChaosErr.NotFound("UID").
func (cm ChaosOnProcessManager) Recover(uid UID, pid PID) error {
	uIDs := cm.taskManager.GetWithPID(pid)
	if len(uIDs) == 0 {
		return ChaosErr.NotFound("PID")
	}
	if len(uIDs) == 1 {
		if uIDs[0] != uid {
			return ChaosErr.NotFound("UID")
		}
		err := cm.ClearProcessChaos(pid, false)
		if err != nil {
			return err
		}
		err = cm.taskManager.RecoverTask(uid)
		if err != nil {
			cm.logger.Error(err, "recover task with error")
		}
		return nil
	}

	err := cm.taskManager.RecoverTask(uid)
	if err != nil {
		cm.logger.Error(err, "recover task with error")
		return nil
	}

	err = cm.commit(uIDs[0], pid)
	if err != nil {
		return errors.Wrapf(err, "update new process")
	}
	return nil
}

func (cm ChaosOnProcessManager) commit(uid UID, pid PID) error {
	task, err := cm.taskManager.SumTask(uid)
	if err != nil {
		return errors.Wrapf(err, "unknown recovering in the taskMap, UID: %v", uid)
	}
	process, ok := cm.processMap[pid]
	if !ok {
		return errors.Wrapf(ChaosErr.NotFound("PID"), "PID : %d", pid)
	}
	tasker, ok := task.Data.(TaskToProcess)
	if !ok {
		return errors.New("task.Data here must implement TaskToProcess")
	}
	_ = tasker.Assign(process)
	if err != nil {
		return err
	}
	err = process.Inject(pid)
	if err != nil {
		return errors.Wrapf(err, "inject existing process PID : %d", pid)
	}
	return nil
}

func (cm ChaosOnProcessManager) ClearProcessChaos(pid PID, ignoreRecoverErr bool) error {
	if process, ok := cm.processMap[pid]; ok {
		pRecover, ok := process.(ChaosCanRecover)
		if !ok {
			return errors.Wrapf(ChaosErr.NotImplemented("ChaosCanRecover"), "process")
		}
		err := pRecover.Recover(pid)
		if err != nil && !ignoreRecoverErr {
			return errors.Wrapf(err, "recover chaos")
		}
		delete(cm.processMap, pid)
		return nil
	}
	return errors.Wrapf(ChaosErr.NotFound("PID"), "recovering task")
}
