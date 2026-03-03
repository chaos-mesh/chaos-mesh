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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/cerr"
	"github.com/chaos-mesh/chaos-mesh/pkg/mapreader"
	"github.com/chaos-mesh/chaos-mesh/pkg/ptrace"
)

// vdsoEntryName is the name of the vDSO entry
const vdsoEntryName = "[vdso]"

// FakeImage introduce the replacement of VDSO ELF entry and customizable variables.
// FakeImage could be constructed by LoadFakeImageFromEmbedFs(), and then used by FakeClockInjector.
type FakeImage struct {
	// symbolName is the name of the symbol to be replaced.
	symbolName string
	// content presents .text section which has been "manually relocation", the address of extern variables have been calculated manually
	content []byte
	// offset stores the table with variable name, and it's address in content.
	// the key presents extern variable name, ths value is the address/offset within the content.
	offset map[string]int
	// OriginFuncCode stores the raw func code like getTimeOfDay & ClockGetTime.
	OriginFuncCode []byte
	// OriginAddress stores the origin address of OriginFuncCode.
	OriginAddress uint64
	// fakeEntry stores the fake entry
	fakeEntry *mapreader.Entry

	logger logr.Logger
}

func NewFakeImage(symbolName string, content []byte, offset map[string]int, logger logr.Logger) *FakeImage {
	return &FakeImage{symbolName: symbolName, content: content, offset: offset, logger: logger}
}

// AttachToProcess would use ptrace to replace the VDSO ELF entry with FakeImage.
// Each item in parameter "variables" needs a corresponding entry in FakeImage.offset.
func (it *FakeImage) AttachToProcess(pid int, variables map[string]uint64) (err error) {
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

	vdsoEntry, err := FindVDSOEntry(program)
	if err != nil {
		return errors.Wrapf(err, "PID : %d", pid)
	}

	fakeEntry, err := it.FindInjectedImage(program, len(variables))
	if err != nil {
		return errors.Wrapf(err, "PID : %d", pid)
	}
	// target process has not been injected yet
	if fakeEntry == nil {
		fakeEntry, err = it.InjectFakeImage(program, vdsoEntry)
		if err != nil {
			return errors.Wrapf(err, "injecting fake image , PID : %d", pid)
		}
		defer func() {
			if err != nil {
				errIn := it.TryReWriteFakeImage(program)
				if errIn != nil {
					it.logger.Error(errIn, "rewrite fail, recover fail")
				}
				it.OriginFuncCode = nil
				it.OriginAddress = 0
			}
		}()
	}

	for k, v := range variables {
		err = it.SetVarUint64(program, fakeEntry, k, v)

		if err != nil {
			return errors.Wrapf(err, "set %s for time skew, pid: %d", k, pid)
		}
	}

	return
}

func FindVDSOEntry(program *ptrace.TracedProgram) (*mapreader.Entry, error) {
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
		return nil, cerr.NotFound("VDSOEntry").Err()
	}
	return vdsoEntry, nil
}

// FindInjectedImage find injected image to avoid redundant inject.
func (it *FakeImage) FindInjectedImage(program *ptrace.TracedProgram, varNum int) (*mapreader.Entry, error) {
	it.logger.Info("finding injected image")

	// minus tailing variable part
	// every variable has 8 bytes
	if it.fakeEntry != nil {
		content, err := program.ReadSlice(it.fakeEntry.StartAddress, it.fakeEntry.EndAddress-it.fakeEntry.StartAddress)
		if err != nil {
			it.logger.Info("ReadSlice fail")
			return nil, nil
		}
		if varNum*8 > len(it.content) {
			return nil, errors.New("variable num bigger than content num")
		}
		contentWithoutVariable := (*content)[:len(it.content)-varNum*varLength]
		expectedContentWithoutVariable := it.content[:len(it.content)-varNum*varLength]
		it.logger.Info("successfully read slice", "content", contentWithoutVariable, "expected content", expectedContentWithoutVariable)

		if bytes.Equal(contentWithoutVariable, expectedContentWithoutVariable) {
			it.logger.Info("slice found")
			return it.fakeEntry, nil
		}
		it.logger.Info("slice not found")
	}
	return nil, nil
}

// InjectFakeImage Usage CheckList:
// When error : TryReWriteFakeImage after InjectFakeImage.
func (it *FakeImage) InjectFakeImage(program *ptrace.TracedProgram,
	vdsoEntry *mapreader.Entry) (*mapreader.Entry, error) {
	fakeEntry, err := program.MmapSlice(it.content)
	if err != nil {
		return nil, errors.Wrapf(err, "mmap fake image")
	}
	it.fakeEntry = fakeEntry
	originAddr, size, err := program.FindSymbolInEntry(it.symbolName, vdsoEntry)
	if err != nil {
		return nil, errors.Wrapf(err, "find origin %s in vdso", it.symbolName)
	}
	funcBytes, err := program.ReadSlice(originAddr, size)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadSlice failed")
	}
	err = program.JumpToFakeFunc(originAddr, fakeEntry.StartAddress)
	if err != nil {
		errIn := it.TryReWriteFakeImage(program)
		if errIn != nil {
			it.logger.Error(errIn, "rewrite fail, recover fail")
		}
		return nil, errors.Wrapf(err, "override origin %s", it.symbolName)
	}

	it.OriginFuncCode = *funcBytes
	it.OriginAddress = originAddr
	return fakeEntry, nil
}

func (it *FakeImage) TryReWriteFakeImage(program *ptrace.TracedProgram) error {
	if it.OriginFuncCode != nil {
		err := program.PtraceWriteSlice(it.OriginAddress, it.OriginFuncCode)
		if err != nil {
			return err
		}
		it.OriginFuncCode = nil
		it.OriginAddress = 0
	}
	return nil
}

// Recover the injected image. If injected image not found ,
// Recover will not return error.
func (it *FakeImage) Recover(pid int, vars map[string]uint64) error {
	runtime.LockOSThread()
	defer func() {
		runtime.UnlockOSThread()
	}()
	if it.OriginFuncCode == nil {
		return nil
	}
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

	fakeEntry, err := it.FindInjectedImage(program, len(vars))
	if err != nil {
		return errors.Wrapf(err, "FindInjectedImage , pid: %d", pid)
	}
	if fakeEntry == nil {
		return nil
	}

	err = it.TryReWriteFakeImage(program)
	return err
}
