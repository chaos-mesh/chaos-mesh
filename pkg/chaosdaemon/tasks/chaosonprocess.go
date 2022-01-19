package tasks

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/ChaosErr"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"os"
	"syscall"
)

type ChaosOnProcess interface {
	Inject(pid PID) error
	Recover(pid PID) error
}

type ChaosOnProcessManager struct {
	taskManager TaskManager
	processMap  map[PID]ChaosOnProcess

	logger logr.Logger
}

func (cm ChaosOnProcessManager) Update(uid UID, pid PID) error {
	task, err := cm.taskManager.SumTask(uid)
	if err != nil {
		return errors.Wrapf(err, "unknown recovering in the taskMap, UID: %v", uid)
	}
	process, ok := cm.processMap[pid]
	if !ok {
		return errors.Wrapf(ChaosErr.NotFound("PID"), "PID : %d", pid)
	}
	_ = task.Data.Assign(process)
	if err != nil {
		return err
	}
	err = process.Inject(pid)
	if err != nil {
		return errors.Wrapf(err, "inject existing process PID : %d", pid)
	}
	return nil
}

func (cm ChaosOnProcessManager) Apply(uid UID, pid PID, config TaskInner) error {
	err := cm.taskManager.AddTask(uid, GetTask(pid, config))
	if err != nil {
		return errors.Wrapf(ChaosErr.ErrNotImplemented, err.Error())
	}
	err = cm.Update(uid, pid)
	if err == nil {
		return nil
	}
	if errors.Cause(err) == ChaosErr.NotFound("PID") {
		processTask, err := config.New(cm.logger)
		if err != nil {
			return errors.Wrapf(err, "fork time skew : %v", config)
		}

		cm.processMap[pid] = processTask
		err = cm.Update(uid, pid)
		if err != nil {
			return errors.Wrapf(err, "update new process")
		}
		return nil
	}
	return errors.Wrapf(err, "update old process")
}

func (cm ChaosOnProcessManager) Recover(uid UID, pid PID) error {
	_, err := cm.taskManager.RecoverTask(uid)
	if err != nil {
		cm.logger.Error(err, "failed to recover task")
		return nil
	}

	uIDs := cm.taskManager.GetTasksUIDByPID(pid)
	if len(uIDs) == 0 {
		if process, ok := cm.processMap[pid]; ok {
			err := process.Recover(pid)
			if err != nil {
				// Check if pid is not exist.If true , return nil.
				p, errf := os.FindProcess(pid)
				if errf != nil {
					cm.logger.Error(err, "can not find process")
					return nil
				}
				errs := p.Signal(syscall.Signal(0))
				if errs != nil {
					cm.logger.Error(err, "can not check process with signal")
					return nil
				}
				return errors.Wrapf(err, "error & find process success")
			}
		}
		cm.logger.Error(ChaosErr.NotFound("process"), "recovering task")
		return nil
	}

	err = cm.Update(uid, pid)
	if err != nil {
		return errors.Wrapf(err, "update new process")
	}
	return nil
}
