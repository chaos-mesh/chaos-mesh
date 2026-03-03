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

package field

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func TestMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	nameFieldSelector, err := New(v1alpha1.GenericSelectorSpec{FieldSelectors: map[string]string{"metadata.name": "p2"}}, generic.Option{})
	g.Expect(err).ShouldNot(HaveOccurred())

	emptySelector, err := New(v1alpha1.GenericSelectorSpec{}, generic.Option{})
	g.Expect(err).ShouldNot(HaveOccurred())

	addressFieldSelector, err := New(v1alpha1.GenericSelectorSpec{FieldSelectors: map[string]string{"spec.address": "123"}}, generic.Option{})
	g.Expect(err).ShouldNot(HaveOccurred())

	p1Pod := NewPod(PodArg{Name: "p1"})
	p2Pod := NewPod(PodArg{Name: "p2"})
	p1PhysicalMachine := v1alpha1.PhysicalMachine{ObjectMeta: metav1.ObjectMeta{Name: "p1"}, Spec: v1alpha1.PhysicalMachineSpec{Address: "123"}}
	p2PhysicalMachine := v1alpha1.PhysicalMachine{ObjectMeta: metav1.ObjectMeta{Name: "p2"}}

	tcs := []struct {
		name     string
		obj      client.Object
		selector generic.Selector
		match    bool
	}{
		{
			name:     "filter by name",
			obj:      &p2Pod,
			selector: nameFieldSelector,
			match:    true,
		}, {
			name:     "filter by name",
			obj:      &p1Pod,
			selector: nameFieldSelector,
			match:    false,
		}, {
			name:     "empty filter",
			obj:      &p1Pod,
			selector: emptySelector,
			match:    true,
		}, {
			name:     "filter by physical machine address",
			obj:      &p1PhysicalMachine,
			selector: addressFieldSelector,
			match:    true,
		}, {
			name:     "filter by physical machine address",
			obj:      &p2PhysicalMachine,
			selector: addressFieldSelector,
			match:    false,
		}, {
			name:     "filter by physical machine address",
			obj:      &p1Pod,
			selector: addressFieldSelector,
			match:    false,
		},
	}

	for _, tc := range tcs {
		g.Expect(tc.selector.Match(tc.obj)).To(Equal(tc.match), tc.name)
	}
}
