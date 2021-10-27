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
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

const Name = "namespace"

type namespaceSelector struct {
	generic.Option
}

var _ generic.Selector = &namespaceSelector{}

func (s *namespaceSelector) ListOption() client.ListOption {
	if !s.ClusterScoped {
		return client.InNamespace(s.TargetNamespace)
	}
	return nil
}

func (s *namespaceSelector) ListFunc(_ client.Reader) generic.ListFunc {
	return nil
}

func (s *namespaceSelector) Match(_ client.Object) bool {
	return true
}

func New(spec v1alpha1.GenericSelectorSpec, option generic.Option) (generic.Selector, error) {
	if !option.ClusterScoped {
		if len(spec.Namespaces) > 1 {
			return nil, fmt.Errorf("could NOT use more than 1 namespace selector within namespace scoped mode")
		} else if len(spec.Namespaces) == 1 {
			if spec.Namespaces[0] != option.TargetNamespace {
				return nil, fmt.Errorf("could NOT list pods from out of scoped namespace: %s", spec.Namespaces[0])
			}
		}
	}

	return &namespaceSelector{
		Option: generic.Option{
			ClusterScoped:         option.ClusterScoped,
			TargetNamespace:       option.TargetNamespace,
			EnableFilterNamespace: option.EnableFilterNamespace,
		},
	}, nil
}
