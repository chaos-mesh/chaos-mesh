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

// AttachToProcess would use ptrace to replace the VDSO ELF entry with FakeImage.
// Each item in parameter "variables" needs a corresponding entry in FakeImage.offset.
func (it *FakeImage) AttachToProcess(pid int, variables map[string]uint64) error {
	if len(variables) != len(it.offset) {
		return errors.New("fake image: extern variable number not match")
	}

	runtime.LockOSThread()
	defer func() {
		runtime.UnlockOSThread()
	}()

	program, err := ptrace.Trace(pid, it.logger.WithName("ptrace").WithValues("pid", pid))
	if err != nil {
		return errors.Wrapf(err, "ptrace on target process, pid: %d", pid)
	}
	defer func() {
		err = program.Detach()
		if err != nil {
			it.logger.Error(err, "fail to detach program", "pid", program.Pid())
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
	constImageLen := len(it.content) - 8*len(it.offset)
	var fakeEntry *mapreader.Entry

	// find injected image to avoid redundant inject (which will lead to memory leak)
	for _, e := range program.Entries {
		e := e

		image, err := program.ReadSlice(e.StartAddress, uint64(constImageLen))
		if err != nil {
			continue
		}

		if bytes.Equal(*image, it.content[0:constImageLen]) {
			fakeEntry = &e
			it.logger.Info("found injected image", "addr", fakeEntry.StartAddress, "pid", pid)
			break
		}
	}

	// target process has not been injected yet
	if fakeEntry == nil {
		fakeEntry, err = program.MmapSlice(it.content)
		if err != nil {
			return errors.Wrapf(err, "mmap fake image, pid: %d", pid)
		}

		originAddr, err := program.FindSymbolInEntry(it.symbolName, vdsoEntry)
		if err != nil {
			return errors.Wrapf(err, "find origin clock_gettime in vdso, pid: %d", pid)
		}

		err = program.JumpToFakeFunc(originAddr, fakeEntry.StartAddress)
		if err != nil {
			return errors.Wrapf(err, "override origin clock_gettime, pid: %d", pid)
		}
	}

	for k, v := range variables {
		if offset, ok := it.offset[k]; ok {
			err = program.WriteUint64ToAddr(fakeEntry.StartAddress+uint64(offset), v)
			if err != nil {
				return errors.Wrapf(err, "set %s for time skew, pid: %d", k, pid)
			}
		} else {
			return errors.Errorf("no such extern variable in fake image: %s", k)
		}
	}

	return nil
}
