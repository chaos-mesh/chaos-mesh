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

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaoserr"
)

var ErrNotFoundChildProcess = chaoserr.NotFound("child process")

// ChaosOnProcessGroup is used for inject a chaos on a linux process group.
// Fork is used for create a new chaos on child process.
// Assign is used for update a chaos on child process.
type ChaosOnProcessGroup interface {
	Fork() (ChaosOnProcessGroup, error)
	Assign

	Injectable
	Recoverable
}

// ProcessGroupHandler implements injecting & recovering on a linux process group.
type ProcessGroupHandler struct {
	Main     ChaosOnProcessGroup
	childMap map[PID]ChaosOnProcessGroup
	logger   logr.Logger
}

func NewProcessGroupHandler(logger logr.Logger, main ChaosOnProcessGroup) ProcessGroupHandler {
	return ProcessGroupHandler{
		Main:     main,
		childMap: make(map[PID]ChaosOnProcessGroup),
		logger:   logger,
	}
}

// Inject try to inject the main process and then try to inject child process.
// If something wrong in injecting a child process, Inject will just log error & continue.
func (gp *ProcessGroupHandler) Inject(pid PID) error {
	err := gp.Main.Inject(pid)
	if err != nil {
		return errors.Wrapf(err, "inject main process : %d", pid)
	}

	childPids, err := util.GetChildProcesses(uint32(pid), gp.logger)
	if err != nil {
		return errors.Wrapf(ErrNotFoundChildProcess, "cause : %v", err)
	}

	for _, childPid := range childPids {
		childPID := PID(childPid)
		if childProcessChaos, ok := gp.childMap[childPID]; ok {
			err := gp.Main.Assign(childProcessChaos)
			if err != nil {
				gp.logger.Error(err, "failed to assign old child process")
				continue
			}
			err = childProcessChaos.Inject(childPID)
			if err != nil {
				gp.logger.Error(err, "failed to inject old child process")
			}
		} else {
			childProcessChaos, err := gp.Main.Fork()
			if err != nil {
				gp.logger.Error(err, "failed to create child process")
				continue
			}
			err = childProcessChaos.Inject(pid)
			if err != nil {
				gp.logger.Error(err, "failed to inject new child process")
				continue
			}
			gp.childMap[childPID] = childProcessChaos
		}
	}
	return nil
}

// Recover try to recover the main process and then try to recover child process.
func (gp *ProcessGroupHandler) Recover(pid PID) error {
	err := gp.Main.Recover(pid)
	if err != nil {
		return errors.Wrapf(err, "recovery main process : %d", pid)
	}

	childPids, err := util.GetChildProcesses(uint32(pid), gp.logger)
	if err != nil {
		return errors.Wrapf(ErrNotFoundChildProcess, "cause : %v", err)
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
