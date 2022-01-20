package tasks

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

type ChaosOnGroupProcess interface {
	Fork() (ChaosOnGroupProcess, error)
	AssignChaosOnProcess

	ChaosOnProcess
	ChaosCanRecover
}

// GroupProcessHandler implements Group PID inject & recover.
type GroupProcessHandler struct {
	Main     ChaosOnGroupProcess
	childMap map[PID]ChaosOnGroupProcess
	logger   logr.Logger
}

func NewGroupProcessHandler(logger logr.Logger, main ChaosOnGroupProcess) GroupProcessHandler {
	return GroupProcessHandler{
		Main:     main,
		childMap: make(map[PID]ChaosOnGroupProcess),
		logger:   logger,
	}
}

func (gp *GroupProcessHandler) Inject(pid PID) error {
	childPids, err := util.GetChildProcesses(uint32(pid))
	if err != nil {
		gp.logger.Error(err, "failed to get child process")
	}

	err = gp.Main.Inject(pid)
	if err != nil {
		return errors.Wrapf(err, "inject main process : %d", pid)
	}

	for _, childPid := range childPids {
		childPID := PID(childPid)
		if childProcessChaos, ok := gp.childMap[childPID]; ok {
			err := gp.Main.Assign(childProcessChaos)
			if err != nil {
				return err
			}
			err = childProcessChaos.Inject(childPID)
			if err != nil {
				gp.logger.Error(err, "failed to inject old child process")
			}
		} else {
			childProcessChaos, err := gp.Main.Fork()
			if err != nil {
				gp.logger.Error(err, "failed to create child process")
				return nil
			}
			err = childProcessChaos.Inject(pid)
			if err != nil {
				gp.logger.Error(err, "failed to inject new child process")
				return nil
			}
			gp.childMap[childPID] = childProcessChaos
		}
	}
	return nil
}

func (gp *GroupProcessHandler) Recover(pid PID) error {
	childPids, err := util.GetChildProcesses(uint32(pid))
	if err != nil {
		gp.logger.Error(err, "failed to get child process")
	}

	err = gp.Main.Recover(pid)
	if err != nil {
		return errors.Wrapf(err, "recovery main process : %d", pid)
	}

	for _, childPid := range childPids {
		childPID := PID(childPid)
		if childProcessChaos, ok := gp.childMap[childPID]; ok {
			err := childProcessChaos.Recover(childPID)
			if err != nil {
				gp.logger.Error(err, "failed to recover old child process")
			}
		}
	}
	return nil
}
