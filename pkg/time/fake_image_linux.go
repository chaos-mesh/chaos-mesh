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
	"encoding/binary"
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
	// validateProcess runs after ptrace has stopped the process, before using
	// addresses from a persisted recovery record.
	validateProcess func(int) error
	// saveRecovery records the original code before the first write to the vDSO.
	saveRecovery       func() error
	recoveryCandidates []imageRecovery

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
		if detachErr := program.Detach(); detachErr != nil {
			it.logger.Error(detachErr, "fail to detach program", "pid", program.Pid())
			if err == nil {
				err = detachErr
			} else {
				err = errors.Wrapf(err, "detach also failed: %v", detachErr)
			}
		}
	}()
	if it.validateProcess != nil {
		if err := it.validateProcess(pid); err != nil {
			return err
		}
	}

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
	} else {
		// A fork can inherit a patch without an explicit record of its own.
		// Save the adopted originals before changing any of its code or data.
		if it.saveRecovery != nil {
			if err := it.saveRecovery(); err != nil {
				return err
			}
		}
		if err := program.JumpToFakeFunc(it.OriginAddress, fakeEntry.StartAddress); err != nil {
			if restoreErr := it.TryReWriteFakeImage(program); restoreErr != nil {
				return errors.Wrapf(err, "restore after failed jump: %v", restoreErr)
			}
			return err
		}
	}

	defer func() {
		if err != nil {
			if restoreErr := it.TryReWriteFakeImage(program); restoreErr != nil {
				err = errors.Wrapf(err, "restore after failed update: %v", restoreErr)
			}
		}
	}()
	for k, v := range variables {
		err = it.SetVarUint64(program, fakeEntry, k, v)

		if err != nil {
			return errors.Wrapf(err, "set %s for time skew, pid: %d", k, pid)
		}
	}

	return
}

func FindVDSOEntry(program *ptrace.TracedProgram) (*mapreader.Entry, error) {
	return findVDSOEntry(program.Entries)
}

func findVDSOEntry(entries []mapreader.Entry) (*mapreader.Entry, error) {
	var vdsoEntry *mapreader.Entry
	for index := range entries {
		// reverse loop is faster
		e := entries[len(entries)-index-1]
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
func (it *FakeImage) FindInjectedImage(program *ptrace.TracedProgram, _ int) (*mapreader.Entry, error) {
	return it.findSavedImage(program, program.Entries)
}

func (it *FakeImage) findSavedImage(program imageProgram, entries []mapreader.Entry) (*mapreader.Entry, error) {
	var firstErr error
	candidates := append([]imageRecovery{snapshotImageRecovery(it)}, it.recoveryCandidates...)
	for _, candidate := range candidates {
		if candidate.FakeEntry == nil || len(candidate.OriginalCode) != 16 {
			continue
		}
		found, err := it.matchesSavedImage(program, entries, candidate)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if found {
			restoreImageRecovery(it, candidate)
			return it.fakeEntry, nil
		}
	}
	return nil, firstErr
}

func (it *FakeImage) matchesSavedImage(program imageProgram, entries []mapreader.Entry, candidate imageRecovery) (bool, error) {
	// A replacement daemon may contain a different compiled fake clock. Check
	// the payload saved at injection time, rather than the new embedded object.
	if len(candidate.Content) == 0 {
		return false, errors.New("saved time chaos image has no validation payload")
	}
	varNum := len(candidate.Offsets)
	vdso, err := findVDSOEntry(entries)
	if err != nil {
		return false, err
	}
	address, _, err := program.FindSymbolInEntry(it.symbolName, vdso)
	if err != nil {
		return false, err
	}
	if address != candidate.OriginalAddress {
		return false, nil
	}
	current, err := program.ReadSlice(address, 16)
	if err != nil {
		return false, err
	}
	restored := bytes.Equal(*current, candidate.OriginalCode)
	mapped := false
	for _, entry := range entries {
		if entry.StartAddress <= candidate.FakeEntry.StartAddress && candidate.FakeEntry.StartAddress < entry.EndAddress && uint64(len(candidate.Content)) <= entry.EndAddress-candidate.FakeEntry.StartAddress {
			mapped = true
			break
		}
	}
	// An exec keeps the PID/start time but discards the old address space.
	if !mapped {
		if !restored && matchesTimePatch(*current, candidate.OriginalCode, timeJump(candidate.FakeEntry.StartAddress)) {
			return false, errors.New("time chaos jump remains but its saved image is unmapped")
		}
		return false, nil
	}
	content, err := program.ReadSlice(candidate.FakeEntry.StartAddress, uint64(len(candidate.Content)))
	if err != nil {
		return false, err
	}
	if varNum < 0 || varNum*varLength > len(candidate.Content) {
		return false, errors.New("invalid fake image variable count")
	}
	size := len(candidate.Content) - varNum*varLength
	if !bytes.Equal((*content)[:size], candidate.Content[:size]) {
		if !restored {
			return false, errors.New("time chaos jump remains but its saved image has changed")
		}
		return false, nil
	}
	if !matchesTimePatch(*current, candidate.OriginalCode, timeJump(candidate.FakeEntry.StartAddress)) {
		return false, errors.New("vDSO code differs from the saved time chaos patch")
	}
	return true, nil
}

// PtraceWriteSlice writes machine words separately. A daemon crash can leave
// either the jump or its restoration only partly written; both are recoverable.
func matchesTimePatch(current, original, jump []byte) bool {
	if len(current) != 16 || len(original) != 16 || len(jump) != 16 {
		return false
	}
	for i := 0; i < 16; i += 8 {
		if !bytes.Equal(current[i:i+8], original[i:i+8]) && !bytes.Equal(current[i:i+8], jump[i:i+8]) {
			return false
		}
	}
	return true
}

func timeJump(address uint64) []byte {
	instructions := make([]byte, 16)
	if runtime.GOARCH == "arm64" {
		binary.LittleEndian.PutUint32(instructions, 0x58000049)
		binary.LittleEndian.PutUint32(instructions[4:], 0xD61F0120)
		binary.LittleEndian.PutUint64(instructions[8:], address)
	} else {
		instructions[0], instructions[1] = 0x48, 0xb8
		binary.LittleEndian.PutUint64(instructions[2:], address)
		instructions[10], instructions[11] = 0xff, 0xe0
	}
	return instructions
}

type imageProgram interface {
	MmapSlice([]byte) (*mapreader.Entry, error)
	FindSymbolInEntry(string, *mapreader.Entry) (uint64, uint64, error)
	ReadSlice(uint64, uint64) (*[]byte, error)
	JumpToFakeFunc(uint64, uint64) error
	PtraceWriteSlice(uint64, []byte) error
}

// InjectFakeImage saves restoration data before replacing the vDSO function.
func (it *FakeImage) InjectFakeImage(program imageProgram,
	vdsoEntry *mapreader.Entry) (*mapreader.Entry, error) {
	fakeEntry, err := program.MmapSlice(it.content)
	if err != nil {
		return nil, errors.Wrapf(err, "mmap fake image")
	}
	it.fakeEntry = fakeEntry
	originAddr, _, err := program.FindSymbolInEntry(it.symbolName, vdsoEntry)
	if err != nil {
		return nil, errors.Wrapf(err, "find origin %s in vdso", it.symbolName)
	}
	funcBytes, err := program.ReadSlice(originAddr, 16)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadSlice failed")
	}
	// Without a journal, adopting an existing hook would save the hook itself
	// as the original function, making later recovery ineffective.
	jump := timeJump(0)
	hooked := bytes.Equal((*funcBytes)[:8], jump[:8])
	if runtime.GOARCH == "amd64" {
		hooked = bytes.Equal((*funcBytes)[:2], jump[:2]) && bytes.Equal((*funcBytes)[10:], jump[10:])
	}
	if hooked {
		return nil, errors.New("time chaos patch has no matching recovery state")
	}

	it.OriginFuncCode = *funcBytes
	it.OriginAddress = originAddr
	if it.saveRecovery != nil {
		if err := it.saveRecovery(); err != nil {
			return nil, errors.Wrap(err, "save time skew recovery state")
		}
	}
	err = program.JumpToFakeFunc(originAddr, fakeEntry.StartAddress)
	if err != nil {
		if restoreErr := program.PtraceWriteSlice(originAddr, it.OriginFuncCode); restoreErr != nil {
			return nil, errors.Wrapf(err, "override %s; restore also failed: %v", it.symbolName, restoreErr)
		}
		return nil, errors.Wrapf(err, "override origin %s", it.symbolName)
	}
	return fakeEntry, nil
}

func (it *FakeImage) TryReWriteFakeImage(program *ptrace.TracedProgram) error {
	if it.OriginFuncCode != nil {
		err := program.PtraceWriteSlice(it.OriginAddress, it.OriginFuncCode)
		if err != nil {
			return err
		}
		if it.saveRecovery == nil {
			it.OriginFuncCode = nil
			it.OriginAddress = 0
		}
	}
	return nil
}

// Recover the injected image. If injected image not found ,
// Recover will not return error.
func (it *FakeImage) Recover(pid int, vars map[string]uint64) (err error) {
	runtime.LockOSThread()
	defer func() {
		runtime.UnlockOSThread()
	}()
	if it.OriginFuncCode == nil && len(it.recoveryCandidates) == 0 {
		return nil
	}
	program, err := ptrace.Trace(pid, it.logger.WithName("ptrace").WithValues("pid", pid))
	if err != nil {
		return errors.Wrapf(err, "ptrace on target process, pid: %d", pid)
	}
	defer func() {
		if detachErr := program.Detach(); detachErr != nil {
			it.logger.Error(detachErr, "fail to detach program", "pid", program.Pid())
			if err == nil {
				err = detachErr
			} else {
				err = errors.Wrapf(err, "detach also failed: %v", detachErr)
			}
		}
	}()
	if it.validateProcess != nil {
		if err := it.validateProcess(pid); err != nil {
			return err
		}
	}

	fakeEntry, err := it.FindInjectedImage(program, len(vars))
	if err != nil {
		return errors.Wrapf(err, "FindInjectedImage , pid: %d", pid)
	}
	if fakeEntry == nil {
		return nil
	}

	if it.saveRecovery != nil {
		if err := it.saveRecovery(); err != nil {
			return err
		}
	}
	err = it.TryReWriteFakeImage(program)
	return err
}
