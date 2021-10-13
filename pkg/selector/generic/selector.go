package generic

import (
	"context"
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
	List
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
