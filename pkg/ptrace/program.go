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
	"runtime"
	"syscall"
	"unsafe"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/pingcap/chaos-mesh/pkg/mapreader"
)

var log = ctrl.Log.WithName("ptrace")

const waitPidErrorMessage = "waitpid ret value: %d"

const ptrSize = 4 << uintptr(^uintptr(0)>>63) // Here is a trick to get pointer size in bytes

type TracedProgram struct {
	pid     int
	Entries *[]mapreader.Entry

	backupRegs syscall.PtraceRegs
	backupCode []byte
}

func (p *TracedProgram) Pid() int {
	return p.pid
}

func waitPid(pid int) error {
	ret := waitpid(pid)
	if ret != -1 {
		return nil
	} else {
		return fmt.Errorf(waitPidErrorMessage, ret)
	}
}

func Trace(pid int) (*TracedProgram, error) {
	err := syscall.PtraceAttach(pid)
	if err != nil {
		return nil, err
	}

	err = waitPid(pid)
	if err != nil {
		return nil, err
	}

	err, entries := mapreader.Read(pid)
	if err != nil {
		return nil, err
	}

	program := &TracedProgram{
		pid:        pid,
		Entries:    entries,
		backupRegs: syscall.PtraceRegs{},
		backupCode: make([]byte, ptrSize),
	}
	runtime.SetFinalizer(program, func(p *TracedProgram) {
		_ = p.Detach()
	})

	return program, nil
}

func (p *TracedProgram) Cont() error {
	return syscall.PtraceCont(p.pid, 0)
}

func (p *TracedProgram) Detach() error {
	err := syscall.PtraceDetach(p.pid)

	if err != nil {
		return err
	}

	log.Info("Successfully detach and rerun process", "pid", p.pid)
	return nil
}

// Protect will backup regs and rip into fields
func (p *TracedProgram) Protect() error {
	err := syscall.PtraceGetRegs(p.pid, &p.backupRegs)
	if err != nil {
		return err
	}

	_, err = syscall.PtracePeekData(p.pid, uintptr(p.backupRegs.Rip), p.backupCode)
	if err != nil {
		return err
	}

	return nil
}

// Restore will restore regs and rip from fields
func (p *TracedProgram) Restore() error {
	err := syscall.PtraceSetRegs(p.pid, &p.backupRegs)
	if err != nil {
		return err
	}

	_, err = syscall.PtracePokeData(p.pid, uintptr(p.backupRegs.Rip), p.backupCode)
	if err != nil {
		return err
	}

	return nil
}

func (p *TracedProgram) Wait() error {
	return waitPid(p.pid)
}

func (p *TracedProgram) Step() error {
	err := syscall.PtraceSingleStep(p.pid)
	if err != nil {
		return err
	}

	return p.Wait()
}

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

func (p *TracedProgram) Mmap(length uint64, fd uint64) (uint64, error) {
	return p.Syscall(9, 0, length, 7, 0x22, fd, 0)
}

func (p *TracedProgram) ReadSlice(addr uint64, size uint64) (*[]byte, error) {
	buffer := make([]byte, size)

	tmpBuffer := C.malloc(C.ulong(size))
	defer C.free(tmpBuffer)
	localIov := C.struct_iovec{
		iov_base: unsafe.Pointer(tmpBuffer),
		iov_len:  C.ulong(size),
	}
	remoteIov := C.struct_iovec{
		iov_base: C.Uint64ToPointer(C.ulong(addr)),
		iov_len:  C.ulong(size),
	}

	ret, err := C.process_vm_readv(C.int(p.pid),
		(*C.struct_iovec)(unsafe.Pointer(&localIov)),
		1,
		(*C.struct_iovec)(unsafe.Pointer(&remoteIov)),
		1,
		0,
	)
	if ret == -1 {
		return nil, err
	}
	// TODO: check size and warn
	C.memcpy(unsafe.Pointer(&buffer[0]), tmpBuffer, C.ulong(size))

	return &buffer, nil
}

func (p *TracedProgram) WriteSlice(addr uint64, buffer []byte) error {
	size := len(buffer)

	tmpBuffer := C.malloc(C.ulong(size))
	defer C.free(tmpBuffer)
	C.memcpy(tmpBuffer, unsafe.Pointer(&buffer[0]), C.ulong(size))

	localIov := C.struct_iovec{
		iov_base: tmpBuffer,
		iov_len:  C.ulong(size),
	}
	remoteIov := C.struct_iovec{
		iov_base: C.Uint64ToPointer(C.ulong(addr)),
		iov_len:  C.ulong(size),
	}

	ret, err := C.process_vm_writev(C.int(p.pid),
		(*C.struct_iovec)(unsafe.Pointer(&localIov)),
		1,
		(*C.struct_iovec)(unsafe.Pointer(&remoteIov)),
		1,
		0,
	)
	if ret == -1 {
		return err
	}
	// TODO: check size and warn

	return nil
}

func (p *TracedProgram) PtraceWriteSlice(addr uint64, buffer []byte) error {
	wroteSize := 0

	for wroteSize < len(buffer) {
		_, err := syscall.PtracePokeData(p.pid, uintptr(addr+uint64(wroteSize)), buffer[wroteSize:wroteSize+ptrSize])
		if err != nil {
			return err
		}

		wroteSize += ptrSize
	}

	return nil
}

func (p *TracedProgram) GetLibBuffer(entry *mapreader.Entry) (*[]byte, error) {
	if entry.PaddingSize > 0 {
		return nil, fmt.Errorf("entry with padding size is not supported")
	}

	size := entry.EndAddress - entry.StatAddress

	return p.ReadSlice(entry.StatAddress, size)
}

func (p *TracedProgram) MmapSlice(slice []byte) (*mapreader.Entry, error) {
	size := uint64(len(slice))

	addr, err := p.Mmap(size, 0)
	if err != nil {
		return nil, err
	}

	err = p.WriteSlice(addr, slice)
	if err != nil {
		return nil, err
	}

	return &mapreader.Entry{
		StatAddress: addr,
		EndAddress:  addr + size,
		Privilege:   "rwxp",
		PaddingSize: 0,
		Path:        "",
	}, nil
}

func (p *TracedProgram) FindSymbolInEntry(symbolName string, entry *mapreader.Entry) (uint64, error) {
	libBuffer, err := p.GetLibBuffer(entry)
	if err != nil {
		return 0, err
	}

	reader := bytes.NewReader(*libBuffer)
	elf, err := elf.NewFile(reader)
	if err != nil {
		return 0, err
	}

	symbols, err := elf.DynamicSymbols()
	if err != nil {
		return 0, err
	}
	for _, symbol := range symbols {
		if symbol.Name == symbolName {
			offset := symbol.Value

			return entry.StatAddress + offset, nil
		}
	}
	return 0, fmt.Errorf("cannot find symbol")
}

func (p *TracedProgram) WriteUint64ToAddr(addr uint64, value uint64) error {
	valueSlice := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueSlice, uint64(value))
	err := p.WriteSlice(addr, valueSlice)
	if err != nil {
		return err
	}

	return nil
}

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
