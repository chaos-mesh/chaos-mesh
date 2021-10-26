package field

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Name = "field"

type fieldSelector struct {
	selectors map[string]string
	r         client.Reader
}

var _ generic.Selector = &fieldSelector{}

func (s *fieldSelector) ListOption() client.ListOption {
	if len(s.selectors) > 0 {
		return client.MatchingFieldsSelector{Selector: fields.SelectorFromSet(s.selectors)}
	}
	return nil
}

func (s *fieldSelector) ListFunc() generic.ListFunc {
	// Since FieldSelectors need to implement index creation, Reader.List is used to get the pod list.
	// Otherwise, just call Client.List directly, which can be obtained through cache.
	if len(s.selectors) > 0 && s.r != nil {
		return s.r.List
	}
	return nil
}

func (s *fieldSelector) Match(_ client.Object) bool {
	return true
}

func New(spec v1alpha1.GenericSelectorSpec, r client.Reader, _ generic.Option) (generic.Selector, error) {
	return &fieldSelector{
		selectors: spec.FieldSelectors,
		r:         r,
	}, nil
}
