// Copyright 2019 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package pod

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/label"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var log = ctrl.Log.WithName("selector")

const injectAnnotationKey = "chaos-mesh.org/inject"

type Option struct {
	ClusterScoped         bool
	TargetNamespace       string
	EnableFilterNamespace bool
}

type SelectImpl struct {
	c client.Client
	r client.Reader

	Option
}

type Pod struct {
	v1.Pod
}

func (pod *Pod) Id() string {
	return (types.NamespacedName{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}).String()
}

func (impl *SelectImpl) Select(ctx context.Context, ps *v1alpha1.PodSelector) ([]*Pod, error) {
	if ps == nil {
		return []*Pod{}, nil
	}

	pods, err := SelectAndFilterPods(ctx, impl.c, impl.r, ps, impl.ClusterScoped, impl.TargetNamespace, impl.EnableFilterNamespace)
	if err != nil {
		return nil, err
	}

	var result []*Pod
	for _, pod := range pods {
		result = append(result, &Pod{
			pod,
		})
	}

	return result, nil
}

type Params struct {
	fx.In

	Client client.Client
	Reader client.Reader `name:"no-cache"`
}

func New(params Params) *SelectImpl {
	return &SelectImpl{
		params.Client,
		params.Reader,
		Option{
			config.ControllerCfg.ClusterScoped,
			config.ControllerCfg.TargetNamespace,
			config.ControllerCfg.EnableFilterNamespace,
		},
	}
}

// SelectAndFilterPods returns the list of pods that filtered by selector and PodMode
func SelectAndFilterPods(ctx context.Context, c client.Client, r client.Reader, spec *v1alpha1.PodSelector, clusterScoped bool, targetNamespace string, enableFilterNamespace bool) ([]v1.Pod, error) {
	if pods := mock.On("MockSelectAndFilterPods"); pods != nil {
		return pods.(func() []v1.Pod)(), nil
	}
	if err := mock.On("MockSelectedAndFilterPodsError"); err != nil {
		return nil, err.(error)
	}

	selector := spec.Selector
	mode := spec.Mode
	value := spec.Value

	pods, err := SelectPods(ctx, c, r, selector, clusterScoped, targetNamespace, enableFilterNamespace)
	if err != nil {
		return nil, err
	}

	if len(pods) == 0 {
		err = errors.New("no pod is selected")
		return nil, err
	}

	filteredPod, err := filterPodsByMode(pods, mode, value)
	if err != nil {
		return nil, err
	}

	return filteredPod, nil
}

//revive:disable:flag-parameter

// SelectPods returns the list of pods that are available for pod chaos action.
// It returns all pods that match the configured label, annotation and namespace selectors.
// If pods are specifically specified by `selector.Pods`, it just returns the selector.Pods.
func SelectPods(ctx context.Context, c client.Client, r client.Reader, selector v1alpha1.PodSelectorSpec, clusterScoped bool, targetNamespace string, enableFilterNamespace bool) ([]v1.Pod, error) {
	// TODO: refactor: make different selectors to replace if-else logics
	var pods []v1.Pod

	// pods are specifically specified
	if len(selector.Pods) > 0 {
		for ns, names := range selector.Pods {
			if !clusterScoped {
				if targetNamespace != ns {
					log.Info("skip namespace because ns is out of scope within namespace scoped mode", "namespace", ns)
					continue
				}
			}
			for _, name := range names {
				var pod v1.Pod
				err := c.Get(ctx, types.NamespacedName{
					Namespace: ns,
					Name:      name,
				}, &pod)
				if err == nil {
					pods = append(pods, pod)
					continue
				}

				if apierrors.IsNotFound(err) {
					log.Error(err, "Pod is not found", "namespace", ns, "pod name", name)
					continue
				}

				return nil, err
			}
		}

		return pods, nil
	}

	if !clusterScoped {
		if len(selector.Namespaces) > 1 {
			return nil, fmt.Errorf("could NOT use more than 1 namespace selector within namespace scoped mode")
		} else if len(selector.Namespaces) == 1 {
			if selector.Namespaces[0] != targetNamespace {
				return nil, fmt.Errorf("could NOT list pods from out of scoped namespace: %s", selector.Namespaces[0])
			}
		}
	}

	var listOptions = client.ListOptions{}
	if !clusterScoped {
		listOptions.Namespace = targetNamespace
	}
	if len(selector.LabelSelectors) > 0 || len(selector.ExpressionSelectors) > 0 {
		metav1Ls := &metav1.LabelSelector{
			MatchLabels:      selector.LabelSelectors,
			MatchExpressions: selector.ExpressionSelectors,
		}
		ls, err := metav1.LabelSelectorAsSelector(metav1Ls)
		if err != nil {
			return nil, err
		}
		listOptions.LabelSelector = ls
	}

	listFunc := c.List

	if len(selector.FieldSelectors) > 0 {
		listOptions.FieldSelector = fields.SelectorFromSet(selector.FieldSelectors)

		// Since FieldSelectors need to implement index creation, Reader.List is used to get the pod list.
		// Otherwise, just call Client.List directly, which can be obtained through cache.
		if r != nil {
			listFunc = r.List
		}
	}

	var podList v1.PodList
	if len(selector.Namespaces) > 0 {
		for _, namespace := range selector.Namespaces {
			listOptions.Namespace = namespace

			if err := listFunc(ctx, &podList, &listOptions); err != nil {
				return nil, err
			}

			pods = append(pods, podList.Items...)
		}
	} else {
		if err := listFunc(ctx, &podList, &listOptions); err != nil {
			return nil, err
		}

		pods = append(pods, podList.Items...)
	}

	var (
		nodes           []v1.Node
		nodeList        v1.NodeList
		nodeListOptions = client.ListOptions{}
	)
	// if both setting Nodes and NodeSelectors, the node list will be combined.
	if len(selector.Nodes) > 0 || len(selector.NodeSelectors) > 0 {
		if len(selector.Nodes) > 0 {
			for _, nodename := range selector.Nodes {
				var node v1.Node
				if err := c.Get(ctx, types.NamespacedName{Name: nodename}, &node); err != nil {
					return nil, err
				}
				nodes = append(nodes, node)
			}
		}
		if len(selector.NodeSelectors) > 0 {
			nodeListOptions.LabelSelector = labels.SelectorFromSet(selector.NodeSelectors)
			if err := c.List(ctx, &nodeList, &nodeListOptions); err != nil {
				return nil, err
			}
			nodes = append(nodes, nodeList.Items...)
		}
		pods = filterPodByNode(pods, nodes)
	}
	if enableFilterNamespace {
		pods = filterByNamespaces(ctx, c, pods)
	}

	namespaceSelector, err := parseSelector(strings.Join(selector.Namespaces, ","))
	if err != nil {
		return nil, err
	}
	pods, err = filterByNamespaceSelector(pods, namespaceSelector)
	if err != nil {
		return nil, err
	}

	annotationsSelector, err := parseSelector(label.Label(selector.AnnotationSelectors).String())
	if err != nil {
		return nil, err
	}
	pods = filterByAnnotations(pods, annotationsSelector)

	phaseSelector, err := parseSelector(strings.Join(selector.PodPhaseSelectors, ","))
	if err != nil {
		return nil, err
	}
	pods, err = filterByPhaseSelector(pods, phaseSelector)
	if err != nil {
		return nil, err
	}

	return pods, nil
}

//revive:enable:flag-parameter

// GetService get k8s service by service name
func GetService(ctx context.Context, c client.Client, namespace, controllerNamespace string, serviceName string) (*v1.Service, error) {
	// use the environment value if namespace is empty
	if len(namespace) == 0 {
		namespace = controllerNamespace
	}

	service := &v1.Service{}
	err := c.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      serviceName,
	}, service)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// CheckPodMeetSelector checks if this pod meets the selection criteria.
// TODO: support to check fieldsSelector
func CheckPodMeetSelector(pod v1.Pod, selector v1alpha1.PodSelectorSpec) (bool, error) {
	if len(selector.Pods) > 0 {
		meet := false
		for ns, names := range selector.Pods {
			if pod.Namespace != ns {
				continue
			}

			for _, name := range names {
				if pod.Name == name {
					meet = true
				}
			}

			if !meet {
				return false, nil
			}
		}
	}

	// check pod labels.
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}

	if selector.LabelSelectors == nil {
		selector.LabelSelectors = make(map[string]string)
	}

	if len(selector.LabelSelectors) > 0 || len(selector.ExpressionSelectors) > 0 {
		metav1Ls := &metav1.LabelSelector{
			MatchLabels:      selector.LabelSelectors,
			MatchExpressions: selector.ExpressionSelectors,
		}
		ls, err := metav1.LabelSelectorAsSelector(metav1Ls)
		if err != nil {
			return false, err
		}
		podLabels := labels.Set(pod.Labels)
		if len(pod.Labels) == 0 || !ls.Matches(podLabels) {
			return false, nil
		}
	}

	pods := []v1.Pod{pod}

	namespaceSelector, err := parseSelector(strings.Join(selector.Namespaces, ","))
	if err != nil {
		return false, err
	}

	pods, err = filterByNamespaceSelector(pods, namespaceSelector)
	if err != nil {
		return false, err
	}

	annotationsSelector, err := parseSelector(label.Label(selector.AnnotationSelectors).String())
	if err != nil {
		return false, err
	}

	pods = filterByAnnotations(pods, annotationsSelector)

	phaseSelector, err := parseSelector(strings.Join(selector.PodPhaseSelectors, ","))
	if err != nil {
		return false, err
	}
	pods, err = filterByPhaseSelector(pods, phaseSelector)
	if err != nil {
		return false, err
	}

	if len(pods) > 0 {
		return true, nil
	}

	return false, nil
}

func filterPodByNode(pods []v1.Pod, nodes []v1.Node) []v1.Pod {
	if len(nodes) == 0 {
		return nil
	}
	var filteredList []v1.Pod
	for _, pod := range pods {
		for _, node := range nodes {
			if pod.Spec.NodeName == node.Name {
				filteredList = append(filteredList, pod)
			}
		}
	}
	return filteredList
}

// filterPodsByMode filters pods by mode from pod list
func filterPodsByMode(pods []v1.Pod, mode v1alpha1.PodMode, value string) ([]v1.Pod, error) {
	if len(pods) == 0 {
		return nil, errors.New("cannot generate pods from empty list")
	}

	switch mode {
	case v1alpha1.OnePodMode:
		index := getRandomNumber(len(pods))
		pod := pods[index]

		return []v1.Pod{pod}, nil
	case v1alpha1.AllPodMode:
		return pods, nil
	case v1alpha1.FixedPodMode:
		num, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		if len(pods) < num {
			num = len(pods)
		}

		if num <= 0 {
			return nil, errors.New("cannot select any pod as value below or equal 0")
		}

		return getFixedSubListFromPodList(pods, num), nil
	case v1alpha1.FixedPercentPodMode:
		percentage, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		if percentage == 0 {
			return nil, errors.New("cannot select any pod as value below or equal 0")
		}

		if percentage < 0 || percentage > 100 {
			return nil, fmt.Errorf("fixed percentage value of %d is invalid, Must be (0,100]", percentage)
		}

		num := int(math.Floor(float64(len(pods)) * float64(percentage) / 100))

		return getFixedSubListFromPodList(pods, num), nil
	case v1alpha1.RandomMaxPercentPodMode:
		maxPercentage, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		if maxPercentage == 0 {
			return nil, errors.New("cannot select any pod as value below or equal 0")
		}

		if maxPercentage < 0 || maxPercentage > 100 {
			return nil, fmt.Errorf("fixed percentage value of %d is invalid, Must be [0-100]", maxPercentage)
		}

		percentage := getRandomNumber(maxPercentage + 1) // + 1 because Intn works with half open interval [0,n) and we want [0,n]
		num := int(math.Floor(float64(len(pods)) * float64(percentage) / 100))

		return getFixedSubListFromPodList(pods, num), nil
	default:
		return nil, fmt.Errorf("mode %s not supported", mode)
	}
}

// filterByAnnotations filters a list of pods by a given annotation selector.
func filterByAnnotations(pods []v1.Pod, annotations labels.Selector) []v1.Pod {
	// empty filter returns original list
	if annotations.Empty() {
		return pods
	}

	var filteredList []v1.Pod

	for _, pod := range pods {
		// convert the pod's annotations to an equivalent label selector
		selector := labels.Set(pod.Annotations)

		// include pod if its annotations match the selector
		if annotations.Matches(selector) {
			filteredList = append(filteredList, pod)
		}
	}

	return filteredList
}

// filterByPhaseSet filters a list of pods by a given PodPhase selector.
func filterByPhaseSelector(pods []v1.Pod, phases labels.Selector) ([]v1.Pod, error) {
	if phases.Empty() {
		return pods, nil
	}

	reqs, _ := phases.Requirements()
	var (
		reqIncl []labels.Requirement
		reqExcl []labels.Requirement

		filteredList []v1.Pod
	)

	for _, req := range reqs {
		switch req.Operator() {
		case selection.Exists:
			reqIncl = append(reqIncl, req)
		case selection.DoesNotExist:
			reqExcl = append(reqExcl, req)
		default:
			return nil, fmt.Errorf("unsupported operator: %s", req.Operator())
		}
	}

	for _, pod := range pods {
		included := len(reqIncl) == 0
		selector := labels.Set{string(pod.Status.Phase): ""}

		// include pod if one including requirement matches
		for _, req := range reqIncl {
			if req.Matches(selector) {
				included = true
				break
			}
		}

		// exclude pod if it is filtered out by at least one excluding requirement
		for _, req := range reqExcl {
			if !req.Matches(selector) {
				included = false
				break
			}
		}

		if included {
			filteredList = append(filteredList, pod)
		}
	}

	return filteredList, nil
}

func filterByNamespaces(ctx context.Context, c client.Client, pods []v1.Pod) []v1.Pod {
	var filteredList []v1.Pod

	for _, pod := range pods {
		ok, err := IsAllowedNamespaces(ctx, c, pod.Namespace)
		if err != nil {
			log.Error(err, "fail to check whether this namespace is allowed", "namespace", pod.Namespace)
			continue
		}

		if ok {
			filteredList = append(filteredList, pod)
		} else {
			log.Info("namespace is not enabled for chaos-mesh", "namespace", pod.Namespace)
		}
	}
	return filteredList
}

func IsAllowedNamespaces(ctx context.Context, c client.Client, namespace string) (bool, error) {
	ns := &v1.Namespace{}

	err := c.Get(ctx, types.NamespacedName{Name: namespace}, ns)
	if err != nil {
		return false, err
	}

	if ns.Annotations[injectAnnotationKey] == "enabled" {
		return true, nil
	}

	return false, nil
}

// filterByNamespaceSelector filters a list of pods by a given namespace selector.
func filterByNamespaceSelector(pods []v1.Pod, namespaces labels.Selector) ([]v1.Pod, error) {
	// empty filter returns original list
	if namespaces.Empty() {
		return pods, nil
	}

	// split requirements into including and excluding groups
	reqs, _ := namespaces.Requirements()

	var (
		reqIncl []labels.Requirement
		reqExcl []labels.Requirement

		filteredList []v1.Pod
	)

	for _, req := range reqs {
		switch req.Operator() {
		case selection.Exists:
			reqIncl = append(reqIncl, req)
		case selection.DoesNotExist:
			reqExcl = append(reqExcl, req)
		default:
			return nil, fmt.Errorf("unsupported operator: %s", req.Operator())
		}
	}

	for _, pod := range pods {
		// if there aren't any including requirements, we're in by default
		included := len(reqIncl) == 0

		// convert the pod's namespace to an equivalent label selector
		selector := labels.Set{pod.Namespace: ""}

		// include pod if one including requirement matches
		for _, req := range reqIncl {
			if req.Matches(selector) {
				included = true
				break
			}
		}

		// exclude pod if it is filtered out by at least one excluding requirement
		for _, req := range reqExcl {
			if !req.Matches(selector) {
				included = false
				break
			}
		}

		if included {
			filteredList = append(filteredList, pod)
		}
	}

	return filteredList, nil
}

func parseSelector(str string) (labels.Selector, error) {
	selector, err := labels.Parse(str)
	if err != nil {
		return nil, err
	}
	return selector, nil
}

func getFixedSubListFromPodList(pods []v1.Pod, num int) []v1.Pod {
	indexes := RandomFixedIndexes(0, uint(len(pods)), uint(num))

	var filteredPods []v1.Pod

	for _, index := range indexes {
		index := index
		filteredPods = append(filteredPods, pods[index])
	}

	return filteredPods
}

// RandomFixedIndexes returns the `count` random indexes between `start` and `end`.
// [start, end)
func RandomFixedIndexes(start, end, count uint) []uint {
	var indexes []uint
	m := make(map[uint]uint, count)

	if end < start {
		return indexes
	}

	if count > end-start {
		for i := start; i < end; i++ {
			indexes = append(indexes, i)
		}

		return indexes
	}

	for i := 0; i < int(count); {
		index := uint(getRandomNumber(int(end-start))) + start

		_, exist := m[index]
		if exist {
			continue
		}

		m[index] = index
		indexes = append(indexes, index)
		i++
	}

	return indexes
}

func getRandomNumber(max int) uint64 {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return num.Uint64()
}
