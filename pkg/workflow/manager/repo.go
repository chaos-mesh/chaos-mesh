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

type WorkflowRepo interface {
	FetchWorkflow(workflowName string) (workflow.WorkflowSpec, workflow.WorkflowStatus, error)
	CreateNodes(workflowName string, templates []template.Template, parentNode string) ([]string, error)
	UpdateNodesToRunning(workflowName string, nodeName string) error
	UpdateNodesToWaitingForChild(workflowName string, nodeName string) error
	UpdateNodesToWaitingForSchedule(workflowName string, nodeName string) error
}
