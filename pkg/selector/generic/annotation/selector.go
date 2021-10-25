package annotation

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/label"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Name = "annotation"

type annotationSelector struct {
	labels.Selector
	next generic.Selector
}

var _ generic.Selector = &annotationSelector{}

func (s *annotationSelector) List(listFunc generic.ListFunc, opts client.ListOptions,
	listObj func(listFunc generic.ListFunc, opts client.ListOptions) error) error {
	if s.next != nil {
		return s.next.List(listFunc, opts, listObj)
	}
	return listObj(listFunc, opts)
}

func (s *annotationSelector) Next(selector generic.Selector) {
	s.next = selector
}

func (s *annotationSelector) Match(obj client.Object) bool {
	if s.Empty() {
		return true
	}
	annotations := labels.Set(obj.GetAnnotations())
	return s.Matches(annotations)
}

func New(spec v1alpha1.GenericSelectorSpec, _ client.Reader, _ generic.Option) (generic.Selector, error) {
	selectorStr := label.Label(spec.AnnotationSelectors).String()
	s, err := labels.Parse(selectorStr)
	if err != nil {
		return nil, err
	}
	return &annotationSelector{Selector: s}, nil
}
