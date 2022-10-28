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
	"strconv"

	"github.com/go-logr/logr"

	"github.com/chaos-mesh/chaos-mesh/pkg/cerr"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

var ErrNotTypeSysID = cerr.NotType[SysPID]()
var ErrNotFoundSysID = cerr.NotFoundType[SysPID]()

type SysPID uint32

func (s SysPID) ToID() string {
	return strconv.FormatUint(uint64(s), 10)
}

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
	LeaderProcess ChaosOnProcessGroup
	childMap      map[IsID]ChaosOnProcessGroup
	Logger        logr.Logger
}

func NewProcessGroupHandler(logger logr.Logger, leader ChaosOnProcessGroup) ProcessGroupHandler {
	return ProcessGroupHandler{
		LeaderProcess: leader,
		childMap:      make(map[IsID]ChaosOnProcessGroup),
		Logger:        logr.New(logger.GetSink()),
	}
}

// Inject try to inject the leader process and then try to inject child process.
// If something wrong in injecting a child process, Inject will just log error & continue.
func (gp *ProcessGroupHandler) Inject(pid IsID) error {
	sysPID, ok := pid.(SysPID)
	if !ok {
		return ErrNotTypeSysID.WrapInput(pid).Err()
	}

	err := gp.LeaderProcess.Inject(sysPID)
	if err != nil {
		return cerr.FromErr(err).Wrapf("inject leader process: %v", sysPID).Err()
	}

	childPIDs, err := util.GetChildProcesses(uint32(sysPID), gp.Logger)
	if err != nil {
		return cerr.NotFound("child process").WrapErr(err).Err()
	}

	for _, childPID := range childPIDs {
		childSysPID := SysPID(childPID)
		if childProcessChaos, ok := gp.childMap[childSysPID]; ok {
			err := gp.LeaderProcess.Assign(childProcessChaos)
			if err != nil {
				gp.Logger.Error(err, "failed to assign old child process")
				continue
			}
			err = childProcessChaos.Inject(childSysPID)
			if err != nil {
				gp.Logger.Error(err, "failed to inject old child process")
			}
		} else {
			childProcessChaos, err := gp.LeaderProcess.Fork()
			if err != nil {
				gp.Logger.Error(err, "failed to create child process")
				continue
			}
			err = childProcessChaos.Inject(childSysPID)
			if err != nil {
				gp.Logger.Error(err, "failed to inject new child process")
				continue
			}
			gp.childMap[childSysPID] = childProcessChaos
		}
	}
	return nil
}

// Recover try to recover the leader process and then try to recover child process.
func (gp *ProcessGroupHandler) Recover(pid IsID) error {
	_, ok := pid.(SysPID)
	if !ok {
		return ErrNotTypeSysID.WrapInput(pid).Err()
	}
	err := gp.LeaderProcess.Recover(pid)
	if err != nil {
		return cerr.FromErr(err).Wrapf("recovery leader process : %v", pid).Err()
	}

	for childID, group := range gp.childMap {
		childSysPID, ok := childID.(SysPID)
		if !ok {
			gp.Logger.Error(cerr.NotType[SysPID]().WrapInput(childID).Err(),
				"failed to recover old child process")
		}

		err := group.Recover(childSysPID)
		if err != nil {
			gp.Logger.Error(err, "failed to recover old child process")
		}
	}
	return nil
}
