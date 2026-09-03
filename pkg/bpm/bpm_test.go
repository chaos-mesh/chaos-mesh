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
	"io"
	"math/rand"
	"path/filepath"
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
	logger := log.NewZapLoggerWithWriter(GinkgoWriter)
	m := StartBackgroundProcessManager(nil, logger)

	Context("failed startup", func() {
		DescribeTable("releases the identifier for a retry", func(prepare func(*ManagedCommand)) {
			ctx, cancel := context.WithCancel(context.Background())
			DeferCleanup(cancel)
			identifier := RandomeIdentifier()
			cmd := DefaultProcessBuilder("sleep", "30").
				SetIdentifier(identifier).
				SetContext(ctx).
				Build(ctx)
			prepare(cmd)

			proc, err := m.StartProcess(ctx, cmd)
			Expect(err).To(HaveOccurred())
			Expect(proc).To(BeNil())
			Expect(m.GetIdentifiers()).NotTo(ContainElement(identifier))

			retry := DefaultProcessBuilder("sleep", "30").
				SetIdentifier(identifier).
				SetContext(ctx).
				Build(ctx)
			proc, err = m.StartProcess(ctx, retry)
			Expect(err).NotTo(HaveOccurred())
			Expect(m.KillBackgroundProcess(ctx, proc.Uid)).To(Succeed())
		},
			Entry("when stdin is already configured", func(cmd *ManagedCommand) {
				cmd.Stdin = strings.NewReader("")
			}),
			Entry("when stdout is already configured", func(cmd *ManagedCommand) {
				cmd.Stdout = io.Discard
			}),
			Entry("when the executable does not exist", func(cmd *ManagedCommand) {
				cmd.Path = filepath.Join(GinkgoT().TempDir(), "missing-executable")
			}),
		)

		It("keeps the live owner's identifier after duplicate attempts", func() {
			ctx, cancel := context.WithCancel(context.Background())
			DeferCleanup(cancel)
			identifier := RandomeIdentifier()
			cmd := DefaultProcessBuilder("sleep", "30").
				SetIdentifier(identifier).
				SetContext(ctx).
				Build(ctx)
			proc, err := m.StartProcess(ctx, cmd)
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < 2; i++ {
				duplicate := DefaultProcessBuilder("sleep", "30").
					SetIdentifier(identifier).
					SetContext(ctx).
					Build(ctx)
				_, err := m.StartProcess(ctx, duplicate)
				Expect(err).To(MatchError(fmt.Sprintf("process with identifier %s is running", identifier)))
				Expect(duplicate.Process).To(BeNil())
				Expect(m.GetIdentifiers()).To(ContainElement(identifier))
			}

			Expect(m.KillBackgroundProcess(ctx, proc.Uid)).To(Succeed())
		})
	})

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
