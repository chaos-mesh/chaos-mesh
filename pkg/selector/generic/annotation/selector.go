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
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/label"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

const Name = "annotation"

type annotationSelector struct {
	labels.Selector
}

var _ generic.Selector = &annotationSelector{}

func (s *annotationSelector) ListOption() client.ListOption {
	return nil
}

func (s *annotationSelector) ListFunc(_ client.Reader) generic.ListFunc {
	return nil
}

func (s *annotationSelector) Match(obj client.Object) bool {
	annotations := labels.Set(obj.GetAnnotations())
	return s.Matches(annotations)
}

func New(spec v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
	selectorStr := label.Label(spec.AnnotationSelectors).String()
	s, err := labels.Parse(selectorStr)
	if err != nil {
		return nil, err
	}
	return &annotationSelector{Selector: s}, nil
}
