package field

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fieldSelector struct {
	FieldSelectors map[string]string
	r              client.Reader
}

func (s *fieldSelector) AddListOption(opts client.ListOptions) client.ListOptions {
	if len(s.FieldSelectors) > 0 {
		opts.FieldSelector = fields.SelectorFromSet(s.FieldSelectors)
	}
	return opts
}

func (s *fieldSelector) SetListFunc(f generic.ListFunc) generic.ListFunc {
	// Since FieldSelectors need to implement index creation, Reader.List is used to get the pod list.
	// Otherwise, just call Client.List directly, which can be obtained through cache.
	if len(s.FieldSelectors) > 0 && s.r != nil {
		return s.r.List
	}
	return f
}

func (s *fieldSelector) Match(obj client.Object) bool {
	return true
}

func New(spec v1alpha1.GenericSelectorSpec) (generic.Selector, error) {

	return &fieldSelector{}, nil
}
