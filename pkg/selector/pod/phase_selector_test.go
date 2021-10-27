package pod

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"testing"

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
		},{
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
