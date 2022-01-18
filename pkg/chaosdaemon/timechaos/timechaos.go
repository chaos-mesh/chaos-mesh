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

package timechaos

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/ChaosErr"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
	"github.com/chaos-mesh/chaos-mesh/pkg/time"
	"github.com/go-logr/logr"
	"github.com/moby/locker"
	"github.com/pkg/errors"
)

type TimeChaos struct {
	manager    tasks.TaskManager
	taskMap    map[tasks.UID]Task
	processMap map[tasks.PID]Process

	logger logr.Logger
	locker.Locker
}

type Task struct {
	targetPID    tasks.PID
	timeSkewInfo time.TimeSkew
}

type Process struct {
	timeSkew *time.TimeSkew
	childMap map[tasks.PID]*time.TimeSkew
	logger   logr.Logger
}

func (p Process) GroupInject(pid tasks.PID) error {
	childPids, err := chaosdaemon.GetChildProcesses(uint32(pid))
	if err != nil {
		p.logger.Error(err, "failed to get child process")
	}

	err = p.timeSkew.Inject(pid)
	if err != nil {
		return errors.Wrapf(err, "inject main process : %d", pid)
	}

	for _, childPid := range childPids {
		childPID := tasks.PID(childPid)
		if childTimeSkew, ok := p.childMap[childPID]; ok {
			childTimeSkew.Assign(*p.timeSkew)
			err := childTimeSkew.Inject(childPID)
			if err != nil {
				p.logger.Error(err, "failed to inject old child process")
			}
		} else {
			childTimeSkew, err := p.timeSkew.Fork()
			if err != nil {
				p.logger.Error(err, "failed to fork child process")
				return nil
			}
			err = childTimeSkew.Inject(pid)
			if err != nil {
				p.logger.Error(err, "failed to inject new child process")
				return nil
			}
			p.childMap[childPID] = childTimeSkew
		}
	}
	return nil
}

func (p Process) GroupRecovery(pid tasks.PID) error {
	childPids, err := chaosdaemon.GetChildProcesses(uint32(pid))
	if err != nil {
		p.logger.Error(err, "failed to get child process")
	}

	err = p.timeSkew.Recover(pid)
	if err != nil {
		return errors.Wrapf(err, "recovery main process : %d", pid)
	}

	for _, childPid := range childPids {
		childPID := tasks.PID(childPid)
		if childTimeSkew, ok := p.childMap[childPID]; ok {
			err := childTimeSkew.Recover(childPID)
			if err != nil {
				p.logger.Error(err, "failed to recover old child process")
			}
		}
	}
	return nil
}

func (timeChaos *TimeChaos) Sum(pid tasks.PID, uIDs []tasks.UID) (*Process, error) {
	t := time.TimeSkew{}

	for _, uID := range uIDs {
		if task, ok := timeChaos.taskMap[uID]; ok {
			t.Add(task.timeSkewInfo)
		} else {
			return nil, errors.Wrapf(ChaosErr.NotFound("task UID"), "task UID : %s", uID)
		}
	}

	if process, ok := timeChaos.processMap[pid]; ok {
		process.timeSkew.Assign(t)
		return &process, nil
	}
	return nil, errors.Wrapf(ChaosErr.NotFound("PID"), "PID : %d", pid)
}

func (timeChaos *TimeChaos) Shutdown() error {
	return ChaosErr.ErrNotImplemented
}
