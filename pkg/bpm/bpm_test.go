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
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shirou/gopsutil/process"
)

func RandomeIdentifier() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, 10)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func WaitProcess(m *BackgroundProcessManager, cmd *ManagedProcess, exceedTime time.Duration) {
	pid := cmd.Process.Pid
	procState, err := process.NewProcess(int32(pid))
	Expect(err).To(BeNil())
	ct, err := procState.CreateTime()
	Expect(err).To(BeNil())
	pair := ProcessPair{
		Pid:        pid,
		CreateTime: ct,
	}
	channel, ok := m.deathSig.Load(pair)
	Expect(ok).To(BeTrue())
	deathChannel := channel.(chan bool)

	timeExceed := false
	select {
	case <-deathChannel:
	case <-time.Tick(exceedTime):
		timeExceed = true
	}
	Expect(timeExceed).To(BeFalse())
}

var _ = Describe("background process manager", func() {
	m := NewBackgroundProcessManager()

	Context("normally exited process", func() {
		It("should work", func() {
			cmd := DefaultProcessBuilder("sleep", "2").Build()
			_, err := m.StartProcess(cmd)
			Expect(err).To(BeNil())

			WaitProcess(&m, cmd, time.Second*5)
		})

		It("processes with the same identifier", func() {
			identifier := RandomeIdentifier()

			cmd := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build()
			_, err := m.StartProcess(cmd)
			Expect(err).To(BeNil())

			startTime := time.Now()
			cmd2 := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build()
			_, err = m.StartProcess(cmd2)
			costedTime := time.Since(startTime)
			Expect(err).To(BeNil())
			Expect(costedTime.Seconds()).Should(BeNumerically(">", 1.9))

			_, err = process.NewProcess(int32(cmd.Process.Pid))
			Expect(err).NotTo(BeNil()) // The first process should have exited

			WaitProcess(&m, cmd2, time.Second*5)
		})
	})

	Context("kill process", func() {
		It("should work", func() {
			cmd := DefaultProcessBuilder("sleep", "2").Build()
			_, err := m.StartProcess(cmd)
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
			identifier := RandomeIdentifier()

			cmd := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build()
			_, err := m.StartProcess(cmd)
			Expect(err).To(BeNil())

			pid := cmd.Process.Pid
			procState, err := process.NewProcess(int32(pid))
			Expect(err).To(BeNil())
			ct, err := procState.CreateTime()
			Expect(err).To(BeNil())

			cmd2 := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build()

			go func() {
				time.Sleep(time.Second)

				err = m.KillBackgroundProcess(context.Background(), pid, ct)
				Expect(err).To(BeNil())
			}()

			startTime := time.Now()
			_, err = m.StartProcess(cmd2)
			costedTime := time.Since(startTime)
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
