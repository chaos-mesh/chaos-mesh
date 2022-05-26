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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

const Name = "namespace"

type namespaceSelector struct {
	generic.Option
	namespaces []string
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
	if len(s.namespaces) == 0 {
		return true
	}

	for _, namespace := range s.namespaces {
		if namespace == obj.GetNamespace() {
			return true
		}
	}
	return false
}

func New(spec v1alpha1.GenericSelectorSpec, option generic.Option) (generic.Selector, error) {
	if !option.ClusterScoped {
		if len(spec.Namespaces) > 1 {
			return nil, errors.New("could NOT use more than 1 namespace selector within namespace scoped mode")
		} else if len(spec.Namespaces) == 1 {
			if spec.Namespaces[0] != option.TargetNamespace {
				return nil, errors.Errorf("could NOT list pods from out of scoped namespace: %s", spec.Namespaces[0])
			}
		}
	}

	return &namespaceSelector{
		Option: generic.Option{
			ClusterScoped:         option.ClusterScoped,
			TargetNamespace:       option.TargetNamespace,
			EnableFilterNamespace: option.EnableFilterNamespace,
		},
		namespaces: spec.Namespaces,
	}, nil
}

func CheckNamespace(ctx context.Context, c client.Client, namespace string, logger logr.Logger) bool {
	ok, err := IsAllowedNamespaces(ctx, c, namespace)
	if err != nil {
		logger.Error(err, "fail to check whether this namespace is allowed", "namespace", namespace)
		return false
	}

	if !ok {
		logger.Info("namespace is not enabled for chaos-mesh", "namespace", namespace)
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
