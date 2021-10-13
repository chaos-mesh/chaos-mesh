package namespace

import (
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type namespaceSelector struct {
	generic.Option
}

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

func New(namespaces []string,
	clusterScoped bool, targetNamespace string, enableFilterNamespace bool) (generic.Selector, error) {
	if !clusterScoped {
		if len(namespaces) > 1 {
			return nil, fmt.Errorf("could NOT use more than 1 namespace selector within namespace scoped mode")
		} else if len(namespaces) == 1 {
			if namespaces[0] != targetNamespace {
				return nil, fmt.Errorf("could NOT list pods from out of scoped namespace: %s", namespaces[0])
			}
		}
	}

	return &namespaceSelector{
		generic.Option{
			clusterScoped,
			targetNamespace,
			enableFilterNamespace,
		},
	}, nil
}

