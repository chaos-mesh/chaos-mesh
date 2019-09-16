// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// ManagerBaseInterface defines some base functions to manager the Runners.
type ManagerBaseInterface interface {
	AddRunner(runner *Runner) error
	DeleteRunner(key string) error
	UpdateRunner(runner *Runner) error
	IsExist(key string) bool
}

// ManagerBase is the ManagerBaseInterface implementation.
type ManagerBase struct {
	cronEngine *cron.Cron
	runners    sync.Map
	sync.Mutex
}

func NewManagerBase(engine *cron.Cron) *ManagerBase {
	return &ManagerBase{
		cronEngine: engine,
		runners:    sync.Map{},
	}
}

func (m *ManagerBase) AddRunner(runner *Runner) error {
	m.Lock()
	defer m.Unlock()

	if err := runner.Validate(); err != nil {
		return err
	}

	return m.addRunnerAction(runner)
}

func (m *ManagerBase) DeleteRunner(key string) error {
	m.Lock()
	defer m.Unlock()

	return m.deleteRunnerAction(key)
}

func (m *ManagerBase) UpdateRunner(runner *Runner) error {
	m.Lock()
	defer m.Unlock()

	if err := runner.Validate(); err != nil {
		return err
	}

	val, ok := m.runners.Load(runner.Name)
	if !ok {
		return fmt.Errorf("runner %s not found", runner.Name)
	}
	oldRunner := val.(*Runner)

	if runner.Rule == oldRunner.Rule {
		return nil
	}

	return m.updateRunnerAction(runner)
}

func (m *ManagerBase) IsExist(key string) bool {
	_, ok := m.runners.Load(key)
	return ok
}

func (m *ManagerBase) addRunnerAction(runner *Runner) error {
	entryID, err := m.cronEngine.AddJob(runner.Name, runner.Job)
	if err != nil {
		return err
	}

	runner.EntryID = int(entryID)
	m.runners.Store(runner.Name, runner)

	return nil
}

func (m *ManagerBase) deleteRunnerAction(key string) error {
	runner, ok := m.runners.Load(key)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("runner %s not found", key))
		return nil
	}

	m.cronEngine.Remove(cron.EntryID(runner.(*Runner).EntryID))
	m.runners.Delete(key)
	return nil
}

func (m *ManagerBase) updateRunnerAction(newRunner *Runner) error {
	m.deleteRunnerAction(newRunner.Name)

	return m.addRunnerAction(newRunner)
}
