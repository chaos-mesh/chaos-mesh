package namespace

import (
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (s *namespaceSelector) ListFunc() generic.ListFunc {
	return nil
}

func (s *namespaceSelector) Match(_ client.Object) bool {
	return true
}

// TODO validate?
func (s *namespaceSelector) Validate() error {
	if !s.ClusterScoped{

	}
	return nil
}

func New(spec v1alpha1.GenericSelectorSpec, _ client.Reader, option generic.Option) (generic.Selector, error) {
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
			option.ClusterScoped,
			option.TargetNamespace,
			option.EnableFilterNamespace,
		},
	}, nil
}
