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

package namespace

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func TestMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	pods := []v1.Pod{
		NewPod(PodArg{Name: "p1", Namespace: "n1"}),
		NewPod(PodArg{Name: "p2", Namespace: "n2"}),
		NewPod(PodArg{Name: "p4", Namespace: "n4"}),
	}

	option := generic.Option{
		ClusterScoped:         true,
		TargetNamespace:       "",
		EnableFilterNamespace: false,
	}

	n2Selector, err := New(v1alpha1.GenericSelectorSpec{Namespaces: []string{"n2"}}, option)
	g.Expect(err).ShouldNot(HaveOccurred())

	emptySelector, err := New(v1alpha1.GenericSelectorSpec{}, option)
	g.Expect(err).ShouldNot(HaveOccurred())

	n2AndN3Selector, err := New(v1alpha1.GenericSelectorSpec{Namespaces: []string{"n2", "n3"}}, option)
	g.Expect(err).ShouldNot(HaveOccurred())

	n2AndN4Selector, err := New(v1alpha1.GenericSelectorSpec{Namespaces: []string{"n2", "n4"}}, option)
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
