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

package repo

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

type KubernetesWorkflowRepo struct {
	operableTrigger trigger.OperableTrigger
	client          client.Client
}

func (it *KubernetesWorkflowRepo) FetchWorkflow(workflowName string) (workflow.WorkflowSpec, workflow.WorkflowStatus, error) {
	panic("implement me")
}

func (it *KubernetesWorkflowRepo) CreateNodes(workflowName string, templates []template.Template, parentNode string) ([]string, error) {
	panic("implement me")
}

func (it *KubernetesWorkflowRepo) UpdateNodesToRunning(workflowName string, nodeName string) error {
	panic("implement me")
}

func (it *KubernetesWorkflowRepo) UpdateNodesToWaitingForChild(workflowName string, nodeName string) error {
	panic("implement me")
}

func (it *KubernetesWorkflowRepo) UpdateNodesToWaitingForSchedule(workflowName string, nodeName string) error {
	panic("implement me")
}
