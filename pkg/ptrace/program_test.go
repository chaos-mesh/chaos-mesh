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

import (
	"encoding/binary"
	"math/rand"
	"os"
	"testing"
	"unsafe"

	"github.com/pingcap/chaos-mesh/test/pkg/timer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestPTrace(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"PTrace Suit",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	rand.Seed(GinkgoRandomSeed())

	By("change working directory")

	err := os.Chdir("../../")
	Expect(err).NotTo(HaveOccurred())

	close(done)
})

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("PTrace", func() {

	var t *timer.Timer
	var program *TracedProgram

	BeforeEach(func() {
		var err error

		t, err = timer.StartTimer()
		Expect(err).ShouldNot(HaveOccurred())

		program, err = Trace(t.Pid())
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		err := program.Detach()
		Expect(err).ShouldNot(HaveOccurred())

		err = t.Stop()
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should mmap slice successfully", func() {
		Expect(program.Pid()).Should(Equal(t.Pid()))

		helloWorld := []byte("Hello World")
		entry, err := program.MmapSlice(helloWorld)
		Expect(err).ShouldNot(HaveOccurred())

		readBuf, err := program.ReadSlice(entry.StartAddress, uint64(len(helloWorld)))
		Expect(err).ShouldNot(HaveOccurred())

		Expect(*readBuf).Should(Equal(helloWorld))
	})

	It("double trace should get error", func() {
		_, err := Trace(t.Pid())
		Expect(err).Should(HaveOccurred())
	})

	It("should ptrace write slice successfully", func() {
		helloWorld := []byte("Hello World")
		addr, err := program.Mmap(uint64(len(helloWorld)), 0)
		Expect(err).ShouldNot(HaveOccurred())

		err = program.PtraceWriteSlice(addr, helloWorld)
		Expect(err).ShouldNot(HaveOccurred())

		readBuf, err := program.ReadSlice(addr, uint64(len(helloWorld)))
		Expect(err).ShouldNot(HaveOccurred())

		Expect(*readBuf).Should(Equal(helloWorld))
	})

	It("should write uint64 successfully", func() {
		number := rand.Uint64()
		size := uint64(unsafe.Sizeof(number))
		expectBuf := make([]byte, size)
		binary.LittleEndian.PutUint64(expectBuf, number)

		addr, err := program.Mmap(size, 0)
		Expect(err).ShouldNot(HaveOccurred())

		err = program.WriteUint64ToAddr(addr, number)
		Expect(err).ShouldNot(HaveOccurred())

		readBuf, err := program.ReadSlice(addr, size)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(*readBuf).Should(Equal(expectBuf))
	})
})
