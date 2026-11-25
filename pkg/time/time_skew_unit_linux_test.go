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
package time

import (
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
)

func TestModifyTime(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Time skew tests")
}

type clockImageCall struct {
	image     string
	operation string
	pid       int
	variables map[string]uint64
}

type recordedClockImage struct {
	name       string
	calls      *[]clockImageCall
	attachErr  error
	recoverErr error
}

func (image *recordedClockImage) AttachToProcess(pid int, variables map[string]uint64) error {
	*image.calls = append(*image.calls, clockImageCall{image.name, "attach", pid, variables})
	return image.attachErr
}

func (image *recordedClockImage) Recover(pid int, variables map[string]uint64) error {
	*image.calls = append(*image.calls, clockImageCall{image.name, "recover", pid, variables})
	return image.recoverErr
}

var _ = Describe("Dispatching time skew to the clock images", func() {
	const pid = tasks.SysPID(42)
	var (
		calls     []clockImageCall
		clock     *recordedClockImage
		timeOfDay *recordedClockImage
		skew      *Skew
	)
	failure := errors.New("clock image operation failed")
	clockVariables := map[string]uint64{
		"CLOCK_IDS_MASK": 5,
		"TV_SEC_DELTA":   0xfffffffffffffffe,
		"TV_NSEC_DELTA":  123,
	}
	timeOfDayVariables := map[string]uint64{"TV_SEC_DELTA": 0xfffffffffffffffe, "TV_NSEC_DELTA": 123}

	BeforeEach(func() {
		calls = nil
		clock = &recordedClockImage{name: "clock_gettime", calls: &calls}
		timeOfDay = &recordedClockImage{name: "gettimeofday", calls: &calls}
		skew = &Skew{SkewConfig: NewConfig(-2, 123, 5), clockGetTime: clock, getTimeOfDay: timeOfDay, logger: logr.Discard()}
	})

	It("passes the process ID and signed offset to both images, and the clock mask only to clock_gettime", func() {
		Expect(skew.Inject(pid)).To(Succeed())
		Expect(calls).To(Equal([]clockImageCall{
			{"clock_gettime", "attach", int(pid), clockVariables},
			{"gettimeofday", "attach", int(pid), timeOfDayVariables},
		}))
	})

	DescribeTable("preserves injection errors in the order the images are called",
		func(firstFails bool, expected []string) {
			if firstFails {
				clock.attachErr = failure
			} else {
				timeOfDay.attachErr = failure
			}
			Expect(skew.Inject(pid)).To(MatchError(failure))
			var images []string
			for _, call := range calls {
				images = append(images, call.image)
			}
			Expect(images).To(Equal(expected))
		},
		Entry("stops before the second attach if the first fails", true, []string{"clock_gettime"}),
		Entry("reports a failed second attach", false, []string{"clock_gettime", "gettimeofday"}),
	)

	It("rejects non-PID identifiers before accessing either image", func() {
		id := tasks.PodContainerName("app")
		Expect(skew.Inject(id)).To(MatchError(ContainSubstring("SysPID")))
		Expect(skew.Recover(id)).To(MatchError(ContainSubstring("SysPID")))
		Expect(calls).To(BeEmpty())
	})

	DescribeTable("attempts both recoveries and preserves their errors",
		func(clockErr, timeOfDayErr, expectedErr error) {
			clock.recoverErr = clockErr
			timeOfDay.recoverErr = timeOfDayErr

			err := skew.Recover(pid)
			if expectedErr == nil {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(MatchError(expectedErr))
				if clockErr != nil && timeOfDayErr != nil {
					Expect(err).To(MatchError(ContainSubstring(timeOfDayErr.Error())))
				}
			}
			Expect(calls).To(Equal([]clockImageCall{
				{"clock_gettime", "recover", int(pid), clockVariables},
				{"gettimeofday", "recover", int(pid), timeOfDayVariables},
			}))
		},
		Entry("success", nil, nil, nil),
		Entry("clock_gettime fails", failure, nil, failure),
		Entry("gettimeofday fails", nil, failure, failure),
		Entry("both fail", failure, errors.New("second recovery failed"), failure),
	)
})

var _ = Describe("Time skew task configuration", func() {
	It("combines overlapping experiments without changing the original configuration", func() {
		original := NewConfig(2, 300, 0b001)
		other := NewConfig(-4, 500, 0b101)
		cloned := original.DeepCopy().(*Config)
		Expect(cloned.Merge(&other)).To(Succeed())
		Expect(*cloned).To(Equal(NewConfig(-2, 800, 0b101)))
		Expect(original).To(Equal(NewConfig(2, 300, 0b001)))
	})

	It("updates the existing handler's configuration and copies it when forking", func() {
		config := NewConfig(2, 30, 5)
		processes := tasks.NewPodProcessMap()
		handler, err := config.New(ConfigCreatorParas{Logger: logr.Discard(), Config: config, PodProcessMap: &processes})
		Expect(err).NotTo(HaveOccurred())
		podHandler := handler.(*tasks.PodHandler)
		group := podHandler.SubProcess.(*tasks.ProcessGroupHandler)
		skew := group.LeaderProcess.(*Skew)
		Expect(skew.SkewConfig).To(Equal(config))

		next := NewConfig(-1, 0, 2)
		Expect(next.Assign(handler)).To(Succeed())
		Expect(skew.SkewConfig).To(Equal(next))
		fork, err := skew.Fork()
		Expect(err).NotTo(HaveOccurred())
		Expect(fork.(*Skew).SkewConfig).To(Equal(next))
		Expect(fork).NotTo(BeIdenticalTo(skew))

		last := NewConfig(1, 20, 4)
		Expect(last.Assign(handler)).To(Succeed())
		Expect(fork.(*Skew).SkewConfig).To(Equal(next))
		Expect(skew.Assign(fork)).To(Succeed())
		Expect(fork.(*Skew).SkewConfig).To(Equal(last))
	})
})
