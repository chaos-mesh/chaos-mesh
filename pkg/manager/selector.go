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

package manager

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/label"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// SelectPods returns the list of pods that are available for pod chaos action.
// It returns all pods that match the configured label, annotation and namespace selectors.
// If pods are specifically specified by `selector.Pods`, it just returns the selector.Pods.
func SelectPods(
	selector v1alpha1.SelectorSpec,
	podLister corelisters.PodLister,
	kubeCli kubernetes.Interface,
) ([]v1.Pod, error) {
	var pods []v1.Pod

	// pods are specifically specified
	if len(selector.Pods) > 0 {
		for ns, names := range selector.Pods {
			for _, name := range names {
				pod, err := podLister.Pods(ns).Get(name)
				if err != nil {
					return nil, err
				}

				pods = append(pods, *pod)
			}
		}

		return pods, nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: label.Label(selector.LabelSelectors).String(),
		FieldSelector: label.Label(selector.FieldSelectors).String(),
	}

	podList, err := kubeCli.CoreV1().Pods(metav1.NamespaceAll).List(listOptions)
	if err != nil {
		return nil, err
	}

	for _, pod := range podList.Items {
		pods = append(pods, pod)
	}

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

	pods = filterByPhase(pods, v1.PodRunning)

	return pods, nil
}

// RandomFixedIndexes returns the `count` random indexes between `start` and `end`.
func RandomFixedIndexes(start, end, count int) []int {
	var indexes []int
	m := make(map[int]int, count)

	if count > end-start {
		for i := start; i < end; i++ {
			indexes = append(indexes, i)
		}

		return indexes
	}

	for i := 0; i < count; {
		index := rand.Intn(end-start) + start

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

// filterByPhase filters a list of pods by a given PodPhase, e.g. Running.
func filterByPhase(pods []v1.Pod, phase v1.PodPhase) []v1.Pod {
	var filteredList []v1.Pod

	for _, pod := range pods {
		if pod.Status.Phase == phase {
			filteredList = append(filteredList, pod)
		}
	}

	return filteredList
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
