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

package registry

import (
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

type Registry map[string]SelectorFactory

// SelectorFactory is a function that builds a selector.
type SelectorFactory = func(selector v1alpha1.GenericSelectorSpec, option generic.Option) (generic.Selector, error)

func Parse(registry Registry, spec v1alpha1.GenericSelectorSpec, option generic.Option) (generic.SelectorChain, error) {
	selectors := make([]generic.Selector, 0, len(registry))
	for name, factory := range registry {
		selector, err := factory(spec, option)
		if err != nil {
			return nil, errors.Errorf("cannot parse %s selector, msg: %+v", name, err)
		}
		selectors = append(selectors, selector)
	}
	return selectors, nil
}
