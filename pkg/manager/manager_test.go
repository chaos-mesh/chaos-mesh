// Copyright 2019 PingCAP, Inc.
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

package manager

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/robfig/cron/v3"
)

func TestManagerBase(t *testing.T) {
	g := NewGomegaWithT(t)

	cronEngine := cron.New()
	cronEngine.Start()

	mgr := NewManagerBase(cronEngine)

	r1 := &Runner{
		Name: "test1",
		Rule: "* * 3 * *",
		Job:  fakeJob{},
	}

	g.Expect(mgr.AddRunner(r1)).Should(Succeed())

	getR, ok := mgr.runners.Load(r1.Name)
	g.Expect(ok).To(Equal(true))
	g.Expect(getR.(*Runner).EntryID).NotTo(Equal(0))
	g.Expect(getR.(*Runner).Name).To(Equal(r1.Name))

	r2 := &Runner{
		Name: "test2",
		Rule: "@every 2m",
		Job:  fakeJob2{},
	}
	g.Expect(mgr.AddRunner(r2)).Should(Succeed())

	lenF := func() int {
		le := 0
		mgr.runners.Range(func(k, v interface{}) bool {
			le++
			return true
		})
		return le
	}

	g.Expect(lenF()).To(Equal(2))

	getR2, ok := mgr.runners.Load(r2.Name)
	g.Expect(ok).To(Equal(true))

	getR2ID := getR2.(*Runner).EntryID
	g.Expect(getR2ID).NotTo(Equal(0))

	g.Expect(mgr.UpdateRunner(r2)).Should(Succeed())
	getR2s, ok := mgr.runners.Load(r2.Name)
	g.Expect(ok).To(Equal(true))
	g.Expect(getR2s.(*Runner).EntryID).To(Equal(getR2ID))

	r3 := &Runner{
		Name: r2.Name,
		Rule: "@every 3m",
		Job:  r2.Job,
	}
	g.Expect(mgr.UpdateRunner(r3)).Should(Succeed())
	getR3, ok := mgr.runners.Load(r3.Name)

	g.Expect(ok).To(Equal(true))
	g.Expect(getR3.(*Runner).EntryID).NotTo(Equal(0))
	g.Expect(getR3.(*Runner).EntryID).NotTo(Equal(getR2ID))

	g.Expect(lenF()).To(Equal(2))

	_, exist := mgr.GetRunner(r1.Name)
	g.Expect(exist).To(Equal(true))

	_, exist = mgr.GetRunner(r2.Name)
	g.Expect(exist).To(Equal(true))

	_, exist = mgr.GetRunner("test-no")
	g.Expect(exist).To(Equal(false))

	g.Expect(mgr.DeleteRunner(r1.Name)).Should(Succeed())
	g.Expect(lenF()).To(Equal(1))
	_, exist = mgr.GetRunner(r1.Name)
	g.Expect(exist).To(Equal(false))

	_, exist = mgr.GetRunner(r2.Name)
	g.Expect(exist).To(Equal(true))
}

type fakeJob struct{}

func (j fakeJob) Run() {}

func (j fakeJob) Equal(_ Job) bool { return false }

type fakeJob2 struct{}

func (j fakeJob2) Run() {}

func (j fakeJob2) Equal(_ Job) bool { return true }
