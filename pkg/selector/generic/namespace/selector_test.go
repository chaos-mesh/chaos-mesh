package namespace

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

	pods := []v1.Pod{
		NewPod(PodArg{Name: "p1", Namespace: "n1"}),
		NewPod(PodArg{Name: "p2", Namespace: "n2"}),
		NewPod(PodArg{Name: "p4", Namespace: "n4"}),
	}

	n2Selector, err := New(v1alpha1.GenericSelectorSpec{Namespaces: []string{"n2"}}, generic.Option{})
	g.Expect(err).ShouldNot(HaveOccurred())

	emptySelector, err := New(v1alpha1.GenericSelectorSpec{}, generic.Option{})
	g.Expect(err).ShouldNot(HaveOccurred())

	n2AndN3Selector, err := New(v1alpha1.GenericSelectorSpec{Namespaces: []string{"n2,n3"}}, generic.Option{})
	g.Expect(err).ShouldNot(HaveOccurred())

	n2AndN4Selector, err := New(v1alpha1.GenericSelectorSpec{Namespaces: []string{"n2,n4"}}, generic.Option{})
	g.Expect(err).ShouldNot(HaveOccurred())

	tcs := []struct {
		name     string
		pod      v1.Pod
		selector generic.Selector
		match    bool
	}{
		{
			name:     "filter n2, exist n2 namespace",
			pod:      pods[1],
			selector: n2Selector,
			match:    true,
		}, {
			name:     "filter n2, not exist n2 namespace",
			pod:      pods[0],
			selector: n2Selector,
			match:    false,
		}, {
			name:     "empty filter",
			pod:      pods[0],
			selector: emptySelector,
			match:    true,
		}, {
			name:     "filter n2 and n3, exist n2 namespace",
			pod:      pods[1],
			selector: n2AndN3Selector,
			match:    true,
		}, {
			name:     "filter n2 and n3, not exist n2 namespace",
			pod:      pods[0],
			selector: n2AndN3Selector,
			match:    false,
		}, {
			name:     "filter n2 and n4, exist n2 namespace",
			pod:      pods[1],
			selector: n2AndN4Selector,
			match:    true,
		}, {
			name:     "filter n2 and n4, exist n4 namespace",
			pod:      pods[2],
			selector: n2AndN4Selector,
			match:    true,
		}, {
			name:     "filter n2 and n4, not exist n2 namespace",
			pod:      pods[0],
			selector: n2AndN4Selector,
			match:    false,
		},
	}

	for _, tc := range tcs {
		g.Expect(tc.selector.Match(&tc.pod)).To(Equal(tc.match), tc.name)
	}
}
