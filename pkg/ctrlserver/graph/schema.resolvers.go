package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"bufio"
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	v11 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/generated"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/model"
)

func (r *hTTPChaosResolver) UID(ctx context.Context, obj *v1alpha1.HTTPChaos) (string, error) {
	return string(obj.UID), nil
}

func (r *hTTPChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.HTTPChaos) (*time.Time, error) {
	return &obj.CreationTimestamp.Time, nil
}

func (r *hTTPChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.HTTPChaos) (*time.Time, error) {
	return &obj.DeletionTimestamp.Time, nil
}

func (r *hTTPChaosResolver) Labels(ctx context.Context, obj *v1alpha1.HTTPChaos) (map[string]interface{}, error) {
	labels := make(map[string]interface{})
	for k, v := range obj.Labels {
		labels[k] = v
	}
	return labels, nil
}

func (r *hTTPChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.HTTPChaos) (map[string]interface{}, error) {
	annotations := make(map[string]interface{})
	for k, v := range obj.Annotations {
		annotations[k] = v
	}
	return annotations, nil
}

func (r *iOChaosResolver) UID(ctx context.Context, obj *v1alpha1.IOChaos) (string, error) {
	return string(obj.UID), nil
}

func (r *iOChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.IOChaos) (*time.Time, error) {
	return &obj.CreationTimestamp.Time, nil
}

func (r *iOChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.IOChaos) (*time.Time, error) {
	return &obj.DeletionTimestamp.Time, nil
}

func (r *iOChaosResolver) Labels(ctx context.Context, obj *v1alpha1.IOChaos) (map[string]interface{}, error) {
	labels := make(map[string]interface{})
	for k, v := range obj.Labels {
		labels[k] = v
	}
	return labels, nil
}

func (r *iOChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.IOChaos) (map[string]interface{}, error) {
	annotations := make(map[string]interface{})
	for k, v := range obj.Annotations {
		annotations[k] = v
	}
	return annotations, nil
}

func (r *iOChaosActionResolver) Type(ctx context.Context, obj *v1alpha1.IOChaosAction) (string, error) {
	return string(obj.Type), nil
}

func (r *iOChaosActionResolver) Methods(ctx context.Context, obj *v1alpha1.IOChaosAction) ([]string, error) {
	methods := make([]string, 0, len(obj.Methods))
	for k, v := range obj.Methods {
		methods[k] = string(v)
	}
	return methods, nil
}

func (r *iOChaosActionResolver) Ino(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	if obj.Ino == nil {
		return nil, nil
	}
	ino := (int)(*obj.Ino)
	return &ino, nil
}

func (r *iOChaosActionResolver) Size(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	if obj.Size == nil {
		return nil, nil
	}
	size := (int)(*obj.Size)
	return &size, nil
}

func (r *iOChaosActionResolver) Blocks(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	if obj.Blocks == nil {
		return nil, nil
	}
	blocks := (int)(*obj.Blocks)
	return &blocks, nil
}

func (r *iOChaosActionResolver) Kind(ctx context.Context, obj *v1alpha1.IOChaosAction) (*string, error) {
	if obj.Kind == nil {
		return nil, nil
	}
	kind := (string)(*obj.Kind)
	return &kind, nil
}

func (r *iOChaosActionResolver) Perm(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	if obj.Perm == nil {
		return nil, nil
	}
	perm := (int)(*obj.Perm)
	return &perm, nil
}

func (r *iOChaosActionResolver) Nlink(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	if obj.Nlink == nil {
		return nil, nil
	}
	nlink := (int)(*obj.Nlink)
	return &nlink, nil
}

func (r *iOChaosActionResolver) UID(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	if obj.UID == nil {
		return nil, nil
	}
	uid := (int)(*obj.UID)
	return &uid, nil
}

func (r *iOChaosActionResolver) Gid(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	if obj.GID == nil {
		return nil, nil
	}
	gid := (int)(*obj.GID)
	return &gid, nil
}

func (r *iOChaosActionResolver) Rdev(ctx context.Context, obj *v1alpha1.IOChaosAction) (*int, error) {
	if obj.Rdev == nil {
		return nil, nil
	}
	rdev := (int)(*obj.Rdev)
	return &rdev, nil
}

func (r *iOChaosActionResolver) Filling(ctx context.Context, obj *v1alpha1.IOChaosAction) (*string, error) {
	filling := string(obj.Filling)
	return &filling, nil
}

func (r *ioFaultResolver) Errno(ctx context.Context, obj *v1alpha1.IoFault) (int, error) {
	return int(obj.Errno), nil
}

func (r *loggerResolver) Component(ctx context.Context, ns string, component model.Component) (<-chan string, error) {
	var list v1.PodList
	if err := r.Client.List(ctx, &list, client.MatchingLabels(componentLabels(component))); err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("instance of %s not found", component)
	}

	return r.Pod(ctx, list.Items[0].Namespace, list.Items[0].Name)
}

func (r *loggerResolver) Pod(ctx context.Context, ns string, name string) (<-chan string, error) {
	logs, err := r.Clientset.CoreV1().Pods(ns).GetLogs(name, &v1.PodLogOptions{Follow: true}).Stream()
	if err != nil {
		return nil, err
	}
	logChan := make(chan string)
	go func() {
		defer logs.Close()
		reader := bufio.NewReader(logs)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				r.Log.Error(err, fmt.Sprintf("fail to read log of pod(%s/%s)", ns, name))
				break
			}
			select {
			case logChan <- string(line):
				continue
			case <-time.NewTimer(time.Minute).C:
				r.Log.Info(fmt.Sprintf("client has not read log of pod(%s/%s) for 1m, close channel", ns, name))
				break
			}
		}
	}()
	return logChan, nil
}

func (r *namespaceResolver) Component(ctx context.Context, obj *model.Namespace, component model.Component) ([]*v1.Pod, error) {
	var list v1.PodList
	var pods []*v1.Pod
	if err := r.Client.List(ctx, &list, client.MatchingLabels(componentLabels(component))); err != nil {
		return nil, err
	}
	for i := range list.Items {
		pods = append(pods, &list.Items[i])
	}
	return pods, nil
}

func (r *namespaceResolver) Pod(ctx context.Context, obj *model.Namespace, name string) (*v1.Pod, error) {
	key := types.NamespacedName{Namespace: obj.Ns, Name: name}
	pod := new(v1.Pod)
	if err := r.Client.Get(ctx, key, pod); err != nil {
		return nil, err
	}
	return pod, nil
}

func (r *namespaceResolver) Pods(ctx context.Context, obj *model.Namespace) ([]*v1.Pod, error) {
	var podList v1.PodList
	var pods []*v1.Pod
	if err := r.Client.List(ctx, &podList, &client.ListOptions{Namespace: obj.Ns}); err != nil {
		return nil, err
	}

	for i := range podList.Items {
		pods = append(pods, &podList.Items[i])
	}

	return pods, nil
}

func (r *namespaceResolver) Stress(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.StressChaos, error) {
	key := types.NamespacedName{Namespace: obj.Ns, Name: name}
	stress := new(v1alpha1.StressChaos)
	if err := r.Client.Get(ctx, key, stress); err != nil {
		return nil, err
	}
	return stress, nil
}

func (r *namespaceResolver) Stresses(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.StressChaos, error) {
	var stressList v1alpha1.StressChaosList
	var stresses []*v1alpha1.StressChaos
	if err := r.Client.List(ctx, &stressList, &client.ListOptions{Namespace: obj.Ns}); err != nil {
		return nil, err
	}

	for i := range stressList.Items {
		stresses = append(stresses, &stressList.Items[i])
	}

	return stresses, nil
}

func (r *namespaceResolver) Io(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.IOChaos, error) {
	key := types.NamespacedName{Namespace: obj.Ns, Name: name}
	io := new(v1alpha1.IOChaos)
	if err := r.Client.Get(ctx, key, io); err != nil {
		return nil, err
	}
	return io, nil
}

func (r *namespaceResolver) Ios(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.IOChaos, error) {
	var ioList v1alpha1.IOChaosList
	var ios []*v1alpha1.IOChaos
	if err := r.Client.List(ctx, &ioList, &client.ListOptions{Namespace: obj.Ns}); err != nil {
		return nil, err
	}

	for i := range ioList.Items {
		ios = append(ios, &ioList.Items[i])
	}

	return ios, nil
}

func (r *namespaceResolver) Podio(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.PodIOChaos, error) {
	key := types.NamespacedName{Namespace: obj.Ns, Name: name}
	io := new(v1alpha1.PodIOChaos)
	if err := r.Client.Get(ctx, key, io); err != nil {
		return nil, err
	}
	return io, nil
}

func (r *namespaceResolver) Podios(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.PodIOChaos, error) {
	var ioList v1alpha1.PodIOChaosList
	var ios []*v1alpha1.PodIOChaos
	if err := r.Client.List(ctx, &ioList, &client.ListOptions{Namespace: obj.Ns}); err != nil {
		return nil, err
	}

	for i := range ioList.Items {
		ios = append(ios, &ioList.Items[i])
	}

	return ios, nil
}

func (r *namespaceResolver) HTTP(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.HTTPChaos, error) {
	key := types.NamespacedName{Namespace: obj.Ns, Name: name}
	http := new(v1alpha1.HTTPChaos)
	if err := r.Client.Get(ctx, key, http); err != nil {
		return nil, err
	}
	return http, nil
}

func (r *namespaceResolver) HTTPS(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.HTTPChaos, error) {
	var httpList v1alpha1.HTTPChaosList
	var https []*v1alpha1.HTTPChaos
	if err := r.Client.List(ctx, &httpList, &client.ListOptions{Namespace: obj.Ns}); err != nil {
		return nil, err
	}

	for i := range httpList.Items {
		https = append(https, &httpList.Items[i])
	}

	return https, nil
}

func (r *namespaceResolver) Podhttp(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.PodHttpChaos, error) {
	key := types.NamespacedName{Namespace: obj.Ns, Name: name}
	http := new(v1alpha1.PodHttpChaos)
	if err := r.Client.Get(ctx, key, http); err != nil {
		return nil, err
	}
	return http, nil
}

func (r *namespaceResolver) Podhttps(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.PodHttpChaos, error) {
	var httpList v1alpha1.PodHttpChaosList
	var https []*v1alpha1.PodHttpChaos
	if err := r.Client.List(ctx, &httpList, &client.ListOptions{Namespace: obj.Ns}); err != nil {
		return nil, err
	}

	for i := range httpList.Items {
		https = append(https, &httpList.Items[i])
	}

	return https, nil
}

func (r *namespaceResolver) Network(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.NetworkChaos, error) {
	key := types.NamespacedName{Namespace: obj.Ns, Name: name}
	network := new(v1alpha1.NetworkChaos)
	if err := r.Client.Get(ctx, key, network); err != nil {
		return nil, err
	}
	return network, nil
}

func (r *namespaceResolver) Networks(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.NetworkChaos, error) {
	var networkList v1alpha1.NetworkChaosList
	var networks []*v1alpha1.NetworkChaos
	if err := r.Client.List(ctx, &networkList, &client.ListOptions{Namespace: obj.Ns}); err != nil {
		return nil, err
	}

	for i := range networkList.Items {
		networks = append(networks, &networkList.Items[i])
	}

	return networks, nil
}

func (r *namespaceResolver) Podnetwork(ctx context.Context, obj *model.Namespace, name string) (*v1alpha1.PodNetworkChaos, error) {
	key := types.NamespacedName{Namespace: obj.Ns, Name: name}
	network := new(v1alpha1.PodNetworkChaos)
	if err := r.Client.Get(ctx, key, network); err != nil {
		return nil, err
	}
	return network, nil
}

func (r *namespaceResolver) Podnetworks(ctx context.Context, obj *model.Namespace) ([]*v1alpha1.PodNetworkChaos, error) {
	var networkList v1alpha1.PodNetworkChaosList
	var networks []*v1alpha1.PodNetworkChaos
	if err := r.Client.List(ctx, &networkList, &client.ListOptions{Namespace: obj.Ns}); err != nil {
		return nil, err
	}

	for i := range networkList.Items {
		networks = append(networks, &networkList.Items[i])
	}

	return networks, nil
}

func (r *networkChaosResolver) UID(ctx context.Context, obj *v1alpha1.NetworkChaos) (string, error) {
	return string(obj.UID), nil
}

func (r *networkChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.NetworkChaos) (*time.Time, error) {
	return &obj.CreationTimestamp.Time, nil
}

func (r *networkChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.NetworkChaos) (*time.Time, error) {
	return &obj.DeletionTimestamp.Time, nil
}

func (r *networkChaosResolver) Labels(ctx context.Context, obj *v1alpha1.NetworkChaos) (map[string]interface{}, error) {
	labels := make(map[string]interface{})
	for k, v := range obj.Labels {
		labels[k] = v
	}
	return labels, nil
}

func (r *networkChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.NetworkChaos) (map[string]interface{}, error) {
	annotations := make(map[string]interface{})
	for k, v := range obj.Annotations {
		annotations[k] = v
	}
	return annotations, nil
}

func (r *ownerReferenceResolver) UID(ctx context.Context, obj *v11.OwnerReference) (string, error) {
	return string(obj.UID), nil
}

func (r *podResolver) UID(ctx context.Context, obj *v1.Pod) (string, error) {
	return string(obj.UID), nil
}

func (r *podResolver) CreationTimestamp(ctx context.Context, obj *v1.Pod) (*time.Time, error) {
	return &obj.CreationTimestamp.Time, nil
}

func (r *podResolver) DeletionTimestamp(ctx context.Context, obj *v1.Pod) (*time.Time, error) {
	return &obj.DeletionTimestamp.Time, nil
}

func (r *podResolver) Labels(ctx context.Context, obj *v1.Pod) (map[string]interface{}, error) {
	labels := make(map[string]interface{})
	for k, v := range obj.Labels {
		labels[k] = v
	}
	return labels, nil
}

func (r *podResolver) Annotations(ctx context.Context, obj *v1.Pod) (map[string]interface{}, error) {
	annotations := make(map[string]interface{})
	for k, v := range obj.Annotations {
		annotations[k] = v
	}
	return annotations, nil
}

func (r *podHTTPChaosResolver) UID(ctx context.Context, obj *v1alpha1.PodHttpChaos) (string, error) {
	return string(obj.UID), nil
}

func (r *podHTTPChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.PodHttpChaos) (*time.Time, error) {
	return &obj.CreationTimestamp.Time, nil
}

func (r *podHTTPChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.PodHttpChaos) (*time.Time, error) {
	return &obj.DeletionTimestamp.Time, nil
}

func (r *podHTTPChaosResolver) Labels(ctx context.Context, obj *v1alpha1.PodHttpChaos) (map[string]interface{}, error) {
	labels := make(map[string]interface{})
	for k, v := range obj.Labels {
		labels[k] = v
	}
	return labels, nil
}

func (r *podHTTPChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.PodHttpChaos) (map[string]interface{}, error) {
	annotations := make(map[string]interface{})
	for k, v := range obj.Annotations {
		annotations[k] = v
	}
	return annotations, nil
}

func (r *podIOChaosResolver) UID(ctx context.Context, obj *v1alpha1.PodIOChaos) (string, error) {
	return string(obj.UID), nil
}

func (r *podIOChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.PodIOChaos) (*time.Time, error) {
	return &obj.CreationTimestamp.Time, nil
}

func (r *podIOChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.PodIOChaos) (*time.Time, error) {
	return &obj.DeletionTimestamp.Time, nil
}

func (r *podIOChaosResolver) Labels(ctx context.Context, obj *v1alpha1.PodIOChaos) (map[string]interface{}, error) {
	labels := make(map[string]interface{})
	for k, v := range obj.Labels {
		labels[k] = v
	}
	return labels, nil
}

func (r *podIOChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.PodIOChaos) (map[string]interface{}, error) {
	annotations := make(map[string]interface{})
	for k, v := range obj.Annotations {
		annotations[k] = v
	}
	return annotations, nil
}

func (r *podIOChaosResolver) Pod(ctx context.Context, obj *v1alpha1.PodIOChaos) (*v1.Pod, error) {
	pod := new(v1.Pod)
	key := types.NamespacedName{Namespace: obj.Namespace, Name: obj.Name}
	if err := r.Client.Get(ctx, key, pod); err != nil {
		return nil, err
	}
	return pod, nil
}

func (r *podIOChaosResolver) Ios(ctx context.Context, obj *v1alpha1.PodIOChaos) ([]*v1alpha1.IOChaos, error) {
	ioNames := make(map[string]bool)
	for _, action := range obj.Spec.Actions {
		ioNames[action.Source] = true
	}

	ios := make([]*v1alpha1.IOChaos, 0, len(ioNames))
	for name := range ioNames {
		namespaced := parseNamespacedName(name)
		io := new(v1alpha1.IOChaos)
		if err := r.Client.Get(ctx, namespaced, io); err != nil {
			return nil, err
		}
		ios = append(ios, io)
	}
	return ios, nil
}

func (r *podNetworkChaosResolver) UID(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (string, error) {
	return string(obj.UID), nil
}

func (r *podNetworkChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (*time.Time, error) {
	return &obj.CreationTimestamp.Time, nil
}

func (r *podNetworkChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (*time.Time, error) {
	return &obj.DeletionTimestamp.Time, nil
}

func (r *podNetworkChaosResolver) Labels(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (map[string]interface{}, error) {
	labels := make(map[string]interface{})
	for k, v := range obj.Labels {
		labels[k] = v
	}
	return labels, nil
}

func (r *podNetworkChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (map[string]interface{}, error) {
	annotations := make(map[string]interface{})
	for k, v := range obj.Annotations {
		annotations[k] = v
	}
	return annotations, nil
}

func (r *queryResolver) Namepsace(ctx context.Context, ns string) (*model.Namespace, error) {
	return &model.Namespace{Ns: ns}, nil
}

func (r *stressChaosResolver) UID(ctx context.Context, obj *v1alpha1.StressChaos) (string, error) {
	return string(obj.UID), nil
}

func (r *stressChaosResolver) CreationTimestamp(ctx context.Context, obj *v1alpha1.StressChaos) (*time.Time, error) {
	return &obj.CreationTimestamp.Time, nil
}

func (r *stressChaosResolver) DeletionTimestamp(ctx context.Context, obj *v1alpha1.StressChaos) (*time.Time, error) {
	return &obj.DeletionTimestamp.Time, nil
}

func (r *stressChaosResolver) Labels(ctx context.Context, obj *v1alpha1.StressChaos) (map[string]interface{}, error) {
	labels := make(map[string]interface{})
	for k, v := range obj.Labels {
		labels[k] = v
	}
	return labels, nil
}

func (r *stressChaosResolver) Annotations(ctx context.Context, obj *v1alpha1.StressChaos) (map[string]interface{}, error) {
	annotations := make(map[string]interface{})
	for k, v := range obj.Annotations {
		annotations[k] = v
	}
	return annotations, nil
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
