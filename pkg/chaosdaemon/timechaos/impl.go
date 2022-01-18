package timechaos

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/ChaosErr"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"syscall"
)

func (timeChaos *TimeChaos) Update(pid tasks.PID) error {
	uIDs := timeChaos.manager.GetTasksUIDByPID(pid)

	if process, err := timeChaos.Sum(pid, uIDs); err != nil {
		if errors.Cause(err) == ChaosErr.NotFound("PID") {
			return err
		}
		return errors.Wrapf(err, "unknown recovering in the taskMap, uIDs: %v", uIDs)
	} else {
		err := process.GroupInject(pid)
		if err != nil {
			return errors.Wrapf(err, "group inject existing process PID : %d", pid)
		}
	}
	return nil
}

func (timeChaos *TimeChaos) Apply(uid tasks.UID, t Task) error {
	timeChaos.Lock(strconv.Itoa(t.targetPID))
	// ignore err here because it will only return ErrNoSuchLock
	defer timeChaos.Unlock(strconv.Itoa(t.targetPID))

	err := timeChaos.manager.AddTask(uid, t.targetPID)
	if err != nil {
		return errors.Wrapf(ChaosErr.ErrNotImplemented, err.Error())
	}

	err = timeChaos.Update(t.targetPID)
	if err == nil {
		return nil
	}
	if errors.Cause(err) == ChaosErr.NotFound("PID") {
		task, err := t.timeSkewInfo.Fork()
		if err != nil {
			return errors.Wrapf(err, "fork time skew : %v", t.timeSkewInfo)
		}

		timeChaos.processMap[t.targetPID] = Process{timeSkew: task}
		err = timeChaos.Update(t.targetPID)
		if err != nil {
			return errors.Wrapf(err, "update new process")
		}
		return nil
	}
	return errors.Wrapf(err, "update old process")
}

func (timeChaos *TimeChaos) Recover(uid tasks.UID) error {
	task, ok := timeChaos.taskMap[uid]
	if !ok {
		return errors.Wrapf(ChaosErr.NotFound("UID"), "recovering UID : %v", uid)
	}

	timeChaos.Lock(strconv.Itoa(task.targetPID))
	defer timeChaos.Unlock(strconv.Itoa(task.targetPID))

	err := timeChaos.manager.RecoverTask(uid)
	if err != nil {
		timeChaos.logger.Error(err, "failed to recover task")
		return nil
	}

	uIDs := timeChaos.manager.GetTasksUIDByPID(task.targetPID)
	if len(uIDs) == 0 {
		if process, ok := timeChaos.processMap[task.targetPID]; ok {
			err := process.GroupRecovery(task.targetPID)
			if err != nil {
				p, errf := os.FindProcess(task.targetPID)
				if errf != nil {
					timeChaos.logger.Error(err, "can not find process")
					return nil
				}
				errs := p.Signal(syscall.Signal(0))
				if errs != nil {
					timeChaos.logger.Error(err, "can not check process with signal")
					return nil
				}
				return errors.Wrapf(err, "error & find process success")
			}
		}
		timeChaos.logger.Error(ChaosErr.NotFound("process"), "recovering task")
		return nil
	}

	err = timeChaos.Update(task.targetPID)
	if err != nil {
		return errors.Wrapf(err, "update new process")
	}
	return nil
}
