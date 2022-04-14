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
//go:build cgo

package ptrace

import (
	"encoding/binary"
	"syscall"

	"github.com/pkg/errors"
)

var endian = binary.LittleEndian

const syscallInstrSize = 2

const nrProcessVMReadv = 310
const nrProcessVMWritev = 311

func getIp(regs *syscall.PtraceRegs) uintptr {
	return uintptr(regs.Rip)
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
			return 0, errors.New("too many arguments for a syscall")
		}
	}
	err = syscall.PtraceSetRegs(p.pid, &regs)
	if err != nil {
		return 0, err
	}

	ip := make([]byte, syscallInstrSize)

	// We only support x86-64 platform now, so using hard coded `LittleEndian` here is ok.
	binary.LittleEndian.PutUint16(ip, 0x050f)
	_, err = syscall.PtracePokeData(p.pid, getIp(p.backupRegs), ip)
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

// JumpToFakeFunc writes jmp instruction to jump to fake function
func (p *TracedProgram) JumpToFakeFunc(originAddr uint64, targetAddr uint64) error {
	instructions := make([]byte, 16)

	// mov rax, targetAddr;
	// jmp rax ;
	instructions[0] = 0x48
	instructions[1] = 0xb8
	endian.PutUint64(instructions[2:10], targetAddr)
	instructions[10] = 0xff
	instructions[11] = 0xe0

	return p.PtraceWriteSlice(originAddr, instructions)
}
