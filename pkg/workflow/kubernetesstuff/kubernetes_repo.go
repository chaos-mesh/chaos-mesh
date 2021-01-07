// Copyright 2020 Chaos Mesh Authors.
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

package kubernetesstuff

import (
	"context"
	"fmt"

	"k8s.io/client-go/util/retry"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workflowv1alpha1 "github.com/chaos-mesh/chaos-mesh/pkg/workflow/apis/workflow/v1alpha1"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
)

type KubernetesWorkflowRepo struct {
	client         client.Client
	nonCacheReader client.Reader
}

func NewKubernetesWorkflowRepo(client client.Client, nonCacheReader client.Reader) *KubernetesWorkflowRepo {
	return &KubernetesWorkflowRepo{client: client, nonCacheReader: nonCacheReader}
}

func (it *KubernetesWorkflowRepo) FetchWorkflow(namespace, workflowName string) (workflow.WorkflowSpec, workflow.WorkflowStatus, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      workflowName,
	}
	result := workflowv1alpha1.Workflow{}
	// TODO: make context work
	err := it.nonCacheReader.Get(context.TODO(), key, &result)
	if err != nil {
		return nil, nil, err
	}
	return &result, &result, nil
}

func (it *KubernetesWorkflowRepo) CreateNodes(namespace, workflowName, parentNodeName, nodeNames, templateName string) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		key := types.NamespacedName{
			Namespace: namespace,
			Name:      workflowName,
		}
		target := workflowv1alpha1.Workflow{}
		// TODO: make context work
		err := it.client.Get(context.TODO(), key, &target)
		if err != nil {
			return err
		}
		copied := target.DeepCopy()
		if len(copied.Status.Nodes) == 0 {
			copied.Status.Nodes = make(map[string]workflowv1alpha1.Node)
		}
		copied.Status.Nodes[nodeNames] = workflowv1alpha1.Node{
			Name:         nodeNames,
			ParentNode:   parentNodeName,
			NodePhase:    node.Init,
			TemplateName: templateName,
		}
		// TODO: make context work
		err = it.client.Update(context.TODO(), copied)
		return err
	})

}

func (it *KubernetesWorkflowRepo) UpdateWorkflowPhase(namespace, workflowName string, newPhase workflow.WorkflowPhase) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		key := types.NamespacedName{
			Namespace: namespace,
			Name:      workflowName,
		}
		target := workflowv1alpha1.Workflow{}
		// TODO: make context work
		err := it.client.Get(context.TODO(), key, &target)
		if err != nil {
			return err
		}
		copied := target.DeepCopy()
		copied.Status.Phase = newPhase
		// TODO: make context work
		err = it.client.Update(context.TODO(), copied)
		return err
	})

}

func (it *KubernetesWorkflowRepo) UpdateNodePhase(namespace, workflowName, nodeName string, newPhase node.NodePhase) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		key := types.NamespacedName{
			Namespace: namespace,
			Name:      workflowName,
		}
		target := workflowv1alpha1.Workflow{}
		// TODO: make context work
		err := it.client.Get(context.TODO(), key, &target)
		if err != nil {
			return err
		}
		copied := target.DeepCopy()
		if targetNode, exist := copied.Status.Nodes[nodeName]; exist {
			targetNode.NodePhase = newPhase
			copied.Status.Nodes[nodeName] = targetNode
		} else {
			return fmt.Errorf("no such node called %s", nodeName)
		}
		// TODO: make context work
		err = it.client.Update(context.TODO(), copied)
		return err

	})
}
