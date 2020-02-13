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
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/pingcap/chaos-mesh/test/pkg/timer"
)

func TestModifyTime(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Time Suit",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	By("change working directory")

	err := os.Chdir("../../")
	Expect(err).NotTo(HaveOccurred())

	close(done)
})

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("ModifyTime", func() {

	var t *timer.Timer

	BeforeEach(func() {
		var err error

		t, err = timer.StartTimer()
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		err := t.Stop()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Modify Time", func() {
		It("should move forward successfully", func() {
			Expect(t).NotTo(BeNil())

			now, err := t.GetTime()
			Expect(err).ShouldNot(HaveOccurred())

			sec := now.Unix()

			err = ModifyTime(t.Pid(), 10000, 100000000)
			Expect(err).ShouldNot(HaveOccurred())

			newTime, err := t.GetTime()
			Expect(err).ShouldNot(HaveOccurred())

			newSec := newTime.Unix()

			Expect(newSec - sec).Should(BeNumerically("==", 10000))
		})

		It("should move backward successfully", func() {
			Expect(t).NotTo(BeNil())

			now, err := t.GetTime()
			Expect(err).ShouldNot(HaveOccurred())

			sec := now.Unix()

			err = ModifyTime(t.Pid(), -10000, -100000000)
			Expect(err).ShouldNot(HaveOccurred())

			newTime, err := t.GetTime()
			Expect(err).ShouldNot(HaveOccurred())

			newSec := newTime.Unix()

			Expect(sec - newSec).Should(BeNumerically("==", 10000))
		})
	})
})
