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

package manager

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
)

// Notice when you update node's status, please make sure already finish the operations.
type WorkflowRepo interface {
	FetchWorkflow(workflowName string) (workflow.WorkflowSpec, workflow.WorkflowStatus, error)
	// func CreateNodes, create the nodes by given templates and parent node.
	// It return the nodes name with the same order of template.
	CreateNodes(workflowName string, templates []template.Template, parentNode string) ([]string, error)
	// func UpdateNodesToRunning, update one specific node status to Running.
	UpdateNodesToRunning(workflowName string, nodeName string) error
	// func UpdateNodesToWaitingForChild, update one specific node status to WaitingForChild.
	UpdateNodesToWaitingForChild(workflowName string, nodeName string) error
	// func UpdateNodesToWaitingForSchedule, update one specific node status to WaitingForSchedule.
	UpdateNodesToWaitingForSchedule(workflowName string, nodeName string) error
}
