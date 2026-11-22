// Copyright Chaos Mesh Authors.
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

package tasks

import (
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTaskManagerLifecycle(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Daemon Task Manager Lifecycle")
}

var _ = Describe("Task manager lifecycle", func() {
	const (
		baseTask  TaskID = "base"
		extraTask TaskID = "extra"
		otherTask TaskID = "other"
		targetPID SysPID = 42
		otherPID  SysPID = 43
	)
	var (
		manager TaskManager
		task    *FakeChaos
	)

	BeforeEach(func() {
		manager = NewTaskManager(logr.Discard())
		task = &FakeChaos{}
		Expect(manager.Create(baseTask, targetPID, &FakeConfig{i: 2}, task)).To(Succeed())
	})

	It("retains both task IDs if reapplying the remaining offset fails during recovery", func() {
		Expect(manager.Apply(extraTask, targetPID, &FakeConfig{i: 3})).To(Succeed())
		task.ErrWhenInject = true
		Expect(manager.Recover(extraTask, targetPID)).NotTo(Succeed())
		Expect(manager.GetUIDsWithPID(targetPID)).To(ConsistOf(baseTask, extraTask))

		task.ErrWhenInject = false
		Expect(manager.Recover(extraTask, targetPID)).To(Succeed())
		Expect(task.C.i).To(Equal(2))
		Expect(manager.GetUIDsWithPID(targetPID)).To(ConsistOf(baseTask))
	})

	It("keeps a task assigned to its original target when recovery names another target", func() {
		other := &FakeChaos{}
		Expect(manager.Create(otherTask, otherPID, &FakeConfig{i: 11}, other)).To(Succeed())
		Expect(manager.Apply(extraTask, targetPID, &FakeConfig{i: 3})).To(Succeed())

		Expect(manager.Recover(otherTask, targetPID)).NotTo(Succeed())
		Expect(manager.GetUIDsWithPID(targetPID)).To(ConsistOf(baseTask, extraTask))
		Expect(manager.GetUIDsWithPID(otherPID)).To(ConsistOf(otherTask))
		Expect(task.C.i).To(Equal(5))
		Expect(other.C.i).To(Equal(11))
		Expect(manager.Recover(otherTask, otherPID)).To(Succeed())
	})
})
