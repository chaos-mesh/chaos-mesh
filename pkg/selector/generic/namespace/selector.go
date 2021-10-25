package namespace

import (
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type namespaceSelector struct {
	generic.Option
}

var _ generic.Selector = &namespaceSelector{}

func (s *namespaceSelector) AddListOption(opts client.ListOptions) client.ListOptions {
	if !s.ClusterScoped {
		opts.Namespace = s.TargetNamespace
	}
	return opts
}

func (s *namespaceSelector) SetListFunc(f generic.ListFunc) generic.ListFunc {
	return f
}

func (s *namespaceSelector) Match(obj client.Object) bool {
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
		generic.Option{
			option.ClusterScoped,
			option.TargetNamespace,
			option.EnableFilterNamespace,
		},
	}, nil
}
