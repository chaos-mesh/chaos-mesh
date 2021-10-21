package label

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type labelSelector struct {
	selector labels.Selector
}

func (s *labelSelector) AddListOption(opts client.ListOptions) client.ListOptions {
	opts.LabelSelector = s.selector
	return opts
}

func (s *labelSelector) SetListFunc(f generic.ListFunc) generic.ListFunc {
	return f
}

func (s *labelSelector) Match(obj client.Object) bool {
	return true
}

func New(spec v1alpha1.GenericSelectorSpec) (generic.Selector, error) {
	labelSelectors := spec.LabelSelectors
	expressions := spec.ExpressionSelectors

	if len(labelSelectors) == 0 && len(expressions) == 0 {
		return &labelSelector{}, nil
	}
	metav1Ls := &metav1.LabelSelector{
		MatchLabels:      labelSelectors,
		MatchExpressions: expressions,
	}
	ls, err := metav1.LabelSelectorAsSelector(metav1Ls)
	if err != nil {
		return nil, err
	}
	return &labelSelector{ls}, nil
}
