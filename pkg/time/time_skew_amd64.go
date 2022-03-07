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
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
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

func NewConfig(deltaSeconds int64, deltaNanoSeconds int64, clockIDsMask uint64) Config {
	return Config{
		deltaSeconds:     deltaSeconds,
		deltaNanoSeconds: deltaNanoSeconds,
		clockIDsMask:     clockIDsMask,
	}
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

type ConfigCreatorParas struct {
	Logger        logr.Logger
	Config        Config
	PodProcessMap *tasks.PodProcessMap
}

func (c *Config) New(values interface{}) (tasks.Injectable, error) {
	paras, ok := values.(ConfigCreatorParas)
	if !ok {
		return nil, errors.New("not ConfigCreatorParas")
	}

	skew, err := GetSkew()
	if err != nil {
		return nil, err
	}
	skew.SkewConfig = *paras.Config.DeepCopy().(*Config)

	newGroupProcessHandler :=
		tasks.NewProcessGroupHandler(paras.Logger, &skew)
	newPodHandler := tasks.NewPodHandler(paras.PodProcessMap,
		&newGroupProcessHandler, paras.Logger)
	return &newPodHandler, nil
}

func (c *Config) Assign(injectable tasks.Injectable) error {
	podHandler, ok := injectable.(*tasks.PodHandler)
	if !ok {
		return errors.New(fmt.Sprintf("type %t is not *tasks.PodHandler", injectable))
	}
	groupProcessHandler, ok := podHandler.Main.(*tasks.ProcessGroupHandler)
	if !ok {
		return errors.New(fmt.Sprintf("type %t is not *tasks.ProcessGroupHandler", podHandler.Main))
	}
	I, ok := groupProcessHandler.Main.(*Skew)
	if !ok {
		return errors.New(fmt.Sprintf("type %t is not *Skew", groupProcessHandler.Main))
	}

	I.SkewConfig = *c
	return nil
}

type Skew struct {
	SkewConfig   Config
	clockGetTime *FakeImage
	getTimeOfDay *FakeImage

	locker sync.Mutex
	logger           logr.Logger
}

func GetSkew() (Skew, error) {
	clockGetTimeImage, err := LoadFakeImageFromEmbedFs(clockGettimeSkewFakeImage, clockGettime)
	if err != nil {
		return Skew{}, errors.Wrap(err, "load fake image")
	}

	getTimeOfDayimage, err := LoadFakeImageFromEmbedFs(timeOfDaySkewFakeImage, getTimeOfDay)
	if err != nil {
		return Skew{}, errors.Wrap(err, "load fake image")
	}

	return Skew{
		SkewConfig:   Config{},
		clockGetTime: clockGetTimeImage,
		getTimeOfDay: getTimeOfDayimage,
		locker:       sync.Mutex{},
	}, nil
}

func (s *Skew) Fork() (tasks.ChaosOnProcessGroup, error) {
	// TODO : to KEAO can I share FakeImage between threads?
	skew, err := GetSkew()
	if err != nil {
		return nil, err
	}
	skew.SkewConfig = *s.SkewConfig.DeepCopy().(*Config)

	return &skew, nil
}

func (s *Skew) Assign(injectable tasks.Injectable) error {
	I, OK := injectable.(*Skew)
	if OK {
		I.SkewConfig = *s.SkewConfig.DeepCopy().(*Config)
		return nil
	}
	return errors.Wrapf(tasks.ErrCanNotAssign, "expect type : *time.Skew, got : %T", injectable)
}

func (s *Skew) Inject(pid tasks.PID) error {
	s.locker.Lock()
	defer s.locker.Unlock()
	sysPID, ok := pid.(tasks.SysPID)
	if !ok {
		return tasks.ErrNotSysPID
	}
	err := s.clockGetTime.AttachToProcess(int(sysPID), map[string]uint64{
		externVarClockIdsMask: s.SkewConfig.clockIDsMask,
		externVarTvSecDelta:   uint64(s.SkewConfig.deltaSeconds),
		externVarTvNsecDelta:  uint64(s.SkewConfig.deltaNanoSeconds),
	})
	if err != nil {
		return err
	}

	err = s.getTimeOfDay.AttachToProcess(int(sysPID), map[string]uint64{
		externVarTvSecDelta:  uint64(s.SkewConfig.deltaSeconds),
		externVarTvNsecDelta: uint64(s.SkewConfig.deltaNanoSeconds),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Skew) Recover(pid tasks.PID) error {
	s.locker.Lock()
	defer s.locker.Unlock()
	sysPID, ok := pid.(tasks.SysPID)
	if !ok {
		return tasks.ErrNotSysPID
	}
	err1 := s.clockGetTime.Recover(int(sysPID))
	if err1 != nil {
		err2 := s.getTimeOfDay.Recover(int(sysPID))
		if err2 != nil {
			return errors.Wrapf(err1, "time skew all failed %v", err2)
		}
		return err1
	}

	err2 := s.getTimeOfDay.Recover(int(sysPID))
	if err2 != nil {
		return err2
	}

	return nil
}
