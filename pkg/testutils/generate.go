// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package testutils

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// PodArg by default use `Status=corev1.PodRunning` and `Namespace=metav1.NamespaceDefault`.
// For the others, the default values are empty.
type PodArg struct {
	Name            string
	Status          v1.PodPhase
	Namespace       string
	Ans             map[string]string
	Labels          map[string]string
	ContainerStatus v1.ContainerStatus
	Nodename        string
}

func NewPod(p PodArg) v1.Pod {
	if p.Status == "" {
		p.Status = v1.PodRunning
	}
	if p.Namespace == "" {
		p.Namespace = metav1.NamespaceDefault
	}
	return v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        p.Name,
			Namespace:   p.Namespace,
			Labels:      p.Labels,
			Annotations: p.Ans,
		},
		Spec: v1.PodSpec{
			NodeName: p.Nodename,
		},
		Status: v1.PodStatus{
			Phase:             p.Status,
			ContainerStatuses: []v1.ContainerStatus{p.ContainerStatus},
		},
	}
}

func GenerateNPods(
	namePrefix string,
	n int,
	podArg PodArg,
) ([]runtime.Object, []v1.Pod) {
	var podObjects []runtime.Object
	var pods []v1.Pod
	for i := 0; i < n; i++ {
		podArg.Name = fmt.Sprintf("%s%d", namePrefix, i)
		pod := NewPod(podArg)
		podObjects = append(podObjects, &pod)
		pods = append(pods, pod)
	}

	return podObjects, pods
}

func NewNode(
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
		node := NewNode(fmt.Sprintf("%s%d", namePrefix, i), label)
		nodeObjects = append(nodeObjects, &node)
		nodes = append(nodes, node)
	}
	return nodeObjects, nodes
}

// PhysicalMachineArg by default use `Namespace=metav1.NamespaceDefault`.
// For the others, the default values are empty.
type PhysicalMachineArg struct {
	Name      string
	Namespace string
	Ans       map[string]string
	Labels    map[string]string
	Address   string
}

func NewPhysicalMachine(p PhysicalMachineArg) v1alpha1.PhysicalMachine {
	if p.Namespace == "" {
		p.Namespace = metav1.NamespaceDefault
	}
	return v1alpha1.PhysicalMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PhysicalMachine",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        p.Name,
			Namespace:   p.Namespace,
			Labels:      p.Labels,
			Annotations: p.Ans,
		},
		Spec: v1alpha1.PhysicalMachineSpec{
			Address: p.Address,
		},
	}
}

func GenerateNPhysicalMachines(
	namePrefix string,
	n int,
	arg PhysicalMachineArg,
) ([]runtime.Object, []v1alpha1.PhysicalMachine) {
	var physicalMachineObjects []runtime.Object
	var physicalMachines []v1alpha1.PhysicalMachine
	for i := 0; i < n; i++ {
		arg.Name = fmt.Sprintf("%s%d", namePrefix, i)
		physicalMachine := NewPhysicalMachine(arg)
		physicalMachineObjects = append(physicalMachineObjects, &physicalMachine)
		physicalMachines = append(physicalMachines, physicalMachine)
	}

	return physicalMachineObjects, physicalMachines
}
