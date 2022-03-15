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

package pod

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func TestPhaseSelectorMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	runningSelector, err := newPhaseSelector(v1alpha1.PodSelectorSpec{PodPhaseSelectors: []string{string(v1.PodRunning)}})
	g.Expect(err).ShouldNot(HaveOccurred())

	emptySelector, err := newPhaseSelector(v1alpha1.PodSelectorSpec{})
	g.Expect(err).ShouldNot(HaveOccurred())

	runningAndPendingSelector, err := newPhaseSelector(v1alpha1.PodSelectorSpec{PodPhaseSelectors: []string{string(v1.PodRunning), string(v1.PodPending)}})
	g.Expect(err).ShouldNot(HaveOccurred())

	failedSelector, err := newPhaseSelector(v1alpha1.PodSelectorSpec{PodPhaseSelectors: []string{string(v1.PodFailed)}})
	g.Expect(err).ShouldNot(HaveOccurred())

	unknownSelector, err := newPhaseSelector(v1alpha1.PodSelectorSpec{PodPhaseSelectors: []string{string(v1.PodUnknown)}})
	g.Expect(err).ShouldNot(HaveOccurred())

	pods := []v1.Pod{
		NewPod(PodArg{Name: "p1", Status: v1.PodRunning}),
		NewPod(PodArg{Name: "p2", Status: v1.PodRunning}),
		NewPod(PodArg{Name: "p3", Status: v1.PodPending}),
		NewPod(PodArg{Name: "p4", Status: v1.PodFailed}),
	}

	tcs := []struct {
		name     string
		pod      v1.Pod
		selector generic.Selector
		match    bool
	}{
		{
			name:     "filter running pod, exist running pod",
			pod:      pods[0],
			selector: runningSelector,
			match:    true,
		}, {
			name:     "filter running pod, not exist running pod",
			pod:      pods[2],
			selector: runningSelector,
			match:    false,
		}, {
			name:     "empty filter",
			pod:      pods[0],
			selector: emptySelector,
			match:    true,
		}, {
			name:     "filter running and pending",
			pod:      pods[0],
			selector: runningAndPendingSelector,
			match:    true,
		}, {
			name:     "filter running and pending",
			pod:      pods[2],
			selector: runningAndPendingSelector,
			match:    true,
		}, {
			name:     "filter running and pending",
			pod:      pods[3],
			selector: runningAndPendingSelector,
			match:    false,
		}, {
			name:     "filter failed",
			pod:      pods[3],
			selector: failedSelector,
			match:    true,
		}, {
			name:     "filter failed",
			pod:      pods[0],
			selector: failedSelector,
			match:    false,
		}, {
			name:     "filter unknown",
			pod:      pods[0],
			selector: unknownSelector,
			match:    false,
		},
	}

	for _, tc := range tcs {
		g.Expect(tc.selector.Match(&tc.pod)).To(Equal(tc.match), tc.name)
	}
}
