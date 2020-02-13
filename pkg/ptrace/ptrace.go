// Copyright 2020 PingCAP, Inc.
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

package ptrace

/*
#define _GNU_SOURCE
#include <sys/wait.h>
#include <sys/uio.h>
#include <errno.h>
#include <stdint.h>
#include <string.h>
#include <stdlib.h>

void* Uint64ToPointer(uint64_t addr) {
	return (void*) addr;
}
*/
import "C"

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/pingcap/chaos-mesh/pkg/mapreader"
)

var log = ctrl.Log.WithName("ptrace")

const waitPidErrorMessage = "waitpid ret value: %d"

const ptrSize = 4 << uintptr(^uintptr(0)>>63) // Here is a trick to get pointer size in bytes

// TracedProgram is a program traced by ptrace
type TracedProgram struct {
	pid     int
	tids    []int
	Entries *[]mapreader.Entry

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

func constructPartialProgram(pid int, tidMap map[int]bool) *TracedProgram {
	var tids []int
	for key := range tidMap {
		tids = append(tids, key)
	}

	return &TracedProgram{
		pid:        pid,
		tids:       tids,
		Entries:    nil,
		backupRegs: nil,
		backupCode: nil,
	}
}

// Trace ptrace all threads of a process
func Trace(pid int) (*TracedProgram, error) {
	tidMap := make(map[int]bool)
	for {
		threads, err := ioutil.ReadDir(fmt.Sprintf("/proc/%d/task", pid))
		if err != nil {
			log.Error(err, "read failed", "pid", pid)
			return constructPartialProgram(pid, tidMap), errors.WithStack(err)
		}

		if len(threads) == len(tidMap) {
			break
		}

		for _, thread := range threads {
			tid64, err := strconv.ParseInt(thread.Name(), 10, 32)
			if err != nil {
				return constructPartialProgram(pid, tidMap), errors.WithStack(err)
			}
			tid := int(tid64)

			_, ok := tidMap[tid]
			if ok {
				continue
			}

			err = syscall.PtraceAttach(tid)
			if err != nil {
				log.Error(err, "attach failed", "tid", tid)
				return constructPartialProgram(pid, tidMap), errors.WithStack(err)
			}
			log.Info("attach successfully", "tid", tid)

			err = waitPid(tid)
			if err != nil {
				return constructPartialProgram(pid, tidMap), errors.WithStack(err)
			}
			tidMap[tid] = true
		}
	}

	var tids []int
	for key := range tidMap {
		tids = append(tids, key)
	}

	entries, err := mapreader.Read(pid)
	if err != nil {
		return constructPartialProgram(pid, tidMap), err
	}

	program := &TracedProgram{
		pid:        pid,
		tids:       tids,
		Entries:    entries,
		backupRegs: &syscall.PtraceRegs{},
		backupCode: make([]byte, ptrSize),
	}

	return program, nil
}

// Detach detaches from all threads of the processs
func (p *TracedProgram) Detach() error {
	for _, tid := range p.tids {
		err := syscall.PtraceDetach(tid)

		if err != nil {
			log.Error(err, "detach failed", "tid", tid)
			return errors.WithStack(err)
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
	binary.LittleEndian.PutUint16(ip, 0x050f) // The endianness is hard coded here
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
	return p.Syscall(9, 0, length, 7, 0x22, fd, 0)
}

// ReadSlice reads from addr and return a slice
func (p *TracedProgram) ReadSlice(addr uint64, size uint64) (*[]byte, error) {
	buffer := make([]byte, size)

	tmpBuffer := C.malloc(C.ulong(size))
	defer C.free(tmpBuffer)
	localIov := C.struct_iovec{
		iov_base: unsafe.Pointer(tmpBuffer),
		iov_len:  C.ulong(size),
	}

	remoteIovBuf := C.malloc(C.sizeof_struct_iovec)
	defer C.free(remoteIovBuf)
	remoteIov := (*C.struct_iovec)(remoteIovBuf)
	remoteIov.iov_base = C.Uint64ToPointer(C.ulong(addr))
	remoteIov.iov_len = C.ulong(size)

	ret, err := C.process_vm_readv(C.int(p.pid),
		(*C.struct_iovec)(unsafe.Pointer(&localIov)),
		1,
		remoteIov,
		1,
		0,
	)
	if ret == -1 {
		return nil, errors.WithStack(err)
	}
	// TODO: check size and warn
	C.memcpy(unsafe.Pointer(&buffer[0]), tmpBuffer, C.ulong(size))

	return &buffer, nil
}

// WriteSlice writes a buffer into addr
func (p *TracedProgram) WriteSlice(addr uint64, buffer []byte) error {
	size := len(buffer)

	tmpBuffer := C.malloc(C.ulong(size))
	defer C.free(tmpBuffer)
	C.memcpy(tmpBuffer, unsafe.Pointer(&buffer[0]), C.ulong(size))

	localIov := C.struct_iovec{
		iov_base: tmpBuffer,
		iov_len:  C.ulong(size),
	}

	remoteIovBuf := C.malloc(C.sizeof_struct_iovec)
	defer C.free(remoteIovBuf)
	remoteIov := (*C.struct_iovec)(remoteIovBuf)
	remoteIov.iov_base = C.Uint64ToPointer(C.ulong(addr))
	remoteIov.iov_len = C.ulong(size)

	ret, err := C.process_vm_writev(C.int(p.pid),
		(*C.struct_iovec)(unsafe.Pointer(&localIov)),
		1,
		remoteIov,
		1,
		0,
	)
	if ret == -1 {
		return errors.WithStack(err)
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
	elf, err := elf.NewFile(reader)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	symbols, err := elf.DynamicSymbols()
	if err != nil {
		return 0, errors.WithStack(err)
	}
	for _, symbol := range symbols {
		if symbol.Name == symbolName {
			offset := symbol.Value

			return entry.StartAddress + offset, nil
		}
	}
	return 0, fmt.Errorf("cannot find symbol")
}

// WriteUint64ToAddr writes uint64 to addr
func (p *TracedProgram) WriteUint64ToAddr(addr uint64, value uint64) error {
	valueSlice := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueSlice, uint64(value))
	err := p.WriteSlice(addr, valueSlice)
	if err != nil {
		return err
	}

	return nil
}

// JumpToFakeFunc writes jmp instruction to jump to fake function
func (p *TracedProgram) JumpToFakeFunc(originAddr uint64, targetAddr uint64, symbolName string) error {
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
