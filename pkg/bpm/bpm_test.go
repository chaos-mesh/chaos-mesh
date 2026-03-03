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

package bpm

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/log"
)

func RandomeIdentifier() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, 10)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func WaitProcess(m *BackgroundProcessManager, proc *Process, exceedTime time.Duration) {
	timeExceed := false
	select {
	case <-proc.Stopped():
	case <-time.Tick(exceedTime):
		timeExceed = true
	}
	Expect(timeExceed).To(BeFalse())
}

var _ = Describe("background process manager", func() {
	log, err := log.NewDefaultZapLogger()
	Expect(err).To(BeNil())
	m := StartBackgroundProcessManager(nil, log)

	Context("normally exited process", func() {
		It("should work", func() {
			cmd := DefaultProcessBuilder("sleep", "2").Build(context.Background())
			p, err := m.StartProcess(context.Background(), cmd)
			Expect(err).To(BeNil())

			WaitProcess(m, p, time.Second*3)
		})

		It("processes with the same identifier", func() {
			identifier := RandomeIdentifier()

			cmd := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build(context.Background())
			p1, err := m.StartProcess(context.Background(), cmd)
			Expect(err).To(BeNil())

			// get error
			cmd2 := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build(context.Background())
			_, err = m.StartProcess(context.Background(), cmd2)
			Expect(err).NotTo(BeNil())
			Expect(strings.Contains(err.Error(), fmt.Sprintf("process with identifier %s is running", identifier))).To(BeTrue())

			WaitProcess(m, p1, time.Second*3)
			cmd3 := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build(context.TODO())
			p3, err := m.StartProcess(context.Background(), cmd3)
			Expect(err).To(BeNil())

			WaitProcess(m, p3, time.Second*3)
		})
	})

	Context("kill process", func() {
		It("should work", func() {
			cmd := DefaultProcessBuilder("sleep", "2").Build(context.Background())
			p, err := m.StartProcess(context.Background(), cmd)
			Expect(err).To(BeNil())

			err = m.KillBackgroundProcess(context.Background(), p.Uid)
			Expect(err).To(BeNil())

			WaitProcess(m, p, time.Second*0)
		})

		It("process with the same identifier", func() {
			identifier := RandomeIdentifier()

			cmd := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build(context.Background())
			p1, err := m.StartProcess(context.Background(), cmd)
			Expect(err).To(BeNil())

			// get error
			cmd2 := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build(context.Background())
			_, err = m.StartProcess(context.Background(), cmd2)
			Expect(err).NotTo(BeNil())
			Expect(strings.Contains(err.Error(), fmt.Sprintf("process with identifier %s is running", identifier))).To(BeTrue())
			WaitProcess(m, p1, time.Second*3)

			cmd3 := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build(context.Background())
			p3, err := m.StartProcess(context.Background(), cmd3)
			Expect(err).To(BeNil())

			err = m.KillBackgroundProcess(context.Background(), p3.Uid)
			Expect(err).To(BeNil())

			cmd4 := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build(context.Background())
			p4, err := m.StartProcess(context.Background(), cmd4)
			Expect(err).To(BeNil())
			WaitProcess(m, p4, time.Second*3)
		})
	})

	Context("get identifiers", func() {
		It("should work", func() {
			identifier := RandomeIdentifier()
			cmd := DefaultProcessBuilder("sleep", "2").
				SetIdentifier(identifier).
				Build(context.Background())

			p, err := m.StartProcess(context.Background(), cmd)
			Expect(err).To(BeNil())

			ids := m.GetIdentifiers()
			Expect(ids).To(Equal([]string{identifier}))

			WaitProcess(m, p, time.Second*3)

			// wait for deleting identifier
			time.Sleep(time.Second * 2)
			ids = m.GetIdentifiers()
			Expect(len(ids)).To(Equal(0))
		})

		It("should work with nil identifier", func() {
			cmd := DefaultProcessBuilder("sleep", "2").Build(context.Background())

			p, err := m.StartProcess(context.Background(), cmd)
			Expect(err).To(BeNil())

			ids := m.GetIdentifiers()
			Expect(len(ids)).To(Equal(0))

			WaitProcess(m, p, time.Second*5)
		})
	})

	Context("get uid", func() {
		It("kill process", func() {
			cmd := DefaultProcessBuilder("sleep", "2").Build(context.Background())
			p, err := m.StartProcess(context.Background(), cmd)
			Expect(err).To(BeNil())

			uid, loaded := m.GetUID(p.Pair)
			Expect(loaded).To(BeTrue())

			err = m.KillBackgroundProcess(context.Background(), uid)
			Expect(err).To(BeNil())

			WaitProcess(m, p, time.Second*0)
		})
	})
})
