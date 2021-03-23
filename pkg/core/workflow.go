// Copyright 2021 Chaos Mesh Authors.
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

package core

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type WorkflowRepository interface {
	ListWorkflowWithNamespace(ctx context.Context, namespace string) ([]Workflow, error)
	ListWorkflowFromAllNamespace(ctx context.Context) ([]Workflow, error)
	GetWorkflowByNamespacedName(ctx context.Context, namespace, name string) (WorkflowDetail, error)
	DeleteWorkflowByNamespacedName(ctx context.Context, namespace, name string) error
}

// Workflow defines the root structure of a workflow.
type Workflow struct {
	Name   string         `json:"name"`
	Entry  string         `json:"entry"` // the entry node name
}

type WorkflowDetail struct {
	Workflow     `json:",inline"`
	Topology     Topology `json:"topology"`
	CurrentNodes []Node   `json:"current_nodes"`
}

// Topology describes the process of a workflow.
type Topology struct {
	Nodes []Node `json:"nodes"`
}

// Node defines the single step of a workflow.
type Node struct {
	Name     string       `json:"name"`
	Type     NodeType     `json:"type"`
	Serial   NodeSerial   `json:"serial,omitempty"`
	Parallel NodeParallel `json:"parallel,omitempty"`
	Template string       `json:"template"`
}

// NodeSerial defines SerialNode's specific fields.
type NodeSerial struct {
	Tasks []string `json:"tasks"`
}

// NodeParallel defines ParallelNode's specific fields.
type NodeParallel struct {
	Tasks []string `json:"tasks"`
}

// NodeType defines the type of a workflow node.
//
// There will be five types can be refered as NodeType: Chaos, Serial, Parallel, Suspend, Task.
//
// Const definitions can be found below this type.
type NodeType string

const (
	// ChaosNode represents a node will perform a single Chaos Experiment.
	ChaosNode NodeType = "ChaosNode"

	// SerialNode represents a node that will perform continuous templates.
	SerialNode NodeType = "SerialNode"

	// ParallelNode represents a node that will perform parallel templates.
	ParallelNode NodeType = "ParallelNode"

	// SuspendNode represents a node that will perform wait operation.
	SuspendNode NodeType = "SuspendNode"

	// TaskNode represents a node that will perform user-defined task.
	TaskNode NodeType = "TaskNode"
)

// Detail defines the detail of a workflow.
type Detail struct {
	WorkflowUID string     `json:"workflow_uid"`
	Templates   []Template `json:"templates"`
}

// Template defines a complete structure of a template.
type Template struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Duration string      `json:"duration,omitempty"`
	Spec     interface{} `json:"spec"`
}

type KubeWorkflowRepository struct {
	kubeclient client.Client
}

func NewKubeWorkflowRepository(kubeclient client.Client) *KubeWorkflowRepository {
	return &KubeWorkflowRepository{kubeclient: kubeclient}
}

func (it *KubeWorkflowRepository) ListWorkflowWithNamespace(ctx context.Context, namespace string) ([]Workflow, error) {
	workflowList := v1alpha1.WorkflowList{}
	err := it.kubeclient.List(ctx, &workflowList, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return nil, err
	}

	var result []Workflow
	for _, item := range workflowList.Items {
		result = append(result, conversionWorkflow(item))
	}

	return result, nil
}

func (it *KubeWorkflowRepository) ListWorkflowFromAllNamespace(ctx context.Context) ([]Workflow, error) {
	return it.ListWorkflowWithNamespace(ctx, "")
}

func (it *KubeWorkflowRepository) GetWorkflowByNamespacedName(ctx context.Context, namespace, name string) (WorkflowDetail, error) {
	kubeWorkflow := v1alpha1.Workflow{}
	err := it.kubeclient.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &kubeWorkflow)

	if err != nil {
		return WorkflowDetail{}, err
	}

	workflowNodes := v1alpha1.WorkflowNodeList{}

	// labeling workflow nodes, see pkg/workflow/controllers/new_node.go
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			v1alpha1.LabelWorkflow: kubeWorkflow.Name,
		},
	})
	if err != nil {
		return WorkflowDetail{}, err
	}
	err = it.kubeclient.List(ctx, &workflowNodes, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: selector,
	})
	if err != nil {
		return WorkflowDetail{}, err
	}

	return conversionWorkflowDetail(kubeWorkflow, workflowNodes.Items...), nil
}

func (it *KubeWorkflowRepository) DeleteWorkflowByNamespacedName(ctx context.Context, namespace, name string) error {
	kubeWorkflow := v1alpha1.Workflow{}
	err := it.kubeclient.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &kubeWorkflow)
	if err != nil {
		return err
	}
	return it.kubeclient.Delete(ctx, &kubeWorkflow)
}

// func MutateWithKubeClient could spawn a new repo with the new kubeclient with another auth session.
func (it *KubeWorkflowRepository) MutateWithKubeClient(anotherKubeclient client.Client) *KubeWorkflowRepository {
	return NewKubeWorkflowRepository(anotherKubeclient)
}

func conversionWorkflow(kubeWorkflow v1alpha1.Workflow) Workflow {
	result := Workflow{
		Name:   kubeWorkflow.Name,
		Entry:  kubeWorkflow.Spec.Entry,
	}
	return result
}

func conversionWorkflowDetail(kubeWorkflow v1alpha1.Workflow, kubeNodes ...v1alpha1.WorkflowNode) WorkflowDetail {
	nodes := make([]Node, 0)
	result := WorkflowDetail{
		Workflow: conversionWorkflow(kubeWorkflow),
		Topology: Topology{
			Nodes: nodes,
		},
		CurrentNodes: []Node{},
	}

	for _, item := range kubeNodes {
		node := conversionWorkflowNode(item)
		nodes = append(nodes, node)
	}

	return result
}

func conversionWorkflowNode(kubeWorkflowNode v1alpha1.WorkflowNode) Node {
	result := Node{
		Name:     kubeWorkflowNode.Name,
		Type:     mappingTemplateType(kubeWorkflowNode.Spec.Type),
		Serial:   NodeSerial{Tasks: []string{}},
		Parallel: NodeParallel{Tasks: []string{}},
		Template: kubeWorkflowNode.Spec.TemplateName,
	}

	if kubeWorkflowNode.Spec.Type == v1alpha1.TypeSerial {
		result.Serial.Tasks = append(kubeWorkflowNode.Spec.Tasks)
	} else if kubeWorkflowNode.Spec.Type == v1alpha1.TypeParallel {
		result.Parallel.Tasks = append(kubeWorkflowNode.Spec.Tasks)
	}

	return result
}

func mappingTemplateType(templateType v1alpha1.TemplateType) NodeType {
	// FIXME: automate this part 😓
	switch templateType {
	case v1alpha1.TypeSerial:
		return SerialNode
	case v1alpha1.TypeSuspend:
		return SuspendNode
	case v1alpha1.TypeParallel:
		return ParallelNode
	case v1alpha1.TypeTask:
		return TaskNode
	case v1alpha1.TypeJvmChaos, v1alpha1.TypePodChaos, v1alpha1.TypeNetworkChaos,
		v1alpha1.TypeDnsChaos, v1alpha1.TypeHttpChaos, v1alpha1.TypeIoChaos,
		v1alpha1.TypeKernelChaos, v1alpha1.TypeStressChaos, v1alpha1.TypeTimeChaos:
		return ChaosNode
	default:
		// TODO: logs or error
		return ""
	}
}
