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

	"github.com/mitchellh/mapstructure"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	wfcontrollers "github.com/chaos-mesh/chaos-mesh/pkg/workflow/controllers"
)

type WorkflowRepository interface {
	CreateWorkflowWithRaw(ctx context.Context, raw KubeObjectYAMLDescription) (WorkflowDetail, error)
	UpdateWorkflowWithRaw(ctx context.Context, raw KubeObjectYAMLDescription) (WorkflowDetail, error)
	ListWorkflowWithNamespace(ctx context.Context, namespace string) ([]Workflow, error)
	ListWorkflowFromAllNamespace(ctx context.Context) ([]Workflow, error)
	GetWorkflowByNamespacedName(ctx context.Context, namespace, name string) (WorkflowDetail, error)
	DeleteWorkflowByNamespacedName(ctx context.Context, namespace, name string) error
}

// Workflow defines the root structure of a workflow.
type Workflow struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Entry     string `json:"entry"` // the entry node name
}

type WorkflowDetail struct {
	Workflow `json:",inline"`
	Topology Topology                  `json:"topology"`
	Yaml     KubeObjectYAMLDescription `json:"yaml,omitempty"`
}

// Topology describes the process of a workflow.
type Topology struct {
	Nodes []Node `json:"nodes"`
}

type NodeState string

const (
	NodeRunning NodeState = "Running"
	NodeSucceed NodeState = "Succeed"
	NodeFailed  NodeState = "Failed"
)

// Node defines the single step of a workflow.
type Node struct {
	Name     string       `json:"name"`
	Type     NodeType     `json:"type"`
	State    NodeState    `json:"state"`
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
// There will be five types can be referred as NodeType: Chaos, Serial, Parallel, Suspend, Task.
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

var nodeTypeTemplateTypeMapping = map[v1alpha1.TemplateType]NodeType{
	v1alpha1.TypeSerial:   SerialNode,
	v1alpha1.TypeParallel: ParallelNode,
	v1alpha1.TypeSuspend:  SuspendNode,
	v1alpha1.TypeTask:     TaskNode,
}

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

func (it *KubeWorkflowRepository) CreateWorkflowWithRaw(ctx context.Context, raw KubeObjectYAMLDescription) (WorkflowDetail, error) {
	workflow := v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:        raw.Metadata.Name,
			Namespace:   raw.Metadata.Namespace,
			Labels:      raw.Metadata.Labels,
			Annotations: raw.Metadata.Annotations,
		},
		Spec: v1alpha1.WorkflowSpec{},
	}

	err := mapstructure.Decode(raw.Spec, &workflow.Spec)
	if err != nil {
		return WorkflowDetail{}, err
	}

	// TODO: we need decode the inlined field again, it's better to resolve that in a common methods
	// TODO: same issue in UpdateWorkflowWithRaw, and http api with experiments
	for index := range workflow.Spec.Templates {
		switch parsed := raw.Spec.(type) {
		case map[string]interface{}:
			switch templates := parsed["templates"].(type) {
			case []interface{}:
				target := v1alpha1.EmbedChaos{}
				err := mapstructure.Decode(templates[index], &target)
				if err != nil {
					return WorkflowDetail{}, err
				}
			}

		}
	}

	err = it.kubeclient.Create(ctx, &workflow)
	if err != nil {
		return WorkflowDetail{}, err
	}
	return it.GetWorkflowByNamespacedName(ctx, workflow.Namespace, workflow.Name)
}

func (it *KubeWorkflowRepository) UpdateWorkflowWithRaw(ctx context.Context, raw KubeObjectYAMLDescription) (WorkflowDetail, error) {
	workflow := v1alpha1.Workflow{}

	err := mapstructure.Decode(raw.Spec, &workflow.Spec)
	if err != nil {
		return WorkflowDetail{}, err
	}

	for index := range workflow.Spec.Templates {
		switch parsed := raw.Spec.(type) {
		case map[string]interface{}:
			switch templates := parsed["templates"].(type) {
			case []interface{}:
				target := v1alpha1.EmbedChaos{}
				err := mapstructure.Decode(templates[index], &target)
				if err != nil {
					return WorkflowDetail{}, err
				}
			}

		}
	}

	err = it.kubeclient.Update(ctx, &workflow)
	if err != nil {
		return WorkflowDetail{}, err
	}
	return it.GetWorkflowByNamespacedName(ctx, workflow.Namespace, workflow.Name)
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

	return conversionWorkflowDetail(kubeWorkflow, workflowNodes.Items)
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

func conversionWorkflow(kubeWorkflow v1alpha1.Workflow) Workflow {
	result := Workflow{
		Namespace: kubeWorkflow.Namespace,
		Name:      kubeWorkflow.Name,
		Entry:     kubeWorkflow.Spec.Entry,
	}
	return result
}

func conversionWorkflowDetail(kubeWorkflow v1alpha1.Workflow, kubeNodes []v1alpha1.WorkflowNode) (WorkflowDetail, error) {
	nodes := make([]Node, 0)

	for _, item := range kubeNodes {
		node, err := conversionWorkflowNode(item)
		if err != nil {
			return WorkflowDetail{}, nil
		}
		nodes = append(nodes, node)
	}

	result := WorkflowDetail{
		Workflow: conversionWorkflow(kubeWorkflow),
		Topology: Topology{
			Nodes: nodes,
		},
	}
	return result, nil
}

func conversionWorkflowNode(kubeWorkflowNode v1alpha1.WorkflowNode) (Node, error) {
	templateType, err := mappingTemplateType(kubeWorkflowNode.Spec.Type)
	if err != nil {
		return Node{}, err
	}
	result := Node{
		Name:     kubeWorkflowNode.Name,
		Type:     templateType,
		Serial:   NodeSerial{Tasks: []string{}},
		Parallel: NodeParallel{Tasks: []string{}},
		Template: kubeWorkflowNode.Spec.TemplateName,
	}

	if kubeWorkflowNode.Spec.Type == v1alpha1.TypeSerial {
		result.Serial.Tasks = append(kubeWorkflowNode.Spec.Tasks)
	} else if kubeWorkflowNode.Spec.Type == v1alpha1.TypeParallel {
		result.Parallel.Tasks = append(kubeWorkflowNode.Spec.Tasks)
	}

	// TODO: refactor this
	if wfcontrollers.ConditionEqualsTo(kubeWorkflowNode.Status, v1alpha1.ConditionAccomplished, corev1.ConditionTrue) ||
		wfcontrollers.ConditionEqualsTo(kubeWorkflowNode.Status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionTrue) {
		result.State = NodeSucceed
	} else {
		result.State = NodeRunning
	}

	return result, nil
}

func mappingTemplateType(templateType v1alpha1.TemplateType) (NodeType, error) {
	if v1alpha1.IsChoasTemplateType(templateType) {
		return ChaosNode, nil
	} else if target, ok := nodeTypeTemplateTypeMapping[templateType]; ok {
		return target, nil
	} else {
		return "", errors.Errorf("can not resolve such type called %s", templateType)
	}
}
