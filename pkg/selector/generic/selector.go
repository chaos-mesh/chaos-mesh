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

type Selector interface {
	List(ListFunc, client.ListOptions, func(listFunc ListFunc, opts client.ListOptions) error) error
	Match(client.Object) bool
	Next(Selector)
}
