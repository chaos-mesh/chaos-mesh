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

package utils

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/label"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
)

type SelectSpec interface {
	GetSelector() v1alpha1.SelectorSpec
	GetMode() v1alpha1.PodMode
	GetValue() string
}

// SelectAndFilterPods returns the list of pods that filtered by selector and PodMode
func SelectAndFilterPods(ctx context.Context, c client.Client, r client.Reader, spec SelectSpec, clusterScoped bool, targetNamespace string, allowedNamespaces, ignoredNamespaces string) ([]v1.Pod, error) {
	if pods := mock.On("MockSelectAndFilterPods"); pods != nil {
		return pods.(func() []v1.Pod)(), nil
	}
	if err := mock.On("MockSelectedAndFilterPodsError"); err != nil {
		return nil, err.(error)
	}

	selector := spec.GetSelector()
	mode := spec.GetMode()
	value := spec.GetValue()

	pods, err := SelectPods(ctx, c, r, selector, clusterScoped, targetNamespace, allowedNamespaces, ignoredNamespaces)
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

// SelectPods returns the list of pods that are available for pod chaos action.
// It returns all pods that match the configured label, annotation and namespace selectors.
// If pods are specifically specified by `selector.Pods`, it just returns the selector.Pods.
func SelectPods(ctx context.Context, c client.Client, r client.Reader, selector v1alpha1.SelectorSpec, clusterScoped bool, targetNamespace string, allowedNamespaces, ignoredNamespaces string) ([]v1.Pod, error) {
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
			if !IsAllowedNamespaces(ns, allowedNamespaces, ignoredNamespaces) {
				log.Info("filter pod by namespaces", "namespace", ns)
				continue
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

	var podList v1.PodList

	var listOptions = client.ListOptions{}
	if !clusterScoped {
		listOptions.Namespace = targetNamespace
	}
	if len(selector.LabelSelectors) > 0 {
		listOptions.LabelSelector = labels.SelectorFromSet(selector.LabelSelectors)
	}
	if len(selector.FieldSelectors) > 0 {
		// Since FieldSelectors need to implement index creation, Reader.List is used to get the pod list.
		listOptions.FieldSelector = fields.SelectorFromSet(selector.FieldSelectors)
		if err := r.List(ctx, &podList, &listOptions); err != nil {
			return nil, err
		}
	} else {
		// Otherwise, just call Client.List directly, which can be obtained through cache.
		if err := c.List(ctx, &podList, &listOptions); err != nil {
			return nil, err
		}
	}
	pods = append(pods, podList.Items...)
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
	pods = filterByNamespaces(pods, allowedNamespaces, ignoredNamespaces)

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
func CheckPodMeetSelector(pod v1.Pod, selector v1alpha1.SelectorSpec) (bool, error) {
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

	if len(selector.LabelSelectors) > 0 {
		ls := labels.SelectorFromSet(selector.LabelSelectors)
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
		index := rand.Intn(len(pods))
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

		percentage := rand.Intn(maxPercentage + 1) // + 1 because Intn works with half open interval [0,n) and we want [0,n]
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

func filterByNamespaces(pods []v1.Pod, allowedNamespaces, ignoredNamespaces string) []v1.Pod {
	var filteredList []v1.Pod

	for _, pod := range pods {
		if IsAllowedNamespaces(pod.Namespace, allowedNamespaces, ignoredNamespaces) {
			filteredList = append(filteredList, pod)
		} else {
			log.Info("filter pod by namespaces",
				"pod", pod.Name, "namespace", pod.Namespace)
		}
	}
	return filteredList
}

// IsAllowedNamespaces returns whether namespace allows the execution of a chaos task
func IsAllowedNamespaces(namespace string, allowedNamespaces, ignoredNamespaces string) bool {
	if allowedNamespaces != "" {
		matched, err := regexp.MatchString(allowedNamespaces, namespace)
		if err != nil {
			return false
		}
		return matched
	}

	if ignoredNamespaces != "" {
		matched, err := regexp.MatchString(ignoredNamespaces, namespace)
		if err != nil {
			return false
		}
		return !matched
	}

	return true
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
		index := uint(rand.Intn(int(end-start))) + start

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

func newPod(
	name string,
	status v1.PodPhase,
	namespace string,
	ans map[string]string,
	ls map[string]string,
	nodename string,
) v1.Pod {
	return v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      ls,
			Annotations: ans,
		},
		Spec: v1.PodSpec{
			NodeName: nodename,
		},
		Status: v1.PodStatus{
			Phase: status,
		},
	}
}

func GenerateNPods(
	namePrefix string,
	n int,
	status v1.PodPhase,
	ns string,
	ans map[string]string,
	ls map[string]string,
	nodename string,
) ([]runtime.Object, []v1.Pod) {
	var podObjects []runtime.Object
	var pods []v1.Pod
	for i := 0; i < n; i++ {
		pod := newPod(fmt.Sprintf("%s%d", namePrefix, i), status, ns, ans, ls, nodename)
		podObjects = append(podObjects, &pod)
		pods = append(pods, pod)
	}

	return podObjects, pods
}

func newNode(
	name string,
	label map[string]string,
) v1.Node {
	return v1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: label,
		},
	}
}

func GenerateNNodes(
	namePrefix string,
	n int,
	label map[string]string,
) ([]runtime.Object, []v1.Node) {
	var nodeObjects []runtime.Object
	var nodes []v1.Node

	for i := 0; i < n; i++ {
		node := newNode(fmt.Sprintf("%s%d", namePrefix, i), label)
		nodeObjects = append(nodeObjects, &node)
		nodes = append(nodes, node)
	}
	return nodeObjects, nodes
}
