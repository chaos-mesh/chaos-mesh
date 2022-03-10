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

package pod

import (
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

const phaseSelectorName = "phase"

type phaseSelector struct {
	reqIncl []labels.Requirement
	reqExcl []labels.Requirement
}

var _ generic.Selector = &phaseSelector{}

func (s *phaseSelector) ListOption() client.ListOption {
	return nil
}

func (s *phaseSelector) ListFunc(_ client.Reader) generic.ListFunc {
	return nil
}

func (s *phaseSelector) Match(obj client.Object) bool {
	included := len(s.reqIncl) == 0
	pod := obj.(*v1.Pod)
	selector := labels.Set{string(pod.Status.Phase): ""}

	// include pod if one including requirement matches
	for _, req := range s.reqIncl {
		if req.Matches(selector) {
			included = true
			break
		}
	}

	// exclude pod if it is filtered out by at least one excluding requirement
	for _, req := range s.reqExcl {
		if !req.Matches(selector) {
			return false
		}
	}

	return included
}

func newPhaseSelector(spec v1alpha1.PodSelectorSpec) (generic.Selector, error) {
	selectorStr := strings.Join(spec.PodPhaseSelectors, ",")
	selector, err := labels.Parse(selectorStr)
	if err != nil {
		return nil, err
	}

	reqs, _ := selector.Requirements()
	var (
		reqIncl []labels.Requirement
		reqExcl []labels.Requirement
	)

	for _, req := range reqs {
		switch req.Operator() {
		case selection.Exists:
			reqIncl = append(reqIncl, req)
		case selection.DoesNotExist:
			reqExcl = append(reqExcl, req)
		default:
			return nil, errors.Errorf("unsupported operator: %s", req.Operator())
		}
	}

	return &phaseSelector{
		reqIncl: reqIncl,
		reqExcl: reqExcl,
	}, nil
}
