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

package workflowrepo

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
)

// Notice when you update node's status, please make sure already finish the operations.
type WorkflowRepo interface {
	FetchWorkflow(namespace, workflowName string) (workflow.WorkflowSpec, workflow.WorkflowStatus, error)
	// func CreateNodes, create the nodes by given templates and parent node.
	// It return the nodes name with the same order of template.
	CreateNodes(namespace, workflowName, parentNodeName, nodeNames, templateName string) error
	// func UpdateWorkflowPhase, update certain workflow to new phase
	UpdateWorkflowPhase(namespace, workflowName string, newPhase workflow.WorkflowPhase) error
	// func UpdateNodePhase, update certain node to new phase
	UpdateNodePhase(namespace, workflowName, nodeName string, newPhase node.NodePhase) error
}
