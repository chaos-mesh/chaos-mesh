package registry

import (
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/label"
)

type Registry map[string]SelectorFactory

// SelectorFactory is a function that builds a selector.
type SelectorFactory = func(selector v1alpha1.GenericSelectorSpec) (generic.Selector, error)

func NewGenericRegistry() Registry {
	return map[string]SelectorFactory{
		"label":label.New,
	}
}

func Parse(registry Registry, spec v1alpha1.GenericSelectorSpec) ([]generic.Selector, error) {
	selectors := make([]generic.Selector, 0, len(registry))
	for name, factory := range registry {
		selector, err := factory(spec)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %s selector", name)
		}
		selectors = append(selectors, selector)
	}
	return selectors, nil
}
