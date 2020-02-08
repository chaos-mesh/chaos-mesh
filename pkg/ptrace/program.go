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

// #include <sys/wait.h>
import (
	"encoding/binary"
	"fmt"
	"syscall"
	"C"
)

const waitPidErrorMessage = "waitpid ret value: %d, status: %d"

const ptrSize = 4 << uintptr(^uintptr(0)>>63) // Here is a trick to get pointer size in bytes

type TracedProgram struct {
	pid int

	backupRegs syscall.PtraceRegs
	backupCode []byte
}

func waitPid(pid int) error {
	status := 0

	ret := int(C.waitpid(pid, &status, 0))
	if ret == 0 {
		return nil
	} else {
		return fmt.Errorf(waitPidErrorMessage, ret, status)
	}
}

func Trace(pid int) (error, *TracedProgram) {
	err := syscall.PtraceAttach(pid)
	if err != nil {
		return err, nil
	}

	err = waitPid(pid)
	if err != nil {
		return err, nil
	}

	return nil, &TracedProgram {
		pid: pid,
		backupRegs: syscall.PtraceRegs{},
		backupCode: make([]byte, ptrSize),
	}
}

func (p *TracedProgram) Cont() error {
	return syscall.PtraceCont(p.pid, 0)
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

func (p *TracedProgram) Step()  error {
	err := syscall.PtraceSingleStep(p.pid)
	if err != nil {
		return err
	}

	return p.Wait()
}

func (p *TracedProgram) Syscall(number uint64, args ...uint64) (error, uint64) {
	err := p.Protect()
	if err != nil {
		return err, 0
	}

	var regs syscall.PtraceRegs

	err = syscall.PtraceGetRegs(p.pid, &regs)
	if err != nil {
		return err, 0
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
			return fmt.Errorf("too many arguments for a syscall"), 0
		}
	}
	err = syscall.PtraceSetRegs(p.pid, &regs)
	if err != nil {
		return err, 0
	}

	ip := make([]byte, ptrSize)
	binary.LittleEndian.PutUint16(ip, 0x050f) // The endianness is hard coded here
	_, err = syscall.PtracePokeData(p.pid, uintptr(p.backupRegs.Rip), ip)
	if err != nil {
		return err, 0
	}

	err = p.Step()
	if err != nil {
		return err, 0
	}

	err = syscall.PtraceGetRegs(p.pid, &regs)
	if err != nil {
		return err, 0
	}

	return p.Restore(), regs.Rax
}

func (p *TracedProgram) Mmap(length uint64, fd uint64) (error, uint64) {
	return p.Syscall(9, 0, length, 7, 0x22, fd, 0)
}