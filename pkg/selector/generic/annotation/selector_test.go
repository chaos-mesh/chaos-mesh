package annotation

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"testing"

	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func TestMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	p2Selector, err := New(v1alpha1.GenericSelectorSpec{AnnotationSelectors: map[string]string{"p2": "p2"}}, generic.Option{})
	g.Expect(err).ShouldNot(HaveOccurred())

	emptySelector, err := New(v1alpha1.GenericSelectorSpec{}, generic.Option{})
	g.Expect(err).ShouldNot(HaveOccurred())

	tcs := []struct {
		name     string
		pod      v1.Pod
		selector generic.Selector
		match    bool
	}{
		{
			name:     "filter p2, exist p2 annotations",
			pod:      NewPod(PodArg{Name: "p2", Ans: map[string]string{"p2": "p2"}}),
			selector: p2Selector,
			match:    true,
		}, {
			name:     "filter p2, not exist p2 annotations",
			pod:      NewPod(PodArg{Name: "p1", Ans: map[string]string{"p1": "p1"}}),
			selector: p2Selector,
			match:    false,
		}, {
			name:     "empty filter",
			pod:      NewPod(PodArg{Name: "p1", Ans: map[string]string{"p1": "p1"}}),
			selector: emptySelector,
			match:    true,
		},
	}

	for _, tc := range tcs {
		g.Expect(tc.selector.Match(&tc.pod)).To(Equal(tc.match), tc.name)
	}
}
