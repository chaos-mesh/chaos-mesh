package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"time"

	v11 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/generated"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/model"
)

func (r *loggerResolver) Component(ctx context.Context, ns string, component model.Component) (<-chan string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *loggerResolver) Pod(ctx context.Context, ns string, name string) (<-chan string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Component(ctx context.Context, obj *model.Namespace, component model.Component) (*v11.Pod, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Pod(ctx context.Context, obj *model.Namespace, name string) (*v11.Pod, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Pods(ctx context.Context, obj *model.Namespace) ([]*v11.Pod, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Stress(ctx context.Context, obj *model.Namespace, name string) (*model.StressChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Stresses(ctx context.Context, obj *model.Namespace) ([]*model.StressChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Io(ctx context.Context, obj *model.Namespace, name string) (*model.IOChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Ios(ctx context.Context, obj *model.Namespace) ([]*model.IOChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podio(ctx context.Context, obj *model.Namespace, name string) (*model.PodIOChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podios(ctx context.Context, obj *model.Namespace) ([]*model.PodIOChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) HTTP(ctx context.Context, obj *model.Namespace, name string) (*model.HTTPChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) HTTPS(ctx context.Context, obj *model.Namespace) ([]*model.HTTPChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podhttp(ctx context.Context, obj *model.Namespace, name string) (*model.PodHTTPChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podhttps(ctx context.Context, obj *model.Namespace) ([]*model.PodHTTPChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Network(ctx context.Context, obj *model.Namespace, name string) (*model.NetworkChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Networks(ctx context.Context, obj *model.Namespace) ([]*model.NetworkChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podnetwork(ctx context.Context, obj *model.Namespace, name string) (*model.PodNetWorkChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podnetworks(ctx context.Context, obj *model.Namespace) ([]*model.PodNetWorkChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *ownerReferenceResolver) UID(ctx context.Context, obj *v1.OwnerReference) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podResolver) UID(ctx context.Context, obj *v11.Pod) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podResolver) CreationTimestamp(ctx context.Context, obj *v11.Pod) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podResolver) DeletionTimestamp(ctx context.Context, obj *v11.Pod) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podResolver) Labels(ctx context.Context, obj *v11.Pod) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podResolver) Annotations(ctx context.Context, obj *v11.Pod) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Namepsace(ctx context.Context, ns *string) (*model.Namespace, error) {
	panic(fmt.Errorf("not implemented"))
}

// Logger returns generated.LoggerResolver implementation.
func (r *Resolver) Logger() generated.LoggerResolver { return &loggerResolver{r} }

// Namespace returns generated.NamespaceResolver implementation.
func (r *Resolver) Namespace() generated.NamespaceResolver { return &namespaceResolver{r} }

// OwnerReference returns generated.OwnerReferenceResolver implementation.
func (r *Resolver) OwnerReference() generated.OwnerReferenceResolver {
	return &ownerReferenceResolver{r}
}

// Pod returns generated.PodResolver implementation.
func (r *Resolver) Pod() generated.PodResolver { return &podResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type loggerResolver struct{ *Resolver }
type namespaceResolver struct{ *Resolver }
type ownerReferenceResolver struct{ *Resolver }
type podResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
