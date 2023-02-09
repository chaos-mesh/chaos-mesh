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
	"math/rand"
	"os"
	"os/exec"
	"testing"
	"time"
	"unsafe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/test/pkg/timer"
)

func TestPTrace(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "PTrace Suit")
}

var _ = BeforeSuite(func(done Done) {
	rand.Seed(GinkgoRandomSeed())

	By("change working directory")

	err := os.Chdir("../../")
	Expect(err).NotTo(HaveOccurred())

	By("register logger")

	close(done)
})

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("PTrace", func() {

	logger, err := log.NewDefaultZapLogger()
	Expect(err).NotTo(HaveOccurred())

	var t *timer.Timer
	var program *TracedProgram

	BeforeEach(func() {
		var err error

		t, err = timer.StartTimer()
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

		time.Sleep(time.Millisecond)

		program, err = Trace(t.Pid(), logger)
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)
	})

	AfterEach(func() {
		err := program.Detach()
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

		err = t.Stop()
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)
	})

	It("should mmap slice successfully", func() {
		Expect(program.Pid()).Should(Equal(t.Pid()))

		helloWorld := []byte("Hello World")
		entry, err := program.MmapSlice(helloWorld)
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

		readBuf, err := program.ReadSlice(entry.StartAddress, uint64(len(helloWorld)))
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

		Expect(*readBuf).Should(Equal(helloWorld))
	})

	It("double trace should get error", func() {
		_, err := Trace(t.Pid(), logger)
		Expect(err).Should(HaveOccurred())
	})

	It("should ptrace write slice successfully", func() {
		helloWorld := []byte("Hello World")
		addr, err := program.Mmap(uint64(len(helloWorld)), 0)
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

		err = program.PtraceWriteSlice(addr, helloWorld)
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v, addr: %d", err, addr)

		readBuf, err := program.ReadSlice(addr, uint64(len(helloWorld)))
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v, addr: %d", err, addr)

		Expect(*readBuf).Should(Equal(helloWorld))
	})

	It("should write uint64 successfully", func() {
		number := rand.Uint64()
		size := uint64(unsafe.Sizeof(number))
		expectBuf := make([]byte, size)
		binary.LittleEndian.PutUint64(expectBuf, number)

		addr, err := program.Mmap(size, 0)
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

		err = program.WriteUint64ToAddr(addr, number)
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v, addr: %d", err, addr)

		readBuf, err := program.ReadSlice(addr, size)
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v, addr: %d", err, addr)

		Expect(*readBuf).Should(Equal(expectBuf))
	})

	It("should be able to detach and reattach", func() {
		err := program.Detach()
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

		program, err = Trace(t.Pid(), logger)
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)
	})

	It("should be able to attach and detach multithread program", func() {
		p := exec.Command("./bin/test/multithread_tracee")
		err := p.Start()
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

		time.Sleep(time.Millisecond)

		pid := p.Process.Pid
		program, err := Trace(pid, logger)
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

		err = program.Detach()
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

		err = p.Process.Kill()
		Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)
	})
})
