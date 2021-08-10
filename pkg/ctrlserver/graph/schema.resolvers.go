package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	v1 "k8s.io/api/core/v1"
	v11 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/generated"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/model"
)

func (r *attrOverrideSpecResolver) Ino(ctx context.Context, obj *v1alpha1.AttrOverrideSpec) (*int, error) {
	if obj.Ino == nil {
		return nil, nil
	}
	ino := (int)(*obj.Ino)
	return &ino, nil
}

func (r *attrOverrideSpecResolver) Size(ctx context.Context, obj *v1alpha1.AttrOverrideSpec) (*int, error) {
	if obj.Size == nil {
		return nil, nil
	}
	size := (int)(*obj.Size)
	return &size, nil
}

func (r *attrOverrideSpecResolver) Blocks(ctx context.Context, obj *v1alpha1.AttrOverrideSpec) (*int, error) {
	if obj.Blocks == nil {
		return nil, nil
	}
	blocks := (int)(*obj.Blocks)
	return &blocks, nil
}

func (r *attrOverrideSpecResolver) Kind(ctx context.Context, obj *v1alpha1.AttrOverrideSpec) (*string, error) {
	if obj.Kind == nil {
		return nil, nil
	}
	kind := (string)(*obj.Kind)
	return &kind, nil
}

func (r *attrOverrideSpecResolver) Perm(ctx context.Context, obj *v1alpha1.AttrOverrideSpec) (*int, error) {
	if obj.Perm == nil {
		return nil, nil
	}
	perm := (int)(*obj.Perm)
	return &perm, nil
}

func (r *attrOverrideSpecResolver) Nlink(ctx context.Context, obj *v1alpha1.AttrOverrideSpec) (*int, error) {
	if obj.Nlink == nil {
		return nil, nil
	}
	nlink := (int)(*obj.Nlink)
	return &nlink, nil
}

func (r *attrOverrideSpecResolver) UID(ctx context.Context, obj *v1alpha1.AttrOverrideSpec) (*int, error) {
	if obj.UID == nil {
		return nil, nil
	}
	uid := (int)(*obj.UID)
	return &uid, nil
}

func (r *attrOverrideSpecResolver) Gid(ctx context.Context, obj *v1alpha1.AttrOverrideSpec) (*int, error) {
	if obj.GID == nil {
		return nil, nil
	}
	gid := (int)(*obj.GID)
	return &gid, nil
}

func (r *attrOverrideSpecResolver) Rdev(ctx context.Context, obj *v1alpha1.AttrOverrideSpec) (*int, error) {
	if obj.Rdev == nil {
		return nil, nil
	}
	rdev := (int)(*obj.Rdev)
	return &rdev, nil
}

func (r *bandwidthSpecResolver) Limit(ctx context.Context, obj *v1alpha1.BandwidthSpec) (int, error) {
	return int(obj.Limit), nil
}

func (r *bandwidthSpecResolver) Buffer(ctx context.Context, obj *v1alpha1.BandwidthSpec) (int, error) {
	return int(obj.Buffer), nil
}

func (r *bandwidthSpecResolver) Peakrate(ctx context.Context, obj *v1alpha1.BandwidthSpec) (*int, error) {
	if obj.Peakrate == nil {
		return nil, nil
	}
	value := int(*obj.Peakrate)
	return &value, nil
}

func (r *bandwidthSpecResolver) Minburst(ctx context.Context, obj *v1alpha1.BandwidthSpec) (*int, error) {
	if obj.Minburst == nil {
		return nil, nil
	}
	value := int(*obj.Minburst)
	return &value, nil
}

func (r *chaosConditionResolver) Type(ctx context.Context, obj *v1alpha1.ChaosCondition) (string, error) {
	return string(obj.Type), nil
}

func (r *chaosConditionResolver) Status(ctx context.Context, obj *v1alpha1.ChaosCondition) (string, error) {
	return string(obj.Status), nil
}

func (r *containerStateRunningResolver) StartedAt(ctx context.Context, obj *v1.ContainerStateRunning) (*time.Time, error) {
	return &obj.StartedAt.Time, nil
}

func (r *containerStateTerminatedResolver) StartedAt(ctx context.Context, obj *v1.ContainerStateTerminated) (*time.Time, error) {
	return &obj.StartedAt.Time, nil
}

func (r *containerStateTerminatedResolver) FinishedAt(ctx context.Context, obj *v1.ContainerStateTerminated) (*time.Time, error) {
	return &obj.FinishedAt.Time, nil
}

func (r *corruptSpecResolver) Corrup(ctx context.Context, obj *v1alpha1.CorruptSpec) (string, error) {
	return obj.Corrupt, nil
}

func (r *experimentStatusResolver) DesiredPhase(ctx context.Context, obj *v1alpha1.ExperimentStatus) (string, error) {
	return string(obj.DesiredPhase), nil
}

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

func (r *hTTPChaosResolver) Podhttp(ctx context.Context, obj *v1alpha1.HTTPChaos) ([]*v1alpha1.PodHttpChaos, error) {
	podhttps := make([]*v1alpha1.PodHttpChaos, 0, len(obj.Status.Instances))
	for id := range obj.Status.Instances {
		podhttp := new(v1alpha1.PodHttpChaos)
		if err := r.Client.Get(ctx, parseNamespacedName(id), podhttp); err != nil {
			return nil, err
		}
		podhttps = append(podhttps, podhttp)
	}
	return podhttps, nil
}

func (r *hTTPChaosSpecResolver) Mode(ctx context.Context, obj *v1alpha1.HTTPChaosSpec) (string, error) {
	return string(obj.Mode), nil
}

func (r *hTTPChaosSpecResolver) Target(ctx context.Context, obj *v1alpha1.HTTPChaosSpec) (string, error) {
	return string(obj.Target), nil
}

func (r *hTTPChaosSpecResolver) RequestHeaders(ctx context.Context, obj *v1alpha1.HTTPChaosSpec) (map[string]interface{}, error) {
	headers := make(map[string]interface{})
	for k, v := range obj.RequestHeaders {
		headers[k] = v
	}
	return headers, nil
}

func (r *hTTPChaosSpecResolver) ResponseHeaders(ctx context.Context, obj *v1alpha1.HTTPChaosSpec) (map[string]interface{}, error) {
	headers := make(map[string]interface{})
	for k, v := range obj.ResponseHeaders {
		headers[k] = v
	}
	return headers, nil
}

func (r *hTTPChaosStatusResolver) Instances(ctx context.Context, obj *v1alpha1.HTTPChaosStatus) (map[string]interface{}, error) {
	instances := make(map[string]interface{})
	for k, v := range obj.Instances {
		instances[k] = v
	}
	return instances, nil
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

func (r *iOChaosResolver) Podios(ctx context.Context, obj *v1alpha1.IOChaos) ([]*v1alpha1.PodIOChaos, error) {
	podios := make([]*v1alpha1.PodIOChaos, 0, len(obj.Status.Instances))
	for id := range obj.Status.Instances {
		podio := new(v1alpha1.PodIOChaos)
		if err := r.Client.Get(ctx, parseNamespacedName(id), podio); err != nil {
			return nil, err
		}
		podios = append(podios, podio)
	}
	return podios, nil
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

func (r *iOChaosSpecResolver) Mode(ctx context.Context, obj *v1alpha1.IOChaosSpec) (string, error) {
	return string(obj.Mode), nil
}

func (r *iOChaosSpecResolver) Action(ctx context.Context, obj *v1alpha1.IOChaosSpec) (string, error) {
	return string(obj.Action), nil
}

func (r *iOChaosSpecResolver) Errno(ctx context.Context, obj *v1alpha1.IOChaosSpec) (*int, error) {
	errno := int(obj.Errno)
	return &errno, nil
}

func (r *iOChaosSpecResolver) Methods(ctx context.Context, obj *v1alpha1.IOChaosSpec) ([]string, error) {
	methods := make([]string, 0, len(obj.Methods))
	for _, method := range obj.Methods {
		methods = append(methods, string(method))
	}
	return methods, nil
}

func (r *iOChaosStatusResolver) Instances(ctx context.Context, obj *v1alpha1.IOChaosStatus) (map[string]interface{}, error) {
	instances := make(map[string]interface{})
	for k, v := range obj.Instances {
		instances[k] = v
	}
	return instances, nil
}

func (r *ioFaultResolver) Errno(ctx context.Context, obj *v1alpha1.IoFault) (int, error) {
	return int(obj.Errno), nil
}

func (r *loggerResolver) Component(ctx context.Context, ns *string, component model.Component) (<-chan string, error) {
	if ns == nil {
		ns = new(string)
		*ns = DefaultNamespace
	}

	var list v1.PodList
	if err := r.Client.List(ctx, &list, client.MatchingLabels(componentLabels(component)), client.InNamespace(*ns)); err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("instance of %s not found", component)
	}

	return r.Pod(ctx, &list.Items[0].Namespace, list.Items[0].Name)
}

func (r *loggerResolver) Pod(ctx context.Context, ns *string, name string) (<-chan string, error) {
	if ns == nil {
		ns = new(string)
		*ns = DefaultNamespace
	}

	logs, err := r.Clientset.CoreV1().Pods(*ns).GetLogs(name, &v1.PodLogOptions{Follow: true}).Stream()
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
				r.Log.Error(err, fmt.Sprintf("fail to read log of pod(%s/%s)", *ns, name))
				break
			}
			select {
			case logChan <- string(line):
				continue
			case <-time.NewTimer(time.Minute).C:
				r.Log.Info(fmt.Sprintf("client has not read log of pod(%s/%s) for 1m, close channel", *ns, name))
				return
			}
		}
	}()
	return logChan, nil
}

func (r *mistakeSpecResolver) Filling(ctx context.Context, obj *v1alpha1.MistakeSpec) (*string, error) {
	filling := string(obj.Filling)
	return &filling, nil
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

func (r *networkChaosResolver) Podnetworks(ctx context.Context, obj *v1alpha1.NetworkChaos) ([]*v1alpha1.PodNetworkChaos, error) {
	podnetworks := make([]*v1alpha1.PodNetworkChaos, 0, len(obj.Status.Instances))
	for id := range obj.Status.Instances {
		podnetwork := new(v1alpha1.PodNetworkChaos)
		if err := r.Client.Get(ctx, parseNamespacedName(id), podnetwork); err != nil {
			return nil, err
		}
		podnetworks = append(podnetworks, podnetwork)
	}
	return podnetworks, nil
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

func (r *podResolver) Logs(ctx context.Context, obj *v1.Pod) (string, error) {
	logs, err := r.Clientset.CoreV1().Pods(obj.Namespace).GetLogs(obj.Name, &v1.PodLogOptions{}).Stream()
	if err != nil {
		return "", err
	}
	defer logs.Close()
	data, err := ioutil.ReadAll(logs)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *podResolver) Daemon(ctx context.Context, obj *v1.Pod) (*v1.Pod, error) {
	var list v1.PodList
	if err := r.Client.List(ctx, &list, client.MatchingLabels(componentLabels(model.ComponentDaemon))); err != nil {
		return nil, err
	}
	for _, daemon := range list.Items {
		if obj.Spec.NodeName == daemon.Spec.NodeName {
			return &daemon, nil
		}
	}
	return nil, fmt.Errorf("daemon of pod(%s/%s) not found", obj.Namespace, obj.Name)
}

func (r *podResolver) Processes(ctx context.Context, obj *v1.Pod) ([]*model.Process, error) {
	return r.GetPidFromPS(ctx, obj)
}

func (r *podResolver) Mounts(ctx context.Context, obj *v1.Pod) ([]string, error) {
	return r.GetMounts(ctx, obj)
}

func (r *podResolver) Ipset(ctx context.Context, obj *v1.Pod) (string, error) {
	return r.GetIpset(ctx, obj)
}

func (r *podResolver) TcQdisc(ctx context.Context, obj *v1.Pod) (string, error) {
	return r.GetTcQdisc(ctx, obj)
}

func (r *podResolver) Iptables(ctx context.Context, obj *v1.Pod) (string, error) {
	return r.GetIptables(ctx, obj)
}

func (r *podConditionResolver) Type(ctx context.Context, obj *v1.PodCondition) (string, error) {
	return string(obj.Type), nil
}

func (r *podConditionResolver) Status(ctx context.Context, obj *v1.PodCondition) (string, error) {
	return string(obj.Status), nil
}

func (r *podConditionResolver) LastProbeTime(ctx context.Context, obj *v1.PodCondition) (*time.Time, error) {
	return &obj.LastProbeTime.Time, nil
}

func (r *podConditionResolver) LastTransitionTime(ctx context.Context, obj *v1.PodCondition) (*time.Time, error) {
	return &obj.LastTransitionTime.Time, nil
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

func (r *podHTTPChaosResolver) Pod(ctx context.Context, obj *v1alpha1.PodHttpChaos) (*v1.Pod, error) {
	pod := new(v1.Pod)
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: obj.Namespace, Name: obj.Name}, pod); err != nil {
		return nil, err
	}
	return pod, nil
}

func (r *podHttpChaosReplaceActionsResolver) Body(ctx context.Context, obj *v1alpha1.PodHttpChaosReplaceActions) (*string, error) {
	data, err := json.Marshal(obj.Body)
	if err != nil {
		return nil, err
	}

	body := string(data)
	return &body, nil
}

func (r *podHttpChaosReplaceActionsResolver) Queries(ctx context.Context, obj *v1alpha1.PodHttpChaosReplaceActions) (map[string]interface{}, error) {
	queries := make(map[string]interface{})
	for k, v := range obj.Queries {
		queries[k] = v
	}
	return queries, nil
}

func (r *podHttpChaosReplaceActionsResolver) Headers(ctx context.Context, obj *v1alpha1.PodHttpChaosReplaceActions) (map[string]interface{}, error) {
	headers := make(map[string]interface{})
	for k, v := range obj.Headers {
		headers[k] = v
	}
	return headers, nil
}

func (r *podHttpChaosRuleResolver) Target(ctx context.Context, obj *v1alpha1.PodHttpChaosRule) (string, error) {
	return string(obj.Target), nil
}

func (r *podHttpChaosSelectorResolver) RequestHeaders(ctx context.Context, obj *v1alpha1.PodHttpChaosSelector) (map[string]interface{}, error) {
	headers := make(map[string]interface{})
	for k, v := range obj.RequestHeaders {
		headers[k] = v
	}
	return headers, nil
}

func (r *podHttpChaosSelectorResolver) ResponseHeaders(ctx context.Context, obj *v1alpha1.PodHttpChaosSelector) (map[string]interface{}, error) {
	headers := make(map[string]interface{})
	for k, v := range obj.ResponseHeaders {
		headers[k] = v
	}
	return headers, nil
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

func (r *podNetworkChaosResolver) Pod(ctx context.Context, obj *v1alpha1.PodNetworkChaos) (*v1.Pod, error) {
	pod := new(v1.Pod)
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: obj.Namespace, Name: obj.Name}, pod); err != nil {
		return nil, err
	}
	return pod, nil
}

func (r *podSelectorSpecResolver) Pods(ctx context.Context, obj *v1alpha1.PodSelectorSpec) (map[string]interface{}, error) {
	pods := make(map[string]interface{})
	for k, v := range obj.Pods {
		pods[k] = v
	}
	return pods, nil
}

func (r *podSelectorSpecResolver) NodeSelectors(ctx context.Context, obj *v1alpha1.PodSelectorSpec) (map[string]interface{}, error) {
	selectors := make(map[string]interface{})
	for k, v := range obj.NodeSelectors {
		selectors[k] = v
	}
	return selectors, nil
}

func (r *podSelectorSpecResolver) FieldSelectors(ctx context.Context, obj *v1alpha1.PodSelectorSpec) (map[string]interface{}, error) {
	selectors := make(map[string]interface{})
	for k, v := range obj.FieldSelectors {
		selectors[k] = v
	}
	return selectors, nil
}

func (r *podSelectorSpecResolver) LabelSelectors(ctx context.Context, obj *v1alpha1.PodSelectorSpec) (map[string]interface{}, error) {
	selectors := make(map[string]interface{})
	for k, v := range obj.LabelSelectors {
		selectors[k] = v
	}
	return selectors, nil
}

func (r *podSelectorSpecResolver) AnnotationSelectors(ctx context.Context, obj *v1alpha1.PodSelectorSpec) (map[string]interface{}, error) {
	selectors := make(map[string]interface{})
	for k, v := range obj.AnnotationSelectors {
		selectors[k] = v
	}
	return selectors, nil
}

func (r *podStatusResolver) Phase(ctx context.Context, obj *v1.PodStatus) (string, error) {
	return string(obj.Phase), nil
}

func (r *podStatusResolver) StartTime(ctx context.Context, obj *v1.PodStatus) (*time.Time, error) {
	return &obj.StartTime.Time, nil
}

func (r *podStatusResolver) QosClass(ctx context.Context, obj *v1.PodStatus) (string, error) {
	return string(obj.QOSClass), nil
}

func (r *processResolver) Fds(ctx context.Context, obj *model.Process) ([]*model.Fd, error) {
	return r.GetFdsOfProcess(ctx, obj)
}

func (r *queryResolver) Namepsace(ctx context.Context, ns *string) (*model.Namespace, error) {
	if ns == nil {
		ns = new(string)
		*ns = DefaultNamespace
	}
	return &model.Namespace{Ns: *ns}, nil
}

func (r *rawIptablesResolver) Direction(ctx context.Context, obj *v1alpha1.RawIptables) (string, error) {
	return string(obj.Direction), nil
}

func (r *rawTrafficControlResolver) Type(ctx context.Context, obj *v1alpha1.RawTrafficControl) (string, error) {
	return string(obj.Type), nil
}

func (r *recordResolver) Phase(ctx context.Context, obj *v1alpha1.Record) (string, error) {
	return string(obj.Phase), nil
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

// AttrOverrideSpec returns generated.AttrOverrideSpecResolver implementation.
func (r *Resolver) AttrOverrideSpec() generated.AttrOverrideSpecResolver {
	return &attrOverrideSpecResolver{r}
}

// BandwidthSpec returns generated.BandwidthSpecResolver implementation.
func (r *Resolver) BandwidthSpec() generated.BandwidthSpecResolver { return &bandwidthSpecResolver{r} }

// ChaosCondition returns generated.ChaosConditionResolver implementation.
func (r *Resolver) ChaosCondition() generated.ChaosConditionResolver {
	return &chaosConditionResolver{r}
}

// ContainerStateRunning returns generated.ContainerStateRunningResolver implementation.
func (r *Resolver) ContainerStateRunning() generated.ContainerStateRunningResolver {
	return &containerStateRunningResolver{r}
}

// ContainerStateTerminated returns generated.ContainerStateTerminatedResolver implementation.
func (r *Resolver) ContainerStateTerminated() generated.ContainerStateTerminatedResolver {
	return &containerStateTerminatedResolver{r}
}

// CorruptSpec returns generated.CorruptSpecResolver implementation.
func (r *Resolver) CorruptSpec() generated.CorruptSpecResolver { return &corruptSpecResolver{r} }

// ExperimentStatus returns generated.ExperimentStatusResolver implementation.
func (r *Resolver) ExperimentStatus() generated.ExperimentStatusResolver {
	return &experimentStatusResolver{r}
}

// HTTPChaos returns generated.HTTPChaosResolver implementation.
func (r *Resolver) HTTPChaos() generated.HTTPChaosResolver { return &hTTPChaosResolver{r} }

// HTTPChaosSpec returns generated.HTTPChaosSpecResolver implementation.
func (r *Resolver) HTTPChaosSpec() generated.HTTPChaosSpecResolver { return &hTTPChaosSpecResolver{r} }

// HTTPChaosStatus returns generated.HTTPChaosStatusResolver implementation.
func (r *Resolver) HTTPChaosStatus() generated.HTTPChaosStatusResolver {
	return &hTTPChaosStatusResolver{r}
}

// IOChaos returns generated.IOChaosResolver implementation.
func (r *Resolver) IOChaos() generated.IOChaosResolver { return &iOChaosResolver{r} }

// IOChaosAction returns generated.IOChaosActionResolver implementation.
func (r *Resolver) IOChaosAction() generated.IOChaosActionResolver { return &iOChaosActionResolver{r} }

// IOChaosSpec returns generated.IOChaosSpecResolver implementation.
func (r *Resolver) IOChaosSpec() generated.IOChaosSpecResolver { return &iOChaosSpecResolver{r} }

// IOChaosStatus returns generated.IOChaosStatusResolver implementation.
func (r *Resolver) IOChaosStatus() generated.IOChaosStatusResolver { return &iOChaosStatusResolver{r} }

// IoFault returns generated.IoFaultResolver implementation.
func (r *Resolver) IoFault() generated.IoFaultResolver { return &ioFaultResolver{r} }

// Logger returns generated.LoggerResolver implementation.
func (r *Resolver) Logger() generated.LoggerResolver { return &loggerResolver{r} }

// MistakeSpec returns generated.MistakeSpecResolver implementation.
func (r *Resolver) MistakeSpec() generated.MistakeSpecResolver { return &mistakeSpecResolver{r} }

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

// PodCondition returns generated.PodConditionResolver implementation.
func (r *Resolver) PodCondition() generated.PodConditionResolver { return &podConditionResolver{r} }

// PodHTTPChaos returns generated.PodHTTPChaosResolver implementation.
func (r *Resolver) PodHTTPChaos() generated.PodHTTPChaosResolver { return &podHTTPChaosResolver{r} }

// PodHttpChaosReplaceActions returns generated.PodHttpChaosReplaceActionsResolver implementation.
func (r *Resolver) PodHttpChaosReplaceActions() generated.PodHttpChaosReplaceActionsResolver {
	return &podHttpChaosReplaceActionsResolver{r}
}

// PodHttpChaosRule returns generated.PodHttpChaosRuleResolver implementation.
func (r *Resolver) PodHttpChaosRule() generated.PodHttpChaosRuleResolver {
	return &podHttpChaosRuleResolver{r}
}

// PodHttpChaosSelector returns generated.PodHttpChaosSelectorResolver implementation.
func (r *Resolver) PodHttpChaosSelector() generated.PodHttpChaosSelectorResolver {
	return &podHttpChaosSelectorResolver{r}
}

// PodIOChaos returns generated.PodIOChaosResolver implementation.
func (r *Resolver) PodIOChaos() generated.PodIOChaosResolver { return &podIOChaosResolver{r} }

// PodNetworkChaos returns generated.PodNetworkChaosResolver implementation.
func (r *Resolver) PodNetworkChaos() generated.PodNetworkChaosResolver {
	return &podNetworkChaosResolver{r}
}

// PodSelectorSpec returns generated.PodSelectorSpecResolver implementation.
func (r *Resolver) PodSelectorSpec() generated.PodSelectorSpecResolver {
	return &podSelectorSpecResolver{r}
}

// PodStatus returns generated.PodStatusResolver implementation.
func (r *Resolver) PodStatus() generated.PodStatusResolver { return &podStatusResolver{r} }

// Process returns generated.ProcessResolver implementation.
func (r *Resolver) Process() generated.ProcessResolver { return &processResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// RawIptables returns generated.RawIptablesResolver implementation.
func (r *Resolver) RawIptables() generated.RawIptablesResolver { return &rawIptablesResolver{r} }

// RawTrafficControl returns generated.RawTrafficControlResolver implementation.
func (r *Resolver) RawTrafficControl() generated.RawTrafficControlResolver {
	return &rawTrafficControlResolver{r}
}

// Record returns generated.RecordResolver implementation.
func (r *Resolver) Record() generated.RecordResolver { return &recordResolver{r} }

// StressChaos returns generated.StressChaosResolver implementation.
func (r *Resolver) StressChaos() generated.StressChaosResolver { return &stressChaosResolver{r} }

type attrOverrideSpecResolver struct{ *Resolver }
type bandwidthSpecResolver struct{ *Resolver }
type chaosConditionResolver struct{ *Resolver }
type containerStateRunningResolver struct{ *Resolver }
type containerStateTerminatedResolver struct{ *Resolver }
type corruptSpecResolver struct{ *Resolver }
type experimentStatusResolver struct{ *Resolver }
type hTTPChaosResolver struct{ *Resolver }
type hTTPChaosSpecResolver struct{ *Resolver }
type hTTPChaosStatusResolver struct{ *Resolver }
type iOChaosResolver struct{ *Resolver }
type iOChaosActionResolver struct{ *Resolver }
type iOChaosSpecResolver struct{ *Resolver }
type iOChaosStatusResolver struct{ *Resolver }
type ioFaultResolver struct{ *Resolver }
type loggerResolver struct{ *Resolver }
type mistakeSpecResolver struct{ *Resolver }
type namespaceResolver struct{ *Resolver }
type networkChaosResolver struct{ *Resolver }
type ownerReferenceResolver struct{ *Resolver }
type podResolver struct{ *Resolver }
type podConditionResolver struct{ *Resolver }
type podHTTPChaosResolver struct{ *Resolver }
type podHttpChaosReplaceActionsResolver struct{ *Resolver }
type podHttpChaosRuleResolver struct{ *Resolver }
type podHttpChaosSelectorResolver struct{ *Resolver }
type podIOChaosResolver struct{ *Resolver }
type podNetworkChaosResolver struct{ *Resolver }
type podSelectorSpecResolver struct{ *Resolver }
type podStatusResolver struct{ *Resolver }
type processResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type rawIptablesResolver struct{ *Resolver }
type rawTrafficControlResolver struct{ *Resolver }
type recordResolver struct{ *Resolver }
type stressChaosResolver struct{ *Resolver }
