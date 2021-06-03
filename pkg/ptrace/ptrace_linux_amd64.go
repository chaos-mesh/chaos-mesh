// Copyright 2020 Chaos Mesh Authors.
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

// +build cgo

package ptrace

/*
#include <stdint.h>
struct iovec {
	intptr_t iov_base;
	size_t iov_len;
};
*/
import "C"

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/pkg/mapreader"
)

var log = ctrl.Log.WithName("ptrace")

// RegisterLogger registers a logger on ptrace pkg
func RegisterLogger(logger logr.Logger) {
	log = logger
}

const waitPidErrorMessage = "waitpid ret value: %d"

// If it's on 64-bit platform, `^uintptr(0)` will get a 64-bit number full of one.
// After shifting right for 63-bit, only 1 will be left. Than we got 8 here.
// If it's on 32-bit platform, After shifting nothing will be left. Than we got 4 here.
const ptrSize = 4 << uintptr(^uintptr(0)>>63)

var threadRetryLimit = 10

// TracedProgram is a program traced by ptrace
type TracedProgram struct {
	pid     int
	tids    []int
	Entries []mapreader.Entry

	backupRegs *syscall.PtraceRegs
	backupCode []byte
}

// Pid return the pid of traced program
func (p *TracedProgram) Pid() int {
	return p.pid
}

func waitPid(pid int) error {
	ret := waitpid(pid)
	if ret == pid {
		return nil
	}

	return errors.Errorf(waitPidErrorMessage, ret)
}

// Trace ptrace all threads of a process
func Trace(pid int) (*TracedProgram, error) {
	traceSuccess := false

	tidMap := make(map[int]bool)
	retryCount := make(map[int]int)
	for {
		threads, err := ioutil.ReadDir(fmt.Sprintf("/proc/%d/task", pid))
		if err != nil {
			log.Error(err, "read failed", "pid", pid)
			return nil, errors.WithStack(err)
		}

		// judge whether `threads` is a subset of `tidMap`
		subset := true

		tids := make(map[int]bool)
		for _, thread := range threads {
			tid64, err := strconv.ParseInt(thread.Name(), 10, 32)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			tid := int(tid64)

			_, ok := tidMap[tid]
			if ok {
				tids[tid] = true
				continue
			}
			subset = false

			err = syscall.PtraceAttach(tid)
			if err != nil {
				_, ok := retryCount[tid]
				if !ok {
					retryCount[tid] = 1
				} else {
					retryCount[tid]++
				}
				if retryCount[tid] < threadRetryLimit {
					log.Info("retry attaching thread", "tid", tid, "retryCount", retryCount[tid], "limit", threadRetryLimit)
					continue
				}

				if !strings.Contains(err.Error(), "no such process") {
					log.Error(err, "attach failed", "tid", tid)
					return nil, errors.WithStack(err)
				}
				continue
			}
			defer func() {
				if !traceSuccess {
					err = syscall.PtraceDetach(tid)
					if err != nil {
						if !strings.Contains(err.Error(), "no such process") {
							log.Error(err, "detach failed", "tid", tid)
						}
					}
				}
			}()

			err = waitPid(tid)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			log.Info("attach successfully", "tid", tid)
			tids[tid] = true
			tidMap[tid] = true
		}

		if subset {
			tidMap = tids
			break
		}
	}

	var tids []int
	for key := range tidMap {
		tids = append(tids, key)
	}

	entries, err := mapreader.Read(pid)
	if err != nil {
		return nil, err
	}

	program := &TracedProgram{
		pid:        pid,
		tids:       tids,
		Entries:    entries,
		backupRegs: &syscall.PtraceRegs{},
		backupCode: make([]byte, ptrSize),
	}

	traceSuccess = true

	return program, nil
}

// Detach detaches from all threads of the processes
func (p *TracedProgram) Detach() error {
	for _, tid := range p.tids {
		log.Info("detaching", "tid", tid)

		err := syscall.PtraceDetach(tid)

		if err != nil {
			if !strings.Contains(err.Error(), "no such process") {
				log.Error(err, "detach failed", "tid", tid)
				return errors.WithStack(err)
			}
		}
	}

	log.Info("Successfully detach and rerun process", "pid", p.pid)
	return nil
}

// Protect will backup regs and rip into fields
func (p *TracedProgram) Protect() error {
	err := syscall.PtraceGetRegs(p.pid, p.backupRegs)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = syscall.PtracePeekData(p.pid, uintptr(p.backupRegs.Rip), p.backupCode)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Restore will restore regs and rip from fields
func (p *TracedProgram) Restore() error {
	err := syscall.PtraceSetRegs(p.pid, p.backupRegs)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = syscall.PtracePokeData(p.pid, uintptr(p.backupRegs.Rip), p.backupCode)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Wait waits until the process stops
func (p *TracedProgram) Wait() error {
	return waitPid(p.pid)
}

// Step moves one step forward
func (p *TracedProgram) Step() error {
	err := syscall.PtraceSingleStep(p.pid)
	if err != nil {
		return errors.WithStack(err)
	}

	return p.Wait()
}

// Syscall runs a syscall at main thread of process
func (p *TracedProgram) Syscall(number uint64, args ...uint64) (uint64, error) {
	err := p.Protect()
	if err != nil {
		return 0, err
	}

	var regs syscall.PtraceRegs

	err = syscall.PtraceGetRegs(p.pid, &regs)
	if err != nil {
		return 0, err
	}
	regs.Rax = number
	for index, arg := range args {
		// All these registers are hard coded for x86 platform
		if index == 0 {
			regs.Rdi = arg
		} else if index == 1 {
			regs.Rsi = arg
		} else if index == 2 {
			regs.Rdx = arg
		} else if index == 3 {
			regs.R10 = arg
		} else if index == 4 {
			regs.R8 = arg
		} else if index == 5 {
			regs.R9 = arg
		} else {
			return 0, fmt.Errorf("too many arguments for a syscall")
		}
	}
	err = syscall.PtraceSetRegs(p.pid, &regs)
	if err != nil {
		return 0, err
	}

	ip := make([]byte, ptrSize)

	// We only support x86-64 platform now, so using hard coded `LittleEndian` here is ok.
	binary.LittleEndian.PutUint16(ip, 0x050f)
	_, err = syscall.PtracePokeData(p.pid, uintptr(p.backupRegs.Rip), ip)
	if err != nil {
		return 0, err
	}

	err = p.Step()
	if err != nil {
		return 0, err
	}

	err = syscall.PtraceGetRegs(p.pid, &regs)
	if err != nil {
		return 0, err
	}

	return regs.Rax, p.Restore()
}

// Mmap runs mmap syscall
func (p *TracedProgram) Mmap(length uint64, fd uint64) (uint64, error) {
	return p.Syscall(syscall.SYS_MMAP, 0, length, syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC, syscall.MAP_ANON|syscall.MAP_PRIVATE, fd, 0)
}

// ReadSlice reads from addr and return a slice
func (p *TracedProgram) ReadSlice(addr uint64, size uint64) (*[]byte, error) {
	buffer := make([]byte, size)

	localIov := C.struct_iovec{
		iov_base: C.long(uintptr(unsafe.Pointer(&buffer[0]))),
		iov_len:  C.ulong(size),
	}

	remoteIov := C.struct_iovec{
		iov_base: C.long(addr),
		iov_len:  C.ulong(size),
	}

	// process_vm_readv syscall number is 310
	_, _, errno := syscall.Syscall6(310, uintptr(p.pid), uintptr(unsafe.Pointer(&localIov)), uintptr(1), uintptr(unsafe.Pointer(&remoteIov)), uintptr(1), uintptr(0))
	if errno != 0 {
		return nil, errors.WithStack(errno)
	}
	// TODO: check size and warn

	return &buffer, nil
}

// WriteSlice writes a buffer into addr
func (p *TracedProgram) WriteSlice(addr uint64, buffer []byte) error {
	size := len(buffer)

	localIov := C.struct_iovec{
		iov_base: C.long(uintptr(unsafe.Pointer(&buffer[0]))),
		iov_len:  C.ulong(size),
	}

	remoteIov := C.struct_iovec{
		iov_base: C.long(addr),
		iov_len:  C.ulong(size),
	}

	// process_vm_writev syscall number is 311
	_, _, errno := syscall.Syscall6(311, uintptr(p.pid), uintptr(unsafe.Pointer(&localIov)), uintptr(1), uintptr(unsafe.Pointer(&remoteIov)), uintptr(1), uintptr(0))
	if errno != 0 {
		return errors.WithStack(errno)
	}
	// TODO: check size and warn

	return nil
}

func alignBuffer(buffer []byte) []byte {
	if buffer == nil {
		return nil
	}

	alignedSize := (len(buffer) / ptrSize) * ptrSize
	if alignedSize < len(buffer) {
		alignedSize += ptrSize
	}
	clonedBuffer := make([]byte, alignedSize)
	copy(clonedBuffer, buffer)

	return clonedBuffer
}

// PtraceWriteSlice uses ptrace rather than process_vm_write to write a buffer into addr
func (p *TracedProgram) PtraceWriteSlice(addr uint64, buffer []byte) error {
	wroteSize := 0

	buffer = alignBuffer(buffer)

	for wroteSize+ptrSize <= len(buffer) {
		addr := uintptr(addr + uint64(wroteSize))
		data := buffer[wroteSize : wroteSize+ptrSize]

		_, err := syscall.PtracePokeData(p.pid, addr, data)
		if err != nil {
			err = errors.WithStack(err)
			return errors.WithMessagef(err, "write to addr %x with %+v failed", addr, data)
		}

		wroteSize += ptrSize
	}

	return nil
}

// GetLibBuffer reads an entry
func (p *TracedProgram) GetLibBuffer(entry *mapreader.Entry) (*[]byte, error) {
	if entry.PaddingSize > 0 {
		return nil, fmt.Errorf("entry with padding size is not supported")
	}

	size := entry.EndAddress - entry.StartAddress

	return p.ReadSlice(entry.StartAddress, size)
}

// MmapSlice mmaps a slice and return it's addr
func (p *TracedProgram) MmapSlice(slice []byte) (*mapreader.Entry, error) {
	size := uint64(len(slice))

	addr, err := p.Mmap(size, 0)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = p.WriteSlice(addr, slice)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &mapreader.Entry{
		StartAddress: addr,
		EndAddress:   addr + size,
		Privilege:    "rwxp",
		PaddingSize:  0,
		Path:         "",
	}, nil
}

// FindSymbolInEntry finds symbol in entry through parsing elf
func (p *TracedProgram) FindSymbolInEntry(symbolName string, entry *mapreader.Entry) (uint64, error) {
	libBuffer, err := p.GetLibBuffer(entry)
	if err != nil {
		return 0, err
	}

	reader := bytes.NewReader(*libBuffer)
	vdsoElf, err := elf.NewFile(reader)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	loadOffset := uint64(0)

	for _, prog := range vdsoElf.Progs {
		if prog.Type == elf.PT_LOAD {
			loadOffset = prog.Vaddr - prog.Off

			// break here is enough for vdso
			break
		}
	}

	symbols, err := vdsoElf.DynamicSymbols()
	if err != nil {
		return 0, errors.WithStack(err)
	}
	for _, symbol := range symbols {
		if symbol.Name == symbolName {
			offset := symbol.Value

			return entry.StartAddress + (offset - loadOffset), nil
		}
	}
	return 0, fmt.Errorf("cannot find symbol")
}

// WriteUint64ToAddr writes uint64 to addr
func (p *TracedProgram) WriteUint64ToAddr(addr uint64, value uint64) error {
	valueSlice := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueSlice, value)
	err := p.WriteSlice(addr, valueSlice)
	return err
}

// JumpToFakeFunc writes jmp instruction to jump to fake function
func (p *TracedProgram) JumpToFakeFunc(originAddr uint64, targetAddr uint64) error {
	instructions := make([]byte, 16)

	// mov rax, targetAddr;
	// jmp rax ;
	instructions[0] = 0x48
	instructions[1] = 0xb8
	binary.LittleEndian.PutUint64(instructions[2:10], targetAddr)
	instructions[10] = 0xff
	instructions[11] = 0xe0

	return p.PtraceWriteSlice(originAddr, instructions)
}
