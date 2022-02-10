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
	"github.com/pkg/errors"
)

type CompositeInjector struct {
	injectors []FakeClockInjector
}

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

type SkewClockGetTime struct {
	deltaSeconds     int64
	deltaNanoSeconds int64
	clockIDsMask     uint64
	fakeImage        *FakeImage
}

func NewSkewClockGetTime(deltaSeconds int64, deltaNanoSeconds int64, clockIDsMask uint64) (*SkewClockGetTime, error) {
	var image *FakeImage
	var err error

	if image, err = LoadFakeImageFromEmbedFs(clockGettimeSkewFakeImage, clockGettime); err != nil {
		return nil, errors.Wrap(err, "load fake image")
	}

	return NewSkewClockGetTimeWithCustomFakeImage(deltaSeconds, deltaNanoSeconds, clockIDsMask, image), nil
}

func NewSkewClockGetTimeWithCustomFakeImage(deltaSeconds int64, deltaNanoSeconds int64, clockIDsMask uint64, fakeImage *FakeImage) *SkewClockGetTime {
	return &SkewClockGetTime{deltaSeconds: deltaSeconds, deltaNanoSeconds: deltaNanoSeconds, clockIDsMask: clockIDsMask, fakeImage: fakeImage}
}

func (it *SkewClockGetTime) Inject(pid int) error {
	return it.fakeImage.AttachToProcess(pid, map[string]uint64{
		externVarClockIdsMask: it.clockIDsMask,
		externVarTvSecDelta:   uint64(it.deltaSeconds),
		externVarTvNsecDelta:  uint64(it.deltaNanoSeconds),
	})
}

func (it *SkewClockGetTime) Recover(pid int) error {
	zeroSkew := NewSkewClockGetTimeWithCustomFakeImage(0, 0, it.clockIDsMask, it.fakeImage)
	return zeroSkew.Inject(pid)
}

// timeofdaySkewFakeImage is the filename of fake image after compiling
const timeOfDaySkewFakeImage = "fake_gettimeofday.o"

// getTimeOfDay is the target function would be replaced
const getTimeOfDay = "gettimeofday"

type SkewGetTimeOfDay struct {
	deltaSeconds     int64
	deltaNanoSeconds int64
	fakeImage        *FakeImage
}

func NewSkewGetTimeOfDay(deltaSeconds int64, deltaNanoSeconds int64) (*SkewGetTimeOfDay, error) {
	var image *FakeImage
	var err error

	if image, err = LoadFakeImageFromEmbedFs(timeOfDaySkewFakeImage, getTimeOfDay); err != nil {
		return nil, errors.Wrap(err, "load fake image")
	}

	return NewSkewGetTimeOfDayWithCustomFakeImage(deltaSeconds, deltaNanoSeconds, image), nil
}

func NewSkewGetTimeOfDayWithCustomFakeImage(deltaSeconds int64, deltaNanoSeconds int64, fakeImage *FakeImage) *SkewGetTimeOfDay {
	return &SkewGetTimeOfDay{deltaSeconds: deltaSeconds, deltaNanoSeconds: deltaNanoSeconds, fakeImage: fakeImage}
}

func (it *SkewGetTimeOfDay) Inject(pid int) error {
	return it.fakeImage.AttachToProcess(pid, map[string]uint64{
		externVarTvSecDelta:  uint64(it.deltaSeconds),
		externVarTvNsecDelta: uint64(it.deltaNanoSeconds),
	})
}

func (it *SkewGetTimeOfDay) Recover(pid int) error {
	zeroSkew := NewSkewGetTimeOfDayWithCustomFakeImage(0, 0, it.fakeImage)
	return zeroSkew.Inject(pid)
}

func (it *CompositeInjector) Inject(pid int) error {
	for _, injector := range it.injectors {
		if err := injector.Inject(pid); err != nil {
			return err
		}
	}
	return nil
}

func (it *CompositeInjector) Recover(pid int) error {
	for _, injector := range it.injectors {
		if err := injector.Recover(pid); err != nil {
			return err
		}
	}
	return nil
}

func NewTimeSkew(deltaSec int64, deltaNsec int64, clockIdsMask uint64) (FakeClockInjector, error) {
	skewClockGetTime, err := NewSkewClockGetTime(deltaSec, deltaNsec, clockIdsMask)
	if err != nil {
		return nil, err
	}
	skewGetTimeOfDay, err := NewSkewGetTimeOfDay(deltaSec, deltaNsec)
	if err != nil {
		return nil, err
	}
	return &CompositeInjector{injectors: []FakeClockInjector{skewClockGetTime, skewGetTimeOfDay}}, nil
}
