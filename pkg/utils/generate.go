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
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

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
