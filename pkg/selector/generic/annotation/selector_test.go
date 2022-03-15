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

package annotation

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
