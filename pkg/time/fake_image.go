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
	"github.com/chaos-mesh/chaos-mesh/pkg/mapreader"
	"github.com/chaos-mesh/chaos-mesh/pkg/ptrace"
	"github.com/pkg/errors"
	"runtime"
)

// vdsoEntryName is the name of the vDSO entry
const vdsoEntryName = "[vdso]"

// clockGettime is the target function would be replaced
const clockGettime = "clock_gettime"

// FakeImage introduce the replacement of VDSO ELF entry and customizable variables.
// FakeImage could be constructed by LoadFakeImageFromEmbedFs(), and then used by FakeClockInjector.
type FakeImage struct {
	// content presents .text section which has been "manually relocation", the address of extern variables have been calculated manually
	content []byte
	// offset stores the table with variable name, and it's address in content.
	// the key presents extern variable name, ths value is the address/offset within the content.
	offset map[string]int
}

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
			log.Info("found injected image", "addr", fakeEntry.StartAddress, "pid", pid)
			break
		}
	}

	// target process has not been injected yet
	if fakeEntry == nil {
		fakeEntry, err = program.MmapSlice(it.content)
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
