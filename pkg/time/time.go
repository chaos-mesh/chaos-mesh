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

package time

import (
	"bytes"
	"errors"
	"runtime"

	"github.com/pingcap/chaos-mesh/pkg/mapreader"
	"github.com/pingcap/chaos-mesh/pkg/ptrace"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("time")

// TODO: support more cpu architecture
// TODO: auto generate these codes
var fakeImage = []byte{
	0xb8, 0xe4, 0x00, 0x00, 0x00, //mov    $0xe4,%eax
	0x0f, 0x05, //syscall
	0x85, 0xff, //test   %edi,%edi
	0x75, 0x67, //jne    <clock_gettime+0x72>
	0x48, 0x8d, 0x15, 0x61, 0x00, 0x00, 0x00, //lea    0x61(%rip),%rdx        # <TV_SEC_DELTA>
	0x4c, 0x8b, 0x46, 0x08, //mov    0x8(%rsi),%r8
	0x48, 0x8b, 0x0a, //mov    (%rdx),%rcx
	0x48, 0x8d, 0x15, 0x5b, 0x00, 0x00, 0x00, //lea    0x2fd8(%rip),%rdx        # <TV_NSEC_DELTA>
	0x48, 0x8b, 0x3a, //mov    (%rdx),%rdi
	0x4a, 0x8d, 0x14, 0x07, //lea    (%rdi,%r8,1),%rdx
	0x48, 0x81, 0xfa, 0x00, 0xca, 0x9a, 0x3b, //cmp    $1000000000,%rdx
	0x7e, 0x18, //jle    1048 <clock_gettime+0x48>
	0x48, 0x81, 0xef, 0x00, 0xca, 0x9a, 0x3b, //sub    $1000000000,%rdi
	0x48, 0x83, 0xc1, 0x01, //add    $0x1,%rcx
	0x49, 0x8d, 0x14, 0x38, //lea    (%r8,%rdi,1),%rdx
	0x48, 0x81, 0xfa, 0x00, 0xca, 0x9a, 0x3b, //cmp    $1000000000,%rdx
	0x7f, 0xe8, //jg     <clock_gettime+0x30>
	0x48, 0x85, 0xd2, //test   %rdx,%rdx
	0x79, 0x1e, //jns    <clock_gettime+0x6b>
	0x4a, 0x8d, 0xbc, 0x07, 0x00, 0xca, 0x9a, 0x3b, //lea    $1000000000(%rdi,%r8,1),%rdi
	0x0f, 0x1f, 0x00, //nopl   (%rax)
	0x48, 0x89, 0xfa, //mov    %rdi,%rdx
	0x48, 0x83, 0xe9, 0x01, //sub    $0x1,%rcx
	0x48, 0x81, 0xc7, 0x00, 0xca, 0x9a, 0x3b, //add    $1000000000,%rdi
	0x48, 0x85, 0xd2, //test   %rdx,%rdx
	0x78, 0xed, //js     <clock_gettime+0x58>
	0x48, 0x01, 0x0e, //add    %rcx,(%rsi)
	0x48, 0x89, 0x56, 0x08, //mov    %rdx,0x8(%rsi)
	0xc3, //retq
	// constant
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TV_SEC_DELTA
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TV_NSEC_DELTA
}

// ModifyTime modifies time of target process
func ModifyTime(pid int, deltaSec int64, deltaNsec int64) error {
	runtime.LockOSThread()

	program, err := ptrace.Trace(pid)
	if err != nil {
		return err
	}
	defer func() {
		err = program.Detach()
		if err != nil {
			log.Error(err, "fail to detach program", "pid", program.Pid())
		}

		runtime.UnlockOSThread()
	}()

	var vdsoEntry *mapreader.Entry
	// find injected image to avoid redundant inject (which will lead to memory leak)
	for index := range program.Entries {
		// reverse loop is faster
		e := program.Entries[len(program.Entries)-index-1]
		if e.Path == "[vdso]" {
			vdsoEntry = &e
			break
		}
	}
	if vdsoEntry == nil {
		return errors.New("cannot find [vdso] entry")
	}

	// minus tailing variable part
	constImageLen := len(fakeImage) - 16
	var fakeEntry *mapreader.Entry
	for _, e := range program.Entries {
		e := e

		image, err := program.ReadSlice(e.StartAddress, uint64(constImageLen))
		if err != nil {
			continue
		}

		if bytes.Equal(*image, fakeImage[0:constImageLen]) {
			fakeEntry = &e
			log.Info("found injected image", "addr", fakeEntry.StartAddress)
			break
		}
	}
	if fakeEntry == nil {
		fakeEntry, err = program.MmapSlice(fakeImage)
		if err != nil {
			return err
		}
	}
	fakeAddr := fakeEntry.StartAddress

	// 115 is the index of TV_SEC_DELTA in fakeImage
	err = program.WriteUint64ToAddr(fakeAddr+115, uint64(deltaSec))
	if err != nil {
		return err
	}

	// 123 is the index of TV_NSEC_DELTA in fakeImage
	err = program.WriteUint64ToAddr(fakeAddr+123, uint64(deltaNsec))
	if err != nil {
		return err
	}

	originAddr, err := program.FindSymbolInEntry("clock_gettime", vdsoEntry)
	if err != nil {
		return err
	}

	err = program.JumpToFakeFunc(originAddr, fakeAddr, "clock_gettime")
	if err != nil {
		return err
	}

	return nil
}
