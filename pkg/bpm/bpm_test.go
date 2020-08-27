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

package bpm

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shirou/gopsutil/process"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestBpm(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Background Process Manager Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = Describe("background process manager", func() {
	m := NewBackgroundProcessManager()

	Context("normally exited process", func() {
		It("should work", func() {
			cmd := DefaultProcessBuilder("sleep", "2").Build()
			err := m.StartProcess(cmd)
			Expect(err).To(BeNil())

			err = cmd.Wait()
			Expect(err).To(BeNil())
		})

		It("processes with the same identifier", func() {
			cmd := DefaultProcessBuilder("sleep", "2").
				SetIdentifier("nep").
				Build()
			err := m.StartProcess(cmd)
			Expect(err).To(BeNil())

			startTime := time.Now()
			cmd2 := DefaultProcessBuilder("sleep", "2").
				SetIdentifier("nep").
				Build()
			err = m.StartProcess(cmd2)
			costedTime := time.Now().Sub(startTime)
			Expect(err).To(BeNil())
			Expect(costedTime.Seconds()).Should(BeNumerically(">", 1.9))

			_, err = process.NewProcess(int32(cmd.Process.Pid))
			Expect(err).NotTo(BeNil()) // The first process should have exited

			err = cmd2.Wait()
			Expect(err).To(BeNil())
		})
	})

	Context("kill process", func() {
		It("should work", func() {
			cmd := DefaultProcessBuilder("sleep", "2").Build()
			err := m.StartProcess(cmd)
			Expect(err).To(BeNil())

			pid := cmd.Process.Pid
			procState, err := process.NewProcess(int32(pid))
			Expect(err).To(BeNil())
			ct, err := procState.CreateTime()
			Expect(err).To(BeNil())

			err = m.KillBackgroundProcess(context.Background(), pid, ct)
			Expect(err).To(BeNil())

			procState, err = process.NewProcess(int32(pid))
			Expect(err).NotTo(BeNil())
		})

		It("process with the same identifier", func() {
			cmd := DefaultProcessBuilder("sleep", "2").
				SetIdentifier("kp").
				Build()
			err := m.StartProcess(cmd)
			Expect(err).To(BeNil())

			pid := cmd.Process.Pid
			procState, err := process.NewProcess(int32(pid))
			Expect(err).To(BeNil())
			ct, err := procState.CreateTime()
			Expect(err).To(BeNil())

			cmd2 := DefaultProcessBuilder("sleep", "2").
				SetIdentifier("kp").
				Build()

			go func() {
				time.Sleep(time.Second)

				err = m.KillBackgroundProcess(context.Background(), pid, ct)
				Expect(err).To(BeNil())
			}()

			startTime := time.Now()
			err = m.StartProcess(cmd2)
			costedTime := time.Now().Sub(startTime)
			Expect(err).To(BeNil())
			Expect(costedTime.Seconds()).Should(And(BeNumerically("<", 2), BeNumerically(">", 1)))

			pid2 := cmd2.Process.Pid
			procState2, err := process.NewProcess(int32(pid2))
			Expect(err).To(BeNil())
			ct2, err := procState2.CreateTime()
			Expect(err).To(BeNil())

			err = m.KillBackgroundProcess(context.Background(), pid2, ct2)
			Expect(err).To(BeNil())
		})
	})
})
