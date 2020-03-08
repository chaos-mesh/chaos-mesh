// Copyright 2019 PingCAP, Inc.
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
	"strconv"
	"strings"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/pkg/label"
	"github.com/pingcap/chaos-mesh/pkg/mock"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
)

type SelectSpec interface {
	GetSelector() v1alpha1.SelectorSpec
	GetMode() v1alpha1.PodMode
	GetValue() string
}

func SelectAndGeneratePods(ctx context.Context, c client.Client, spec SelectSpec) ([]v1.Pod, error) {
	if pods := mock.On("MockSelectAndGeneratePods"); pods != nil {
		return pods.(func() []v1.Pod)(), nil
	}

	selector := spec.GetSelector()
	mode := spec.GetMode()
	value := spec.GetValue()

	pods, err := SelectPods(ctx, c, selector)
	if err != nil {
		return nil, err
	}

	if len(pods) == 0 {
		err = errors.New("no pod is selected")
		return nil, err
	}

	filteredPod, err := GeneratePods(pods, mode, value)
	if err != nil {
		return nil, err
	}

	return filteredPod, nil
}

// SelectPods returns the list of pods that are available for pod chaos action.
// It returns all pods that match the configured label, annotation and namespace selectors.
// If pods are specifically specified by `selector.Pods`, it just returns the selector.Pods.
func SelectPods(ctx context.Context, c client.Client, selector v1alpha1.SelectorSpec) ([]v1.Pod, error) {
	var pods []v1.Pod

	// pods are specifically specified
	if len(selector.Pods) > 0 {
		for ns, names := range selector.Pods {
			for _, name := range names {
				var pod v1.Pod
				err := c.Get(ctx, types.NamespacedName{
					Namespace: ns,
					Name:      name,
				}, &pod)
				if err != nil {
					return nil, err
				}

				pods = append(pods, pod)
			}
		}

		return pods, nil
	}

	var podList v1.PodList

	var listOptions = client.ListOptions{}
	if len(selector.LabelSelectors) > 0 {
		listOptions.LabelSelector = labels.SelectorFromSet(selector.LabelSelectors)
	}
	if len(selector.FieldSelectors) > 0 {
		listOptions.FieldSelector = fields.SelectorFromSet(selector.FieldSelectors)
	}
	err := c.List(ctx, &podList, &listOptions)

	if err != nil {
		return nil, err
	}

	pods = append(pods, podList.Items...)

	namespaceSelector, err := parseSelector(strings.Join(selector.Namespaces, ","))
	if err != nil {
		return nil, err
	}
	pods, err = filterByNamespaces(pods, namespaceSelector)
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

	pods, err = filterByNamespaces(pods, namespaceSelector)
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

// GeneratePods generate pods according to mode from pod list
func GeneratePods(pods []v1.Pod, mode v1alpha1.PodMode, value string) ([]v1.Pod, error) {
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

// filterByNamespaces filters a list of pods by a given namespace selector.
func filterByNamespaces(pods []v1.Pod, namespaces labels.Selector) ([]v1.Pod, error) {
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
