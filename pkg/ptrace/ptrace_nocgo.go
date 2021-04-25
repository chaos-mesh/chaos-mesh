// Copyright 2021 Chaos Mesh Authors.
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

// +build !cgo

package ptrace

import (
	"github.com/go-logr/logr"

	"github.com/chaos-mesh/chaos-mesh/pkg/mapreader"
)

// RegisterLogger registers a logger on ptrace pkg
func RegisterLogger(logger logr.Logger) {
	panic("unimplemented")
}

// TracedProgram is a program traced by ptrace
type TracedProgram struct {
	Entries []mapreader.Entry
}

// Pid return the pid of traced program
func (p *TracedProgram) Pid() int {
	panic("unimplemented")
}

// Trace ptrace all threads of a process
func Trace(pid int) (*TracedProgram, error) {
	panic("unimplemented")
}

// Detach detaches from all threads of the processes
func (p *TracedProgram) Detach() error {
	panic("unimplemented")
}

// Protect will backup regs and rip into fields
func (p *TracedProgram) Protect() error {
	panic("unimplemented")
}

// Restore will restore regs and rip from fields
func (p *TracedProgram) Restore() error {
	panic("unimplemented")
}

// Wait waits until the process stops
func (p *TracedProgram) Wait() error {
	panic("unimplemented")
}

// Step moves one step forward
func (p *TracedProgram) Step() error {
	panic("unimplemented")
}

// Syscall runs a syscall at main thread of process
func (p *TracedProgram) Syscall(number uint64, args ...uint64) (uint64, error) {
	panic("unimplemented")
}

// Mmap runs mmap syscall
func (p *TracedProgram) Mmap(length uint64, fd uint64) (uint64, error) {
	panic("unimplemented")
}

// ReadSlice reads from addr and return a slice
func (p *TracedProgram) ReadSlice(addr uint64, size uint64) (*[]byte, error) {
	panic("unimplemented")
}

// WriteSlice writes a buffer into addr
func (p *TracedProgram) WriteSlice(addr uint64, buffer []byte) error {
	panic("unimplemented")
}

// PtraceWriteSlice uses ptrace rather than process_vm_write to write a buffer into addr
func (p *TracedProgram) PtraceWriteSlice(addr uint64, buffer []byte) error {
	panic("unimplemented")
}

// GetLibBuffer reads an entry
func (p *TracedProgram) GetLibBuffer(entry *mapreader.Entry) (*[]byte, error) {
	panic("unimplemented")
}

// MmapSlice mmaps a slice and return it's addr
func (p *TracedProgram) MmapSlice(slice []byte) (*mapreader.Entry, error) {
	panic("unimplemented")
}

// FindSymbolInEntry finds symbol in entry through parsing elf
func (p *TracedProgram) FindSymbolInEntry(symbolName string, entry *mapreader.Entry) (uint64, error) {
	panic("unimplemented")
}

// WriteUint64ToAddr writes uint64 to addr
func (p *TracedProgram) WriteUint64ToAddr(addr uint64, value uint64) error {
	panic("unimplemented")
}

// JumpToFakeFunc writes jmp instruction to jump to fake function
func (p *TracedProgram) JumpToFakeFunc(originAddr uint64, targetAddr uint64) error {
	panic("unimplemented")
}
