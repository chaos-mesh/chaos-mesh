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
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
)

type podProcessRecorder struct {
	injected  []IsID
	recovered []IsID
	failure   error
}

func (r *podProcessRecorder) Inject(pid IsID) error {
	r.injected = append(r.injected, pid)
	return r.failure
}

func (r *podProcessRecorder) Recover(pid IsID) error {
	r.recovered = append(r.recovered, pid)
	return r.failure
}

func TestPodHandlerUsesTheCurrentProcessForTheContainer(t *testing.T) {
	g := NewWithT(t)
	const container PodContainerName = "namespace/pod/container"
	const other PodContainerName = "namespace/pod/other"
	mapping := NewPodProcessMap()
	mapping.Write(container, 11)
	mapping.Write(other, 43)
	process := &podProcessRecorder{}
	handler := NewPodHandler(&mapping, process, logr.Discard())

	g.Expect(handler.Inject(container)).To(Succeed())
	mapping.Write(container, 12)
	g.Expect(handler.Inject(container)).To(Succeed())
	g.Expect(handler.Recover(container)).To(Succeed())
	g.Expect(process.injected).To(Equal([]IsID{SysPID(11), SysPID(12)}))
	g.Expect(process.recovered).To(Equal([]IsID{SysPID(12)}))

	mapping.Delete(container)
	g.Expect(handler.Inject(container)).NotTo(Succeed())
	g.Expect(handler.Recover(container)).NotTo(Succeed())
	g.Expect(mapping.Read(other)).To(Equal(SysPID(43)))
	g.Expect(process.injected).To(HaveLen(2))
	g.Expect(process.recovered).To(HaveLen(1))
}

func TestPodHandlerRejectsInvalidTargetsAndPropagatesProcessFailures(t *testing.T) {
	g := NewWithT(t)
	const container PodContainerName = "namespace/pod/container"
	const pid SysPID = 42
	mapping := NewPodProcessMap()
	mapping.Write(container, pid)
	process := &podProcessRecorder{failure: errors.New("process operation failed")}
	handler := NewPodHandler(&mapping, process, logr.Discard())
	uninitialized := NewPodHandler(nil, process, logr.Discard())

	g.Expect(handler.Inject(pid)).NotTo(Succeed())
	g.Expect(handler.Recover(pid)).NotTo(Succeed())
	g.Expect(uninitialized.Inject(container)).To(MatchError(ErrPodProcessMapNotInit))
	g.Expect(uninitialized.Recover(container)).To(MatchError(ErrPodProcessMapNotInit))
	g.Expect(process.injected).To(BeEmpty())
	g.Expect(process.recovered).To(BeEmpty())

	g.Expect(handler.Inject(container)).To(MatchError(process.failure))
	g.Expect(handler.Recover(container)).To(MatchError(process.failure))
	g.Expect(process.injected).To(Equal([]IsID{pid}))
	g.Expect(process.recovered).To(Equal([]IsID{pid}))
}
