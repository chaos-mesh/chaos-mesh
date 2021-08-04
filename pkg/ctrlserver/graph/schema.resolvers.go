package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	v11 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/generated"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/model"
)

func (r *hTTPChaosResolver) UID(ctx context.Context, obj *v1alpha1.HTTPChaos) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *hTTPChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.HTTPChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *hTTPChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.HTTPChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *hTTPChaosResolver) Labels(ctx context.Context, obj *v1alpha1.HTTPChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *hTTPChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.HTTPChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosResolver) UID(ctx context.Context, obj *v1alpha1.IOChaos) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.IOChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.IOChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosResolver) Labels(ctx context.Context, obj *v1alpha1.IOChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.IOChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Type(ctx context.Context, obj *v1alpha1.IOChaosAction) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Methods(ctx context.Context, obj *v1alpha1.IOChaosAction) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Ino(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Size(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Blocks(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Kind(ctx context.Context, obj *v1alpha1.IOChaosAction) (*string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Perm(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Nlink(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) UID(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Gid(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Rdev(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *iOChaosActionResolver) Filling(ctx context.Context, obj *v1alpha1.IOChaosAction) (*string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *ioFaultResolver) Errno(ctx context.Context, obj *v1alpha1.IoFault) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *loggerResolver) Component(ctx context.Context, ns string, component model.Component) (<-chan string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *loggerResolver) Pod(ctx context.Context, ns string, name string) (<-chan string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Component(ctx context.Context, obj *model.Namespace, component model.Component) (*v1.Pod, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Pod(ctx context.Context, obj *model.Namespace, name string) (*v1.Pod, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Pods(ctx context.Context, obj *model.Namespace) ([]*v1.Pod, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Stress(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.StressChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Stresses(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.StressChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Io(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.IOChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Ios(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.IOChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podio(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.PodIOChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podios(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.PodIOChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) HTTP(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.HTTPChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) HTTPS(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.HTTPChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podhttp(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.PodHttpChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podhttps(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.PodHttpChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Network(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.NetworkChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Networks(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.NetworkChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podnetwork(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.PodNetworkChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *namespaceResolver) Podnetworks(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.PodNetworkChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *networkChaosResolver) UID(ctx context.Context, obj *v1alpha1.NetworkChaos) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *networkChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.NetworkChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *networkChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.NetworkChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *networkChaosResolver) Labels(ctx context.Context, obj *v1alpha1.NetworkChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *networkChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.NetworkChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *ownerReferenceResolver) UID(ctx context.Context, obj *v11.OwnerReference) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podResolver) UID(ctx context.Context, obj *v1.Pod) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podResolver) CreationTimestamp(ctx context.Context, obj *v1.Pod) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podResolver) DeletionTimestamp(ctx context.Context, obj *v1.Pod) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podResolver) Labels(ctx context.Context, obj *v1.Pod) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podResolver) Annotations(ctx context.Context, obj *v1.Pod) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podHTTPChaosResolver) UID(ctx context.Context, obj *v1alpha1.PodHttpChaos) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podHTTPChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.PodHttpChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podHTTPChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.PodHttpChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podHTTPChaosResolver) Labels(ctx context.Context, obj *v1alpha1.PodHttpChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podHTTPChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.PodHttpChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podIOChaosResolver) UID(ctx context.Context, obj *v1alpha1.PodIOChaos) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podIOChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.PodIOChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podIOChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.PodIOChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podIOChaosResolver) Labels(ctx context.Context, obj *v1alpha1.PodIOChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podIOChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.PodIOChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podIOChaosResolver) Pod(ctx context.Context, obj *v1alpha1.PodIOChaos) (*v1.Pod, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podIOChaosResolver) Ios(ctx context.Context, obj *v1alpha1.PodIOChaos) ([]*v1alpha1.IOChaos, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podNetworkChaosResolver) UID(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podNetworkChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podNetworkChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podNetworkChaosResolver) Labels(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *podNetworkChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Namepsace(ctx context.Context, ns *string) (*model.Namespace, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *stressChaosResolver) UID(ctx context.Context, obj *v1alpha1.StressChaos) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *stressChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.StressChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *stressChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.StressChaos) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *stressChaosResolver) Labels(ctx context.Context, obj *v1alpha1.StressChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *stressChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.StressChaos) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

// HTTPChaos returns generated.HTTPChaosResolver implementation.
func (r *Resolver) HTTPChaos() generated.HTTPChaosResolver { return &hTTPChaosResolver{r} }

// IOChaos returns generated.IOChaosResolver implementation.
func (r *Resolver) IOChaos() generated.IOChaosResolver { return &iOChaosResolver{r} }

// IOChaosAction returns generated.IOChaosActionResolver implementation.
func (r *Resolver) IOChaosAction() generated.IOChaosActionResolver { return &iOChaosActionResolver{r} }

// IoFault returns generated.IoFaultResolver implementation.
func (r *Resolver) IoFault() generated.IoFaultResolver { return &ioFaultResolver{r} }

// Logger returns generated.LoggerResolver implementation.
func (r *Resolver) Logger() generated.LoggerResolver { return &loggerResolver{r} }

// Namespace returns generated.NamespaceResolver implementation.
func (r *Resolver) Namespace() generated.NamespaceResolver { return &namespaceResolver{r} }

// NetworkChaos returns generated.NetworkChaosResolver implementation.
func (r *Resolver) NetworkChaos() generated.NetworkChaosResolver { return &networkChaosResolver{r} }

// OwnerReference returns generated.OwnerReferenceResolver implementation.
func (r *Resolver) OwnerReference() generated.OwnerReferenceResolver {
	return &ownerReferenceResolver{r}
}

// Pod returns generated.PodResolver implementation.
func (r *Resolver) Pod() generated.PodResolver { return &podResolver{r} }

// PodHTTPChaos returns generated.PodHTTPChaosResolver implementation.
func (r *Resolver) PodHTTPChaos() generated.PodHTTPChaosResolver { return &podHTTPChaosResolver{r} }

// PodIOChaos returns generated.PodIOChaosResolver implementation.
func (r *Resolver) PodIOChaos() generated.PodIOChaosResolver { return &podIOChaosResolver{r} }

// PodNetworkChaos returns generated.PodNetworkChaosResolver implementation.
func (r *Resolver) PodNetworkChaos() generated.PodNetworkChaosResolver {
	return &podNetworkChaosResolver{r}
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// StressChaos returns generated.StressChaosResolver implementation.
func (r *Resolver) StressChaos() generated.StressChaosResolver { return &stressChaosResolver{r} }

type hTTPChaosResolver struct{ *Resolver }
type iOChaosResolver struct{ *Resolver }
type iOChaosActionResolver struct{ *Resolver }
type ioFaultResolver struct{ *Resolver }
type loggerResolver struct{ *Resolver }
type namespaceResolver struct{ *Resolver }
type networkChaosResolver struct{ *Resolver }
type ownerReferenceResolver struct{ *Resolver }
type podResolver struct{ *Resolver }
type podHTTPChaosResolver struct{ *Resolver }
type podIOChaosResolver struct{ *Resolver }
type podNetworkChaosResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type stressChaosResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *iOChaosActionResolver) Int64(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	panic(fmt.Errorf("not implemented"))
}
