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
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	wfcontrollers "github.com/chaos-mesh/chaos-mesh/pkg/workflow/controllers"
)

type WorkflowRepository interface {
	List(ctx context.Context) ([]WorkflowMeta, error)
	ListByNamespace(ctx context.Context, namespace string) ([]WorkflowMeta, error)
	Create(ctx context.Context, workflow v1alpha1.Workflow) (WorkflowDetail, error)
	Get(ctx context.Context, namespace, name string) (WorkflowDetail, error)
	Delete(ctx context.Context, namespace, name string) error
	Update(ctx context.Context, namespace, name string, workflow v1alpha1.Workflow) (WorkflowDetail, error)
}

type WorkflowStatus string

const (
	WorkflowRunning WorkflowStatus = "Running"
	WorkflowSucceed WorkflowStatus = "Succeed"
	WorkflowFailed  WorkflowStatus = "Failed"
	WorkflowUnknown WorkflowStatus = "Unknown"
)

// Workflow defines the root structure of a workflow.
type WorkflowMeta struct {
	ID        uint   `gorm:"primary_key" json:"id"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	// the entry node name
	Entry     string         `json:"entry"`
	CreatedAt string         `json:"created_at"`
	EndTime   string         `json:"end_time"`
	Status    WorkflowStatus `json:"status,omitempty"`
	UID       string         `gorm:"index:uid" json:"uid"`
}

type WorkflowDetail struct {
	WorkflowMeta `json:",inline"`
	Topology     Topology       `json:"topology"`
	KubeObject   KubeObjectDesc `json:"kube_object,omitempty"`
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

// Node defines a single step of a workflow.
type Node struct {
	Name     string        `json:"name"`
	Type     NodeType      `json:"type"`
	State    NodeState     `json:"state"`
	Serial   *NodeSerial   `json:"serial,omitempty"`
	Parallel *NodeParallel `json:"parallel,omitempty"`
	Template string        `json:"template"`
	UID      string        `json:"uid"`
}

type NodeNameWithTemplate struct {
	Name     string `json:"name,omitempty"`
	Template string `json:"template,omitempty"`
}

// NodeSerial defines SerialNode's specific fields.
type NodeSerial struct {
	Tasks []NodeNameWithTemplate `json:"tasks"`
}

// NodeParallel defines ParallelNode's specific fields.
type NodeParallel struct {
	Tasks []NodeNameWithTemplate `json:"tasks"`
}

// NodeType represents the type of a workflow node.
//
// There will be five types can be referred as NodeType:
// ChaosNode, SerialNode, ParallelNode, SuspendNode, TaskNode.
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

type KubeWorkflowRepository struct {
	kubeclient client.Client
}

func NewKubeWorkflowRepository(kubeclient client.Client) *KubeWorkflowRepository {
	return &KubeWorkflowRepository{kubeclient: kubeclient}
}

func (it *KubeWorkflowRepository) Create(ctx context.Context, workflow v1alpha1.Workflow) (WorkflowDetail, error) {
	err := it.kubeclient.Create(ctx, &workflow)
	if err != nil {
		return WorkflowDetail{}, err
	}

	return it.Get(ctx, workflow.Namespace, workflow.Name)
}

func (it *KubeWorkflowRepository) Update(ctx context.Context, namespace, name string, workflow v1alpha1.Workflow) (WorkflowDetail, error) {
	current := v1alpha1.Workflow{}

	err := it.kubeclient.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &current)
	if err != nil {
		return WorkflowDetail{}, err
	}
	workflow.ObjectMeta.ResourceVersion = current.ObjectMeta.ResourceVersion

	err = it.kubeclient.Update(ctx, &workflow)
	if err != nil {
		return WorkflowDetail{}, err
	}

	return it.Get(ctx, workflow.Namespace, workflow.Name)
}

func (it *KubeWorkflowRepository) ListByNamespace(ctx context.Context, namespace string) ([]WorkflowMeta, error) {
	workflowList := v1alpha1.WorkflowList{}

	err := it.kubeclient.List(ctx, &workflowList, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return nil, err
	}

	var result []WorkflowMeta
	for _, item := range workflowList.Items {
		result = append(result, convertWorkflow(item))
	}

	return result, nil
}

func (it *KubeWorkflowRepository) List(ctx context.Context) ([]WorkflowMeta, error) {
	return it.ListByNamespace(ctx, "")
}

func (it *KubeWorkflowRepository) Get(ctx context.Context, namespace, name string) (WorkflowDetail, error) {
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

	return convertWorkflowDetail(kubeWorkflow, workflowNodes.Items)
}

func (it *KubeWorkflowRepository) Delete(ctx context.Context, namespace, name string) error {
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

func convertWorkflow(kubeWorkflow v1alpha1.Workflow) WorkflowMeta {
	result := WorkflowMeta{
		Namespace: kubeWorkflow.Namespace,
		Name:      kubeWorkflow.Name,
		Entry:     kubeWorkflow.Spec.Entry,
		UID:       string(kubeWorkflow.UID),
	}

	if kubeWorkflow.Status.StartTime != nil {
		result.CreatedAt = kubeWorkflow.Status.StartTime.Format(time.RFC3339)
	}

	if kubeWorkflow.Status.EndTime != nil {
		result.EndTime = kubeWorkflow.Status.EndTime.Format(time.RFC3339)
	}

	if wfcontrollers.WorkflowConditionEqualsTo(kubeWorkflow.Status, v1alpha1.WorkflowConditionAccomplished, corev1.ConditionTrue) {
		result.Status = WorkflowSucceed
	} else if wfcontrollers.WorkflowConditionEqualsTo(kubeWorkflow.Status, v1alpha1.WorkflowConditionScheduled, corev1.ConditionTrue) {
		result.Status = WorkflowRunning
	} else {
		result.Status = WorkflowUnknown
	}

	// TODO: status failed

	return result
}

func convertWorkflowDetail(kubeWorkflow v1alpha1.Workflow, kubeNodes []v1alpha1.WorkflowNode) (WorkflowDetail, error) {
	nodes := make([]Node, 0)

	for _, item := range kubeNodes {
		node, err := convertWorkflowNode(item)
		if err != nil {
			return WorkflowDetail{}, nil
		}

		nodes = append(nodes, node)
	}

	result := WorkflowDetail{
		WorkflowMeta: convertWorkflow(kubeWorkflow),
		Topology: Topology{
			Nodes: nodes,
		},
		KubeObject: KubeObjectDesc{
			TypeMeta: kubeWorkflow.TypeMeta,
			Meta: KubeObjectMeta{
				Name:        kubeWorkflow.Name,
				Namespace:   kubeWorkflow.Namespace,
				Labels:      kubeWorkflow.Labels,
				Annotations: kubeWorkflow.Annotations,
			},
			Spec: kubeWorkflow.Spec,
		},
	}

	return result, nil
}

func convertWorkflowNode(kubeWorkflowNode v1alpha1.WorkflowNode) (Node, error) {
	templateType, err := mappingTemplateType(kubeWorkflowNode.Spec.Type)
	if err != nil {
		return Node{}, err
	}

	result := Node{
		Name:     kubeWorkflowNode.Name,
		Type:     templateType,
		Serial:   nil,
		Parallel: nil,
		Template: kubeWorkflowNode.Spec.TemplateName,
		UID:      string(kubeWorkflowNode.UID),
	}

	if kubeWorkflowNode.Spec.Type == v1alpha1.TypeSerial {
		var nodes []string
		for _, child := range kubeWorkflowNode.Status.FinishedChildren {
			nodes = append(nodes, child.Name)
		}
		for _, child := range kubeWorkflowNode.Status.ActiveChildren {
			nodes = append(nodes, child.Name)
		}
		result.Serial = &NodeSerial{
			Tasks: composeSerialTaskAndNodes(kubeWorkflowNode.Spec.Tasks, nodes),
		}
	} else if kubeWorkflowNode.Spec.Type == v1alpha1.TypeParallel {
		var nodes []string
		for _, child := range kubeWorkflowNode.Status.FinishedChildren {
			nodes = append(nodes, child.Name)
		}
		for _, child := range kubeWorkflowNode.Status.ActiveChildren {
			nodes = append(nodes, child.Name)
		}
		result.Parallel = &NodeParallel{
			Tasks: composeParallelTaskAndNodes(kubeWorkflowNode.Spec.Tasks, nodes),
		}
	}

	if wfcontrollers.WorkflowNodeFinished(kubeWorkflowNode.Status) {
		result.State = NodeSucceed
	} else {
		result.State = NodeRunning
	}

	return result, nil
}

// composeSerialTaskAndNodes need nodes to be ordered with its creation time
func composeSerialTaskAndNodes(tasks []string, nodes []string) []NodeNameWithTemplate {
	var result []NodeNameWithTemplate
	for _, node := range nodes {
		// TODO: that reverse the generated name, maybe we could use WorkflowNode.TemplateName in the future
		templateName := node[0:strings.LastIndex(node, "-")]
		result = append(result, NodeNameWithTemplate{Name: node, Template: templateName})
	}
	for _, task := range tasks[len(nodes):] {
		result = append(result, NodeNameWithTemplate{Template: task})
	}
	return result
}

func composeParallelTaskAndNodes(tasks []string, nodes []string) []NodeNameWithTemplate {
	var result []NodeNameWithTemplate
	for _, task := range tasks {
		result = append(result, NodeNameWithTemplate{
			Name:     "",
			Template: task,
		})
	}
	for _, node := range nodes {
		for i, item := range result {
			if len(item.Name) == 0 && strings.HasPrefix(node, item.Template) {
				result[i].Name = node
				break
			}
		}
	}
	return result
}

func mappingTemplateType(templateType v1alpha1.TemplateType) (NodeType, error) {
	if v1alpha1.IsChaosTemplateType(templateType) {
		return ChaosNode, nil
	} else if target, ok := nodeTypeTemplateTypeMapping[templateType]; ok {
		return target, nil
	} else {
		return "", errors.Errorf("can not resolve such type called %s", templateType)
	}
}

// The WorkflowStore of workflow is not so similar with others store.
type WorkflowStore interface {
	List(ctx context.Context, namespace, name string, archived bool) ([]*WorkflowEntity, error)
	ListMeta(ctx context.Context, namespace, name string, archived bool) ([]*WorkflowMeta, error)
	FindByID(ctx context.Context, ID uint) (*WorkflowEntity, error)
	FindByUID(ctx context.Context, UID string) (*WorkflowEntity, error)
	FindMetaByUID(ctx context.Context, UID string) (*WorkflowMeta, error)
	Save(ctx context.Context, entity WorkflowEntity) error
	DeleteByUID(ctx context.Context, UID string) error
	DeleteByUIDs(ctx context.Context, UIDs []string) error
	MarkAsArchived(ctx context.Context, namespace, name string) error
	MarkAsArchivedWithUID(ctx context.Context, UID string) error
}

// WorkflowEntity is the gorm entity, refers to a row of data
type WorkflowEntity struct {
	WorkflowMeta
	Workflow string `gorm:"size:32768"`
}

func WorkflowCR2WorkflowEntity(workflow *v1alpha1.Workflow) (*WorkflowEntity, error) {
	if workflow == nil {
		return nil, nil

	}
	jsonContent, err := json.Marshal(workflow)
	if err != nil {
		return nil, err
	}
	return &WorkflowEntity{
		WorkflowMeta: convertWorkflow(*workflow),
		Workflow:     string(jsonContent),
	}, nil

}

func WorkflowEntity2WorkflowCR(entity *WorkflowEntity) (*v1alpha1.Workflow, error) {
	if entity == nil {
		return nil, nil
	}
	result := v1alpha1.Workflow{}
	err := json.Unmarshal([]byte(entity.Workflow), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
