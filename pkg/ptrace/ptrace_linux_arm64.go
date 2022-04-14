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

const syscallInstrSize = 4

const nrProcessVMReadv = 270
const nrProcessVMWritev = 271

func getIp(regs *syscall.PtraceRegs) uintptr {
	return uintptr(regs.Pc)
}

// Syscall runs a syscall at main thread of process
func (p *TracedProgram) Syscall(number uint64, args ...uint64) (uint64, error) {
	err := p.Protect()
	if err != nil {
		return 0, err
	}

	var regs syscall.PtraceRegs

	err = getRegs(p.pid, &regs)
	if err != nil {
		return 0, err
	}
	regs.Regs[8] = number
	for index, arg := range args {
		// All these registers are hard coded for x86 platform
		if index > 6 {
			return 0, errors.New("too many arguments for a syscall")
		} else {
			regs.Regs[index] = arg
		}
	}
	err = setRegs(p.pid, &regs)
	if err != nil {
		return 0, err
	}

	ip := make([]byte, syscallInstrSize)

	// most aarch64 devices are big endian
	// 0xd4000001 is `svc #0` to call the system call
	endian.PutUint32(ip, 0xd4000001)
	_, err = syscall.PtracePokeData(p.pid, getIp(p.backupRegs), ip)
	if err != nil {
		return 0, err
	}

	err = p.Step()
	if err != nil {
		return 0, err
	}

	err = getRegs(p.pid, &regs)
	if err != nil {
		return 0, err
	}

	return regs.Regs[0], p.Restore()
}

// JumpToFakeFunc writes jmp instruction to jump to fake function
func (p *TracedProgram) JumpToFakeFunc(originAddr uint64, targetAddr uint64) error {
	instructions := make([]byte, 16)

	// LDR x9, #8
	// BR x9
	// targetAddr
	endian.PutUint32(instructions[0:], 0x58000049)
	endian.PutUint32(instructions[4:], 0xD61F0120)

	endian.PutUint64(instructions[8:], targetAddr)

	return p.PtraceWriteSlice(originAddr, instructions)
}