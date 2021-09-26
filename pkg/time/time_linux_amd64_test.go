// Copyright 2020 Chaos Mesh Authors.
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

	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	"github.com/chaos-mesh/chaos-mesh/test/pkg/timer"
)

func TestModifyTime(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Time Suit",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	By("change working directory")

	err := os.Chdir("../../")
	Expect(err).NotTo(HaveOccurred())

	By("register logger")
	zapLog, err := zap.NewDevelopment()
	Expect(err).NotTo(HaveOccurred())
	log := zapr.NewLogger(zapLog)
	RegisterLogger(log)

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
			Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

			sec := now.Unix()

			err = ModifyTime(t.Pid(), 10000, 0, 1)
			Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

			newTime, err := t.GetTime()
			Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

			newSec := newTime.Unix()

			Expect(newSec-sec).Should(BeNumerically(">=", 10000), "sec %d newSec %d", sec, newSec)
		})

		It("should move backward successfully", func() {
			Expect(t).NotTo(BeNil())

			now, err := t.GetTime()
			Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

			sec := now.Unix()

			err = ModifyTime(t.Pid(), -10000, 0, 1)
			Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

			newTime, err := t.GetTime()
			Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

			newSec := newTime.Unix()

			Expect(10000-(sec-newSec)).Should(BeNumerically("<=", 1), "sec %d newSec %d", sec, newSec)
		})

		It("should handle nsec overflow", func() {
			Expect(t).NotTo(BeNil())

			now, err := t.GetTime()
			Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

			sec := now.Unix()

			err = ModifyTime(t.Pid(), 0, 1000000000, 1)
			Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

			newTime, err := t.GetTime()
			Expect(err).ShouldNot(HaveOccurred(), "error: %+v", err)

			newSec := newTime.Unix()

			Expect(newSec-sec).Should(BeNumerically(">=", 1), "sec %d newSec %d", sec, newSec)
		})
	})
})
