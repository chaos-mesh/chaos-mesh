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
	next      generic.Selector
}

var _ generic.Selector = &fieldSelector{}

func (s *fieldSelector) List(listFunc generic.ListFunc, opts client.ListOptions,
	listObj func(listFunc generic.ListFunc, opts client.ListOptions) error) error {
	if len(s.selectors) > 0 {
		opts.FieldSelector = fields.SelectorFromSet(s.selectors)
	}
	// Since FieldSelectors need to implement index creation, Reader.List is used to get the pod list.
	// Otherwise, just call Client.List directly, which can be obtained through cache.
	if len(s.selectors) > 0 && s.r != nil {
		listFunc = s.r.List
	}
	if s.next != nil {
		return s.next.List(listFunc, opts, listObj)
	}
	return listObj(listFunc, opts)
}

func (s *fieldSelector) Match(obj client.Object) bool {
	if s.next != nil {
		return s.next.Match(obj)
	}
	return true
}

func (s *fieldSelector) Next(selector generic.Selector) {
	s.next = selector
}

func New(spec v1alpha1.GenericSelectorSpec, r client.Reader, _ generic.Option) (generic.Selector, error) {
	return &fieldSelector{
		selectors: spec.FieldSelectors,
		r:         r,
	}, nil
}
