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
)

func TestMatch(t *testing.T) {
	// TODO
	//g := NewGomegaWithT(t)
	//
	//nameFiledSelector, err := New(v1alpha1.GenericSelectorSpec{FieldSelectors: map[string]string{"metadata.name": "p2"}}, generic.Option{})
	//g.Expect(err).ShouldNot(HaveOccurred())
	//
	//emptySelector, err := New(v1alpha1.GenericSelectorSpec{}, generic.Option{})
	//g.Expect(err).ShouldNot(HaveOccurred())
	//
	//p1Pod := NewPod(PodArg{Name: "p1"})
	//p2Pod := NewPod(PodArg{Name: "p2"})
	//
	//tcs := []struct {
	//	name     string
	//	obj      client.Object
	//	selector generic.Selector
	//	match    bool
	//}{
	//	{
	//		name:     "filter by name",
	//		obj:      &p2Pod,
	//		selector: nameFiledSelector,
	//		match:    true,
	//	}, {
	//		name:     "filter by name",
	//		obj:      NewPod(PodArg{Name: "p1", Labels: map[string]string{"p1": "p1"}}),
	//		selector: nameFiledSelector,
	//		match:    false,
	//	}, {
	//		name:     "empty filter",
	//		obj:      NewPod(PodArg{Name: "p1", Labels: map[string]string{"p1": "p1"}}),
	//		selector: emptySelector,
	//		match:    true,
	//	},
	//}
	//
	//for _, tc := range tcs {
	//	g.Expect(tc.selector.Match(tc.obj)).To(Equal(tc.match), tc.name)
	//}
}
