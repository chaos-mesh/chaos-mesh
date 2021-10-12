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
	AddListOption(client.ListOptions) client.ListOptions
	SetListFunc(ListFunc) ListFunc
}

type Selector interface {
	Match(client.Object) bool
}

func ListObjects(c client.Client, lists []List,
	listObj func(listFunc ListFunc, opts client.ListOptions) error) error {
	opts := client.ListOptions{}
	listF := c.List

	for _, list := range lists {
		opts = list.AddListOption(opts)
		listF = list.SetListFunc(listF)
	}

	return listObj(listF, opts)
}

func Filter(objs []client.Object, selectors []Selector) ([]client.Object, error) {
	filterObjs := make([]client.Object, 0, len(objs))

	for _, obj := range objs {
		var ok bool
		for _, selector := range selectors {
			ok = selector.Match(obj)
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

func (s *labelSelector) AddListOption(opts client.ListOptions) client.ListOptions {
	opts.LabelSelector = s.selector
	return opts
}

func (s *labelSelector) SetListFunc(f ListFunc) ListFunc {
	return f
}

func (s *labelSelector) Match(obj client.Object) bool {
	return true
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

func (s *fieldSelector) AddListOption(opts client.ListOptions) client.ListOptions {
	if len(s.FieldSelectors) > 0 {
		opts.FieldSelector = fields.SelectorFromSet(s.FieldSelectors)
	}
	return opts
}

func (s *fieldSelector) SetListFunc(f ListFunc) ListFunc {
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

func ParseFieldSelector() (Selector, error) {

	return &fieldSelector{}, nil
}

type annotationSelector struct {
	labels.Selector
}

func (s *annotationSelector) AddListOption(opts client.ListOptions) client.ListOptions {
	return opts
}

func (s *annotationSelector) SetListFunc(f ListFunc) ListFunc {
	return f
}

func (s *annotationSelector) Match(obj client.Object) bool {
	// TODO fix
	if s.Empty() {
		return true
	}
	annotations := labels.Set(obj.GetAnnotations())
	return s.Matches(annotations)
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

func (s *namespaceSelector) AddListOption(opts client.ListOptions) client.ListOptions {
	if !s.ClusterScoped {
		opts.Namespace = s.TargetNamespace
	}
	return opts
}

func (s *namespaceSelector) SetListFunc(f ListFunc) ListFunc {
	return f
}

func (s *namespaceSelector) Match(obj client.Object) bool {
	return true
}

func ParseNamespaceSelector(namespaces []string,
	clusterScoped bool, targetNamespace string, enableFilterNamespace bool) (Selector, error) {
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
		Option{
			clusterScoped,
			targetNamespace,
			enableFilterNamespace,
		},
	}, nil
}
