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
	"bytes"
	"runtime"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/mapreader"
	"github.com/chaos-mesh/chaos-mesh/pkg/ptrace"
)

type CompositeInjector struct {
	injectors []FakeClockInjector
}

// timeSkewFakeImage is the filename of fake image after compiling
const timeSkewFakeImage = "fake_clock_gettime.o"

// vdsoEntryName is the name of the vDSO entry
const vdsoEntryName = "[vdso]"

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

	if image, err = LoadFakeImageFromEmbedFs(timeSkewFakeImage); err != nil {
		return nil, errors.Wrap(err, "load fake image")
	}

	return NewSkewClockGetTimeWithCustomFakeImage(deltaSeconds, deltaNanoSeconds, clockIDsMask, image), nil
}

func NewSkewClockGetTimeWithCustomFakeImage(deltaSeconds int64, deltaNanoSeconds int64, clockIDsMask uint64, fakeImage *FakeImage) *SkewClockGetTime {
	return &SkewClockGetTime{deltaSeconds: deltaSeconds, deltaNanoSeconds: deltaNanoSeconds, clockIDsMask: clockIDsMask, fakeImage: fakeImage}
}

func (it *SkewClockGetTime) Inject(pid int) error {

	runtime.LockOSThread()
	defer func() {
		runtime.UnlockOSThread()
	}()

	program, err := ptrace.Trace(pid)
	if err != nil {
		return errors.Wrapf(err, "ptrace on target process, pid: %d", pid)
	}
	defer func() {
		err = program.Detach()
		if err != nil {
			log.Error(err, "fail to detach program", "pid", program.Pid())
		}
	}()

	var vdsoEntry *mapreader.Entry
	for index := range program.Entries {
		// reverse loop is faster
		e := program.Entries[len(program.Entries)-index-1]
		if e.Path == vdsoEntryName {
			vdsoEntry = &e
			break
		}
	}
	if vdsoEntry == nil {
		return errors.Errorf("cannot find [vdso] entry, pid: %d", pid)
	}

	// minus tailing variable part
	// every variable has 8 bytes
	constImageLen := len(it.fakeImage.content) - 8*len(it.fakeImage.offset)
	var fakeEntry *mapreader.Entry

	// find injected image to avoid redundant inject (which will lead to memory leak)
	for _, e := range program.Entries {
		e := e

		image, err := program.ReadSlice(e.StartAddress, uint64(constImageLen))
		if err != nil {
			continue
		}

		if bytes.Equal(*image, it.fakeImage.content[0:constImageLen]) {
			fakeEntry = &e
			log.Info("found injected image", "addr", fakeEntry.StartAddress, "pid", pid)
			break
		}
	}

	// target process has not been injected yet
	if fakeEntry == nil {
		fakeEntry, err = program.MmapSlice(it.fakeImage.content)
		if err != nil {
			return errors.Wrapf(err, "mmap fake image, pid: %d", pid)
		}

		originAddr, err := program.FindSymbolInEntry(clockGettime, vdsoEntry)
		if err != nil {
			return errors.Wrapf(err, "find origin clock_gettime in vdso, pid: %d", pid)
		}

		err = program.JumpToFakeFunc(originAddr, fakeEntry.StartAddress)
		if err != nil {
			return errors.Wrapf(err, "override origin clock_gettime, pid: %d", pid)
		}
	}

	err = program.WriteUint64ToAddr(fakeEntry.StartAddress+uint64(it.fakeImage.offset[externVarClockIdsMask]), it.clockIDsMask)
	if err != nil {
		return errors.Wrapf(err, "set %s for time skew, pid: %d", externVarClockIdsMask, pid)
	}

	err = program.WriteUint64ToAddr(fakeEntry.StartAddress+uint64(it.fakeImage.offset[externVarTvSecDelta]), uint64(it.deltaSeconds))
	if err != nil {
		return errors.Wrapf(err, "set %s for time skew, pid: %d", externVarTvSecDelta, pid)
	}

	err = program.WriteUint64ToAddr(fakeEntry.StartAddress+uint64(it.fakeImage.offset[externVarTvNsecDelta]), uint64(it.deltaNanoSeconds))
	if err != nil {
		return errors.Wrapf(err, "set %s for time skew, pid: %d", externVarTvNsecDelta, pid)
	}
	return nil
}

func (it *SkewClockGetTime) Recover(pid int) error {
	zeroSkew := NewSkewClockGetTimeWithCustomFakeImage(0, 0, it.clockIDsMask, it.fakeImage)
	return zeroSkew.Inject(pid)
}

// timeofdaySkewFakeImage is the filename of fake image after compiling
const timeOfDaySkewFakeImage = "fake_gettimeofday.o"

// vdsoEntryName is the name of the vDSO entry
// const vdsoEntryName = "[vdso]"

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

	if image, err = LoadFakeImageFromEmbedFs(timeOfDaySkewFakeImage); err != nil {
		return nil, errors.Wrap(err, "load fake image")
	}

	return NewSkewGetTimeOfDayWithCustomFakeImage(deltaSeconds, deltaNanoSeconds, image), nil
}

func NewSkewGetTimeOfDayWithCustomFakeImage(deltaSeconds int64, deltaNanoSeconds int64, fakeImage *FakeImage) *SkewGetTimeOfDay {
	return &SkewGetTimeOfDay{deltaSeconds: deltaSeconds, deltaNanoSeconds: deltaNanoSeconds, fakeImage: fakeImage}
}

func (it *SkewGetTimeOfDay) Inject(pid int) error {

	runtime.LockOSThread()
	defer func() {
		runtime.UnlockOSThread()
	}()

	program, err := ptrace.Trace(pid)
	if err != nil {
		return errors.Wrapf(err, "ptrace on target process, pid: %d", pid)
	}
	defer func() {
		err = program.Detach()
		if err != nil {
			log.Error(err, "fail to detach program", "pid", program.Pid())
		}
	}()

	var vdsoEntry *mapreader.Entry
	for index := range program.Entries {
		// reverse loop is faster
		e := program.Entries[len(program.Entries)-index-1]
		if e.Path == vdsoEntryName {
			vdsoEntry = &e
			break
		}
	}
	if vdsoEntry == nil {
		return errors.Errorf("cannot find [vdso] entry, pid: %d", pid)
	}

	// minus tailing variable part
	// every variable has 8 bytes
	constImageLen := len(it.fakeImage.content) - 8*len(it.fakeImage.offset)
	var fakeEntry *mapreader.Entry

	// find injected image to avoid redundant inject (which will lead to memory leak)
	for _, e := range program.Entries {
		e := e

		image, err := program.ReadSlice(e.StartAddress, uint64(constImageLen))
		if err != nil {
			continue
		}

		if bytes.Equal(*image, it.fakeImage.content[0:constImageLen]) {
			fakeEntry = &e
			log.Info("found injected image", "addr", fakeEntry.StartAddress, "pid", pid)
			break
		}
	}

	// target process has not been injected yet
	if fakeEntry == nil {
		fakeEntry, err = program.MmapSlice(it.fakeImage.content)
		if err != nil {
			return errors.Wrapf(err, "mmap fake image, pid: %d", pid)
		}

		originAddr, err := program.FindSymbolInEntry(getTimeOfDay, vdsoEntry)
		if err != nil {
			return errors.Wrapf(err, "find origin gettimeofday in vdso, pid: %d", pid)
		}

		err = program.JumpToFakeFunc(originAddr, fakeEntry.StartAddress)
		if err != nil {
			return errors.Wrapf(err, "override origin gettimeofday, pid: %d", pid)
		}
	}

	err = program.WriteUint64ToAddr(fakeEntry.StartAddress+uint64(it.fakeImage.offset[externVarTvSecDelta]), uint64(it.deltaSeconds))
	if err != nil {
		return errors.Wrapf(err, "set %s for time skew, pid: %d", externVarTvSecDelta, pid)
	}

	err = program.WriteUint64ToAddr(fakeEntry.StartAddress+uint64(it.fakeImage.offset[externVarTvNsecDelta]), uint64(it.deltaNanoSeconds))
	if err != nil {
		return errors.Wrapf(err, "set %s for time skew, pid: %d", externVarTvNsecDelta, pid)
	}
	return nil
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
