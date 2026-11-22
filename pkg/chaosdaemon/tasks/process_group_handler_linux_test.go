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

package tasks

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
)

type processGroupFixture struct {
	child                                     ChaosOnProcessGroup
	injections, recoveries                    []IsID
	forks, assignments                        int
	injectErr, recoverErr, forkErr, assignErr error
}

func (f *processGroupFixture) Fork() (ChaosOnProcessGroup, error) {
	f.forks++
	return f.child, f.forkErr
}

func (f *processGroupFixture) Assign(Injectable) error {
	f.assignments++
	return f.assignErr
}

func (f *processGroupFixture) Inject(pid IsID) error {
	f.injections = append(f.injections, pid)
	return f.injectErr
}

func (f *processGroupFixture) Recover(pid IsID) error {
	f.recoveries = append(f.recoveries, pid)
	return f.recoverErr
}

type invalidProcessID string

func (id invalidProcessID) ToID() string { return string(id) }

func TestProcessGroupHandlerRetriesChildren(t *testing.T) {
	g := NewWithT(t)
	parent := exec.Command("sh", "-c", `sleep 30 & child=$!; printf '%s\n' "$child"; wait "$child"`)
	stdout, err := parent.StdoutPipe()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(parent.Start()).To(Succeed())
	var childProcess *os.Process
	t.Cleanup(func() {
		if childProcess == nil {
			_ = parent.Process.Kill()
		} else {
			_ = childProcess.Kill()
			_ = childProcess.Release()
		}
		_ = parent.Wait()
	})
	var childID uint64
	scanner := bufio.NewScanner(stdout)
	g.Expect(scanner.Scan()).To(BeTrue())
	childID, err = strconv.ParseUint(scanner.Text(), 10, 32)
	g.Expect(err).NotTo(HaveOccurred())
	childProcess, err = os.FindProcess(int(childID))
	g.Expect(err).NotTo(HaveOccurred())

	leaderPID, childPID := SysPID(parent.Process.Pid), SysPID(childID)
	child := &processGroupFixture{}
	leader := &processGroupFixture{child: child}
	handler := NewProcessGroupHandler(logr.Discard(), leader)
	g.Expect(handler.Inject(invalidProcessID("invalid"))).NotTo(Succeed())
	g.Expect(handler.Recover(invalidProcessID("invalid"))).NotTo(Succeed())
	g.Expect(leader.injections).To(BeEmpty())

	failure := errors.New("fixture failed")
	leader.injectErr = failure
	g.Expect(handler.Inject(leaderPID)).To(MatchError(ContainSubstring("inject leader process")))
	g.Expect(leader.forks).To(BeZero())
	leader.injectErr = nil
	leader.forkErr = failure
	g.Expect(handler.Inject(leaderPID)).To(Succeed())
	g.Expect(handler.childMap).To(BeEmpty())

	leader.forkErr = nil
	child.injectErr = failure
	g.Expect(handler.Inject(leaderPID)).To(Succeed())
	g.Expect(child.injections).To(Equal([]IsID{childPID}))
	g.Expect(handler.childMap).To(BeEmpty())

	child.injectErr = nil
	g.Expect(handler.Inject(leaderPID)).To(Succeed())
	g.Expect(handler.childMap).To(HaveKeyWithValue(childPID, child))
	g.Expect(leader.forks).To(Equal(3))

	leader.assignErr = failure
	g.Expect(handler.Inject(leaderPID)).To(Succeed())
	g.Expect(child.injections).To(HaveLen(2))
	leader.assignErr = nil
	child.injectErr = failure
	g.Expect(handler.Inject(leaderPID)).To(Succeed())
	g.Expect(child.injections).To(Equal([]IsID{childPID, childPID, childPID}))
	g.Expect(leader.forks).To(Equal(3))
	g.Expect(leader.assignments).To(Equal(2))

	leader.recoverErr = failure
	g.Expect(handler.Recover(leaderPID)).To(MatchError(ContainSubstring("recovery leader process")))
	g.Expect(child.recoveries).To(BeEmpty())
	leader.recoverErr = nil
	child.recoverErr = failure
	g.Expect(handler.Recover(leaderPID)).To(Succeed())
	g.Expect(leader.recoveries).To(Equal([]IsID{leaderPID, leaderPID}))
	g.Expect(child.recoveries).To(Equal([]IsID{childPID}))
}
