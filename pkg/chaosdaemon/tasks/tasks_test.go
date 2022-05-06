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
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaos-mesh/pkg/cerr"
)

type FakeConfig struct {
	i int
}

func (f *FakeConfig) Merge(a Mergeable) error {
	A, OK := a.(*FakeConfig)
	if OK {
		f.i += A.i
		return nil
	}
	return cerr.NotType[*FakeConfig]().WrapInput(a).Err()
}

func (f *FakeConfig) DeepCopy() Object {
	temp := *f
	return &temp
}

func (f *FakeConfig) Assign(c Injectable) error {
	C, OK := c.(*FakeChaos)
	if OK {
		C.C.i = f.i
		return nil
	}
	return cerr.NotType[*FakeConfig]().WrapInput(c).Err()
}

func (f *FakeConfig) New(immutableValues interface{}) (Injectable, error) {
	temp := immutableValues.(*FakeChaos)
	f.Assign(temp)
	return temp, nil
}

type FakeChaos struct {
	C              FakeConfig
	ErrWhenRecover bool
	ErrWhenInject  bool
	logger         logr.Logger
}

func (f *FakeChaos) Inject(pid IsID) error {
	if f.ErrWhenInject {
		return cerr.NotImpl[Injectable]().Err()
	}
	return nil
}

func (f *FakeChaos) Recover(pid IsID) error {
	if f.ErrWhenRecover {
		return cerr.NotImpl[Recoverable]().Err()
	}
	return nil
}

func TestTasksManager(t *testing.T) {
	var log logr.Logger

	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)

	m := NewTaskManager(log)

	chaos := FakeChaos{
		ErrWhenRecover: false,
		ErrWhenInject:  false,
		logger:         log,
	}
	task1 := FakeConfig{i: 1}
	uid1 := "1"
	err = m.Create(uid1, SysPID(1), &task1, &chaos)
	chaosInterface, err := m.GetTaskWithPID(SysPID(1))
	assert.NoError(t, err)
	chaoso := chaosInterface.(*FakeChaos)
	assert.Equal(t, chaoso.C, task1)
	assert.Equal(t, chaoso, &chaos)

	task2 := FakeConfig{i: 1}
	uid2 := "2"
	err = m.Apply(uid2, SysPID(1), &task2)
	chaosInterface, err = m.GetTaskWithPID(SysPID(1))
	assert.NoError(t, err)
	chaoso = chaosInterface.(*FakeChaos)
	assert.Equal(t, chaoso.C, FakeConfig{i: 2})
	assert.Equal(t, chaos.C, FakeConfig{i: 2})

	assert.Equal(t, task1, FakeConfig{1})
	assert.Equal(t, task2, FakeConfig{1})
}

func TestTasksManagerError(t *testing.T) {
	var log logr.Logger

	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)

	m := NewTaskManager(log)

	chaos := FakeChaos{
		ErrWhenRecover: false,
		ErrWhenInject:  false,
		logger:         log,
	}
	task1 := FakeConfig{i: 1}
	uid1 := "1"
	err = m.Create(uid1, SysPID(1), &task1, &chaos)
	assert.NoError(t, err)
	err = m.Apply(uid1, SysPID(1), &task1)
	assert.Equal(t, errors.Cause(err), cerr.ErrDuplicateEntity)
	err = m.Recover(uid1, SysPID(1))
	assert.NoError(t, err)
	err = m.Recover(uid1, SysPID(1))
	assert.Equal(t, errors.Cause(err), ErrNotFoundTaskID.Err())

	chaos.ErrWhenInject = true
	tasks2 := FakeConfig{i: 1}
	err = m.Create(uid1, SysPID(1), &tasks2, &chaos)
	assert.Equal(t, errors.Cause(err).Error(), cerr.NotImpl[Injectable]().Err().Error())
	_, err = m.GetConfigWithUID(uid1)
	assert.Equal(t, errors.Cause(err), ErrNotFoundTaskID.Err())

	chaos.ErrWhenInject = false
	chaos.ErrWhenRecover = true
	tasks3 := FakeConfig{i: 1}
	err = m.Create(uid1, SysPID(1), &tasks3, &chaos)
	assert.NoError(t, err)
	err = m.Recover(uid1, SysPID(1))
	assert.Equal(t, errors.Cause(err).Error(), cerr.NotImpl[Recoverable]().Err().Error())
	p, err := m.GetTaskWithPID(SysPID(1))
	inner := p.(*FakeChaos)
	inner.ErrWhenRecover = false
	err = m.Recover(uid1, SysPID(1))
	assert.NoError(t, err)
}
