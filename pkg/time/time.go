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
	0x0f, 0x05, // syscall
	0x85, 0xff, //test   %edi,%edi
	0x75, 0x20, //jne
	0x48, 0x8d, 0x15, 0x1a, 0x00, 0x00, 0x00, //lea    0x1a(%rip),%rdx  # <TV_SEC_DELTA>
	0xf3, 0x0f, 0x6f, 0x0e, //movdqu (%rsi),%xmm1
	0xf3, 0x0f, 0x7e, 0x02, //movq   (%rdx),%xmm0
	0x48, 0x8d, 0x15, 0x13, 0x00, 0x00, 0x00, //lea    0x13(%rip),%rdx  # <TV_NSEC_DELTA>
	0x0f, 0x16, 0x02, //movhps (%rdx),%xmm0
	0x66, 0x0f, 0xd4, 0xc1, //paddq  %xmm1,%xmm0
	0x0f, 0x11, 0x06, //movups %xmm0,(%rsi)
	0xc3, //retq
	// constant
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TV_SEC_DELTA
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //TV_NSEC_DELTA
}

// ModifyTime modifies time of target process
func ModifyTime(pid int, deltaSec int64, deltaNsec int64) error {
	runtime.LockOSThread()

	program, err := ptrace.Trace(pid)
	defer func() {
		err = program.Detach()
		if err != nil {
			log.Error(err, "fail to detach program", "pid", program.Pid())
		}

		runtime.UnlockOSThread()
	}()
	if err != nil {
		return err
	}

	var vdsoEntry *mapreader.Entry
	for _, e := range *program.Entries {
		e := e
		if e.Path == "[vdso]" {
			vdsoEntry = &e
		}
	}
	if vdsoEntry == nil {
		return errors.New("cannot find [vdso] entry")
	}

	constImageLen := len(fakeImage) - 16
	var fakeEntry *mapreader.Entry
	for _, e := range *program.Entries {
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

	err = program.WriteUint64ToAddr(fakeAddr+44, uint64(deltaSec))
	if err != nil {
		return err
	}

	err = program.WriteUint64ToAddr(fakeAddr+52, uint64(deltaNsec))
	if err != nil {
		return err
	}

	originAddr, err := program.FindSymbolInEntry("clock_gettime", vdsoEntry)
	if err != nil {
		return err
	}

	_, err = program.Mprotect(vdsoEntry.StartAddress, vdsoEntry.EndAddress-vdsoEntry.StartAddress, 7)
	if err != nil {
		return err
	}

	err = program.JumpToFakeFunc(originAddr, fakeAddr, "clock_gettime")
	if err != nil {
		return err
	}

	_, err = program.Mprotect(vdsoEntry.StartAddress, vdsoEntry.EndAddress-vdsoEntry.StartAddress, 5)
	if err != nil {
		return err
	}

	return nil
}
