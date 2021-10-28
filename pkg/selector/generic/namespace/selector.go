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
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/selection"

	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

const Name = "namespace"

var log = ctrl.Log.WithName("namespaceselector")

type namespaceSelector struct {
	generic.Option
	reqIncl []labels.Requirement
	reqExcl []labels.Requirement
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

func (s *namespaceSelector) Match(obj client.Object) bool {
	included := len(s.reqIncl) == 0
	selector := labels.Set{obj.GetNamespace(): ""}

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

	selectorStr := strings.Join(spec.Namespaces, ",")
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
			return nil, fmt.Errorf("unsupported operator: %s", req.Operator())
		}
	}

	return &namespaceSelector{
		Option: generic.Option{
			ClusterScoped:         option.ClusterScoped,
			TargetNamespace:       option.TargetNamespace,
			EnableFilterNamespace: option.EnableFilterNamespace,
		},
		reqIncl: reqIncl,
		reqExcl: reqExcl,
	}, nil
}

func CheckNamespace(ctx context.Context, c client.Client, namespace string) bool {
	ok, err := IsAllowedNamespaces(ctx, c, namespace)
	if err != nil {
		log.Error(err, "fail to check whether this namespace is allowed", "namespace", namespace)
		return false
	}

	if !ok {
		log.Info("namespace is not enabled for chaos-mesh", "namespace", namespace)
	}
	return ok
}

func IsAllowedNamespaces(ctx context.Context, c client.Client, namespace string) (bool, error) {
	ns := &v1.Namespace{}

	err := c.Get(ctx, types.NamespacedName{Name: namespace}, ns)
	if err != nil {
		return false, err
	}

	if ns.Annotations[generic.InjectAnnotationKey] == "enabled" {
		return true, nil
	}

	return false, nil
}
