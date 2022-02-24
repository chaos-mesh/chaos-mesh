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

package time

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
	"github.com/pkg/errors"
	"sync"
)

// clockGettimeSkewFakeImage is the filename of fake image after compiling
const clockGettimeSkewFakeImage = "fake_clock_gettime.o"

// clockGettime is the target function would be replaced
const clockGettime = "clock_gettime"

// These three consts corresponding to the three extern variables in the fake_clock_gettime.c
const (
	externVarClockIdsMask = "CLOCK_IDS_MASK"
	externVarTvSecDelta   = "TV_SEC_DELTA"
	externVarTvNsecDelta  = "TV_NSEC_DELTA"
)

// timeofdaySkewFakeImage is the filename of fake image after compiling
const timeOfDaySkewFakeImage = "fake_gettimeofday.o"

// getTimeOfDay is the target function would be replaced
const getTimeOfDay = "gettimeofday"

type Config struct {
	deltaSeconds     int64
	deltaNanoSeconds int64
	clockIDsMask     uint64
}

func (c *Config) DeepCopy() tasks.Object {
	return &Config{
		c.deltaSeconds,
		c.deltaNanoSeconds,
		c.clockIDsMask,
	}
}

func (c *Config) Add(a tasks.Addable) error {
	A, OK := a.(*Config)
	if OK {
		c.deltaSeconds += A.deltaSeconds
		c.deltaNanoSeconds += A.deltaNanoSeconds
		c.clockIDsMask |= A.clockIDsMask
		return nil
	}
	return errors.Wrapf(tasks.ErrCanNotAdd, "expect type : *time.Config, got : %T", a)
}

func (c *Config) New(values interface{}) (tasks.Injectable, error) {
	clockGetTimeImage, err := LoadFakeImageFromEmbedFs(clockGettimeSkewFakeImage, clockGettime)
	if err != nil {
		return nil, errors.Wrap(err, "load fake image")
	}

	getTimeOfDayimage, err := LoadFakeImageFromEmbedFs(timeOfDaySkewFakeImage, getTimeOfDay)
	if err != nil {
		return nil, errors.Wrap(err, "load fake image")
	}

	return &Skew{
		SkewConfig:   *c.DeepCopy().(*Config),
		clockGetTime: clockGetTimeImage,
		getTimeOfDay: getTimeOfDayimage,
	}, nil
}

func (c *Config) Assign(injectable tasks.Injectable) error {
	I, OK := injectable.(*Skew)
	if OK {
		I.SkewConfig = *c
		return nil
	}
	return errors.Wrapf(tasks.ErrCanNotAssign, "expect type : *time.Skew, got : %T", injectable)
}

type Skew struct {
	SkewConfig   Config
	clockGetTime *FakeImage
	getTimeOfDay *FakeImage

	locker sync.Mutex
	logger           logr.Logger
}

func (s *Skew) Fork() (tasks.ChaosOnProcessGroup, error) {
	// TODO : to KEAO can I share FakeImage between threads?
	injectable, err := s.SkewConfig.New(nil)
	if err != nil {
		return nil, err
	}
	return injectable.(*Skew), nil
}

func (s *Skew) Assign(injectable tasks.Injectable) error {
	I, OK := injectable.(*Skew)
	if OK {
		I.SkewConfig = s.SkewConfig
		return nil
	}
	return errors.Wrapf(tasks.ErrCanNotAssign, "expect type : *time.Skew, got : %T", injectable)
}

func (s *Skew) Inject(pid tasks.PID) error {
	s.locker.Lock()
	defer s.locker.Unlock()
	err := s.clockGetTime.AttachToProcess(pid, map[string]uint64{
		externVarClockIdsMask: s.SkewConfig.clockIDsMask,
		externVarTvSecDelta:   uint64(s.SkewConfig.deltaSeconds),
		externVarTvNsecDelta:  uint64(s.SkewConfig.deltaNanoSeconds),
	})
	if err != nil {
		return err
	}

	err := s.getTimeOfDay.AttachToProcess(pid, map[string]uint64{
		externVarTvSecDelta:  uint64(s.SkewConfig.deltaSeconds),
		externVarTvNsecDelta: uint64(s.SkewConfig.deltaNanoSeconds),
	})
	if err != nil {
		return err
	}
	panic("implement me")
}

func (s *Skew) Recover(pid tasks.PID) error {
	s.locker.Lock()
	defer s.locker.Unlock()
	//TODO implement me
	panic("implement me")
}
