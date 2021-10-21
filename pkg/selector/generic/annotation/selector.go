package annotation

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/label"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type annotationSelector struct {
	labels.Selector
}

func (s *annotationSelector) AddListOption(opts client.ListOptions) client.ListOptions {
	return opts
}

func (s *annotationSelector) SetListFunc(f generic.ListFunc) generic.ListFunc {
	return f
}

func (s *annotationSelector) Match(obj client.Object) bool {
	if s.Empty() {
		return true
	}
	annotations := labels.Set(obj.GetAnnotations())
	return s.Matches(annotations)
}

func New(spec v1alpha1.GenericSelectorSpec) (generic.Selector, error) {
	selectorStr := label.Label(spec.AnnotationSelectors).String()
	s, err := labels.Parse(selectorStr)
	if err != nil {
		return nil, err
	}
	return &annotationSelector{s}, nil
}
