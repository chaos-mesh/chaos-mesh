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

	It("reapplies only the remaining tasks on the same target during recovery", func() {
		other := &FakeChaos{}
		Expect(manager.Create(otherTask, otherPID, &FakeConfig{i: 11}, other)).To(Succeed())
		Expect(manager.Apply(extraTask, targetPID, &FakeConfig{i: 3})).To(Succeed())
		Expect(task.C.i).To(Equal(5))
		Expect(manager.Update(extraTask, targetPID, &FakeConfig{i: 4})).To(Succeed())
		Expect(task.C.i).To(Equal(6))
		Expect(manager.GetUIDsWithPID(targetPID)).To(ConsistOf(baseTask, extraTask))

		Expect(manager.Recover(extraTask, targetPID)).To(Succeed())
		Expect(task.C.i).To(Equal(2))
		Expect(manager.GetUIDsWithPID(targetPID)).To(Equal([]TaskID{baseTask}))
		Expect(other.C.i).To(Equal(11))
		Expect(manager.GetUIDsWithPID(otherPID)).To(Equal([]TaskID{otherTask}))

		Expect(manager.Recover(baseTask, targetPID)).To(Succeed())
		_, err := manager.GetTaskWithPID(targetPID)
		Expect(err).To(HaveOccurred())
		Expect(manager.GetUIDsWithPID(targetPID)).To(BeEmpty())
		Expect(manager.GetTaskWithPID(otherPID)).To(BeIdenticalTo(other))
	})

	It("retains the committed config after a failed update and accepts a retry", func() {
		task.ErrWhenInject = true
		Expect(manager.Update(baseTask, targetPID, &FakeConfig{i: 7})).NotTo(Succeed())
		config, err := manager.GetConfigWithUID(baseTask)
		Expect(err).NotTo(HaveOccurred())
		Expect(config.Data).To(Equal(&FakeConfig{i: 2}))

		task.ErrWhenInject = false
		Expect(manager.Update(baseTask, targetPID, &FakeConfig{i: 7})).To(Succeed())
		config, err = manager.GetConfigWithUID(baseTask)
		Expect(err).NotTo(HaveOccurred())
		Expect(config.Data).To(Equal(&FakeConfig{i: 7}))
		Expect(task.C.i).To(Equal(7))
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

	It("rejects applying a task to an unregistered target and moving an existing task", func() {
		Expect(manager.Apply(extraTask, otherPID, &FakeConfig{i: 4})).NotTo(Succeed())
		_, err := manager.GetConfigWithUID(extraTask)
		Expect(err).To(HaveOccurred())
		Expect(manager.Update(baseTask, otherPID, &FakeConfig{i: 7})).NotTo(Succeed())
		Expect(manager.CheckTasks(baseTask, otherPID)).NotTo(Succeed())
		Expect(manager.CheckTasks(baseTask, targetPID)).To(Succeed())
		Expect(manager.GetUIDsWithPID(otherPID)).To(BeEmpty())
		Expect(task.C.i).To(Equal(2))
	})
})
