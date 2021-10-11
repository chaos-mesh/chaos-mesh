package generic

import (
	"context"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const InjectAnnotationKey = "chaos-mesh.org/inject"

type Option struct {
	ClusterScoped         bool
	TargetNamespace       string
	EnableFilterNamespace bool
}

type ListFunc func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error

type List interface {
	AddListOption(client.ListOptions) (client.ListOptions, error)
	SetListFunc(ListFunc) ListFunc
}

type Selector interface {
	Match(client.Object) (bool, error)
}

func ListObjects(c client.Client, lists []List) {
	opts := client.ListOptions{}
	listF := c.List

	var err error
	for _, list := range lists {
		opts, err = list.AddListOption(opts)
		if err != nil {
			return
		}
		listF = list.SetListFunc(listF)
	}

}

func Filter(objs []client.Object, selectors []Selector) ([]client.Object, error) {
	filterObjs := make([]client.Object, 0, len(objs))

	var err error
	for _, obj := range objs {
		var ok bool
		for _, selector := range selectors {
			ok, err = selector.Match(obj)
			if err != nil {
				return nil, err
			}
			if !ok {
				break
			}
		}
		if ok {
			filterObjs = append(filterObjs, obj)
		}
	}
	return filterObjs, nil
}

type labelSelector struct {
	selector labels.Selector
}

func (s *labelSelector) AddListOption(opts client.ListOptions) (client.ListOptions, error) {
	opts.LabelSelector = s.selector
	return opts, nil
}

func (s *labelSelector) SetListFunc(f ListFunc) ListFunc {
	return f
}

func (s *labelSelector) Match(obj client.Object) (bool, error) {
	return true, nil
}

func ParseLabelSelector(labels map[string]string, expressions v1alpha1.LabelSelectorRequirements) (Selector, error) {
	if len(labels) == 0 && len(expressions) == 0 {
		return &labelSelector{}, nil
	}
	metav1Ls := &metav1.LabelSelector{
		MatchLabels:      labels,
		MatchExpressions: expressions,
	}
	ls, err := metav1.LabelSelectorAsSelector(metav1Ls)
	if err != nil {
		return nil, err
	}
	return &labelSelector{ls}, nil
}

type fieldSelector struct {
	FieldSelectors map[string]string
	r              client.Reader
}

func (s *fieldSelector) AddListOption(opts client.ListOptions) (client.ListOptions, error) {
	if len(s.FieldSelectors) > 0 {
		opts.FieldSelector = fields.SelectorFromSet(s.FieldSelectors)
	}
	return opts, nil
}

func (s *fieldSelector) SetListFunc(f ListFunc) ListFunc {
	// Since FieldSelectors need to implement index creation, Reader.List is used to get the pod list.
	// Otherwise, just call Client.List directly, which can be obtained through cache.
	if len(s.FieldSelectors) > 0 && s.r != nil {
		return s.r.List
	}
	return f
}

func (s *fieldSelector) Match(obj client.Object) (bool, error) {
	return true, nil
}

func ParseFieldSelector() (Selector, error) {

	return &fieldSelector{}, nil
}

type annotationSelector struct {
	labels.Selector
}

func (s *annotationSelector) AddListOption(opts client.ListOptions) (client.ListOptions, error) {
	return opts, nil
}

func (s *annotationSelector) SetListFunc(f ListFunc) ListFunc {
	return f
}

func (s *annotationSelector) Match(obj client.Object) (bool, error) {
	// TODO fix
	if s.Empty() {
		return true, nil
	}
	annotations := labels.Set(obj.GetAnnotations())
	return s.Matches(annotations), nil
}

func ParseAnnotationSelector(selectors map[string]string) (Selector, error) {
	selectorStr := label.Label(selectors).String()
	s, err := labels.Parse(selectorStr)
	if err != nil {
		return nil, err
	}
	return &annotationSelector{s}, nil
}

type namespaceSelector struct {
	Option
}

func (s *namespaceSelector) AddListOption(opts client.ListOptions) (client.ListOptions, error) {
	if !s.ClusterScoped {
		opts.Namespace = s.TargetNamespace
	}
	return opts, nil
}

func (s *namespaceSelector) SetListFunc(f ListFunc) ListFunc {
	return f
}

func (s *namespaceSelector) Match(obj client.Object) (bool, error) {

	return false, nil
}

func ParseNamespaceSelector(nodeNames []string, selectors map[string]string,
	clusterScoped bool, targetNamespace string, enableFilterNamespace bool) (Selector, error) {
	if !clusterScoped {
		if len(nodeNames) > 1 {
			return nil, fmt.Errorf("could NOT use more than 1 namespace selector within namespace scoped mode")
		} else if len(nodeNames) == 1 {
			if nodeNames[0] != targetNamespace {
				return nil, fmt.Errorf("could NOT list pods from out of scoped namespace: %s", nodeNames[0])
			}
		}
	}
	return &namespaceSelector{
		Option{
			clusterScoped,
			targetNamespace,
			enableFilterNamespace,
		},
	}, nil
}
