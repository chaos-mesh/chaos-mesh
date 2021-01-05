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
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

type KubernetesWorkflowRepo struct {
	operableTrigger trigger.OperableTrigger
	client          client.Client
}

func (it *KubernetesWorkflowRepo) FetchWorkflow(namespace, workflowName string) (workflow.WorkflowSpec, workflow.WorkflowStatus, error) {
	panic("implement me")
}

func (it *KubernetesWorkflowRepo) CreateNodes(namespace, workflowName, parentNodeName, nodeNames, templateName string) error {
	panic("implement me")
}

func (it *KubernetesWorkflowRepo) UpdateWorkflowPhase(namespace, workflowName string, newPhase workflow.WorkflowPhase) error {
	panic("implement me")
}

func (it *KubernetesWorkflowRepo) UpdateNodePhase(namespace, workflowName, nodeName string, newPhase node.NodePhase) error {
	panic("implement me")
}
