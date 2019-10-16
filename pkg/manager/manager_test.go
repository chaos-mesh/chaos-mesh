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
	"strconv"
	"sync"
	"testing"

	. "github.com/onsi/gomega"
	gtype "github.com/onsi/gomega/types"
	"github.com/robfig/cron/v3"
)

func newFakeManagerBase() *ManagerBase {
	cronEngine := cron.New()
	cronEngine.Start()

	return NewManagerBase(cronEngine)
}

type resultF func() gtype.GomegaMatcher

func TestManagerBaseAddRunner(t *testing.T) {
	g := NewGomegaWithT(t)

	mgr := newFakeManagerBase()

	type TestCase struct {
		name           string
		newRunner      *Runner
		expectedResult resultF
	}

	tcs := []TestCase{
		{
			name: "runner is valid",
			newRunner: &Runner{
				Name: "test1",
				Rule: "* * 3 * *",
				Job:  &fakeJob{},
			},
			expectedResult: Succeed,
		},
		{
			name: "runner is valid",
			newRunner: &Runner{
				Name: "test2",
				Rule: "@every 2m",
				Job:  &fakeJob{},
			},
			expectedResult: Succeed,
		},
		{
			name: "name is empty",
			newRunner: &Runner{
				Name: "",
				Rule: "* * 3 * *",
				Job:  &fakeJob{},
			},
			expectedResult: HaveOccurred,
		},
		{
			name: "rule is empty",
			newRunner: &Runner{
				Name: "test1",
				Rule: "",
				Job:  &fakeJob{},
			},
			expectedResult: HaveOccurred,
		},
		{
			name: "rule is invalid",
			newRunner: &Runner{
				Name: "test1",
				Rule: "* * * * 1 *",
				Job:  &fakeJob{},
			},
			expectedResult: HaveOccurred,
		},
	}

	for _, tc := range tcs {
		g.Expect(mgr.AddRunner(tc.newRunner)).Should(tc.expectedResult(), tc.name)
	}

	g.Expect(lenSyncMap(&mgr.runners)).To(Equal(2))
}

func lenSyncMap(m *sync.Map) int {
	le := 0
	m.Range(func(k, v interface{}) bool {
		le++
		return true
	})
	return le
}

func TestManagerBaseDeleteRunner(t *testing.T) {
	g := NewGomegaWithT(t)

	mgr := newFakeManagerBase()

	for i := 0; i < 5; i++ {
		r := &Runner{
			Name: "test-" + strconv.Itoa(i),
			Rule: "* * 3 * *",
			Job:  &fakeJob{},
		}

		g.Expect(mgr.AddRunner(r)).Should(Succeed())
	}

	g.Expect(lenSyncMap(&mgr.runners)).To(Equal(5))

	type TestCase struct {
		name              string
		key               string
		expectedResult    resultF
		expectedPodsCount int
	}

	tcs := []TestCase{
		{
			name:              "delete test-2",
			key:               "test-2",
			expectedResult:    Succeed,
			expectedPodsCount: 4,
		},
		{
			name:              "delete test-4",
			key:               "test-4",
			expectedResult:    Succeed,
			expectedPodsCount: 3,
		},
		{
			name:              "delete test-4 again",
			key:               "test-4",
			expectedResult:    Succeed,
			expectedPodsCount: 3,
		},
		{
			name:              "delete test-10",
			key:               "test-10",
			expectedResult:    Succeed,
			expectedPodsCount: 3,
		},
	}

	for _, tc := range tcs {
		g.Expect(mgr.DeleteRunner(tc.key)).Should(tc.expectedResult(), tc.name)
		g.Expect(lenSyncMap(&mgr.runners)).To(Equal(tc.expectedPodsCount), tc.name)
	}
}

func TestManagerBaseUpdateRunner(t *testing.T) {
	g := NewGomegaWithT(t)

	mgr := newFakeManagerBase()

	type TestCase struct {
		name           string
		addRunner      *Runner
		updateRunner   *Runner
		expectedResult resultF
		updated        bool
		isChange       bool
	}

	tcs := []TestCase{
		{
			name: "update same runner",
			addRunner: &Runner{
				Name: "test-1",
				Rule: "* * 3 * *",
				Job:  &fakeJob2{},
			},
			updateRunner: &Runner{
				Name: "test-1",
				Rule: "* * 3 * *",
				Job:  &fakeJob2{},
			},
			expectedResult: Succeed,
			updated:        true,
			isChange:       false,
		},
		{
			name: "runner not exist",
			updateRunner: &Runner{
				Name: "test-no-found",
				Rule: "* * 3 * *",
				Job:  &fakeJob2{},
			},
			expectedResult: HaveOccurred,
			updated:        false,
			isChange:       false,
		},
		{
			name: "different runner rule",
			addRunner: &Runner{
				Name: "test-2",
				Rule: "* * 3 * *",
				Job:  &fakeJob2{},
			},
			updateRunner: &Runner{
				Name: "test-2",
				Rule: "@every 2m",
				Job:  &fakeJob2{},
			},
			expectedResult: Succeed,
			updated:        true,
			isChange:       true,
		},
		{
			name: "different Job",
			addRunner: &Runner{
				Name: "test-3",
				Rule: "* * 3 * *",
				Job:  &fakeJob{},
			},
			updateRunner: &Runner{
				Name: "test-3",
				Rule: "@every 2m",
				Job:  &fakeJob2{},
			},
			expectedResult: Succeed,
			updated:        true,
			isChange:       true,
		},
	}

	for _, tc := range tcs {
		var expectedID int
		if tc.addRunner != nil {
			g.Expect(mgr.AddRunner(tc.addRunner)).Should(Succeed(), tc.name)

			getRunner, ok := mgr.runners.Load(tc.addRunner.Name)
			g.Expect(ok).To(Equal(true), tc.name)

			expectedID = getRunner.(*Runner).EntryID
		}

		g.Expect(mgr.UpdateRunner(tc.updateRunner)).Should(tc.expectedResult(), tc.name)

		if !tc.updated {
			continue
		}

		getRunner, ok := mgr.runners.Load(tc.updateRunner.Name)
		g.Expect(ok).To(Equal(true), tc.name)

		newID := getRunner.(*Runner).EntryID
		if tc.isChange {
			g.Expect(newID).NotTo(Equal(expectedID), tc.name)
		} else {
			g.Expect(newID).To(Equal(expectedID), tc.name)
		}
	}

	g.Expect(lenSyncMap(&mgr.runners)).To(Equal(3))
}

type fakeJob struct{}

func (j *fakeJob) Run() {}

func (j *fakeJob) Equal(_ Job) bool { return false }

func (j *fakeJob) Close() error { return nil }

func (j *fakeJob) Clean() error { return nil }

func (j *fakeJob) Sync() error { return nil }

type fakeJob2 struct{}

func (j *fakeJob2) Run() {}

func (j *fakeJob2) Equal(_ Job) bool { return true }

func (j *fakeJob2) Close() error { return nil }

func (j *fakeJob2) Clean() error { return nil }

func (j *fakeJob2) Sync() error { return nil }
