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
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

const Name = "field"

type fieldSelector struct {
	selectors map[string]string
}

var _ generic.Selector = &fieldSelector{}

func (s *fieldSelector) ListOption() client.ListOption {
	if len(s.selectors) > 0 {
		return client.MatchingFieldsSelector{Selector: fields.SelectorFromSet(s.selectors)}
	}
	return nil
}

func (s *fieldSelector) ListFunc(r client.Reader) generic.ListFunc {
	// Since FieldSelectors need to implement index creation, Reader.List is used to get the pod list.
	// Otherwise, just call Client.List directly, which can be obtained through cache.
	if len(s.selectors) > 0 && r != nil {
		return r.List
	}
	return nil
}

func (s *fieldSelector) Match(_ client.Object) bool {
	return true
}

func New(spec v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
	return &fieldSelector{
		selectors: spec.FieldSelectors,
	}, nil
}
