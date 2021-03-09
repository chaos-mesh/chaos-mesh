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

package scheduler

import (
	"context"
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/model/workflow"
)

type serialScheduler struct {
	workflowSpec workflow.WorkflowSpec
	nodeStatus   node.Node
	treeNode     node.NodeTreeNode
}

func NewSerialScheduler(workflowSpec workflow.WorkflowSpec, nodeStatus node.Node, treeNode node.NodeTreeNode) *serialScheduler {
	return &serialScheduler{workflowSpec: workflowSpec, nodeStatus: nodeStatus, treeNode: treeNode}
}

func (it *serialScheduler) ScheduleNext(ctx context.Context) (nextTemplates []template.Template, parentNodeName string, err error) {
	op := "serialScheduler.ScheduleNext"
	parentTemplate, err := it.workflowSpec.FetchTemplateByName(it.nodeStatus.TemplateName())
	if err != nil {
		return nil, "", err
	}

	if parentTemplate.TemplateType() != template.Serial {
		return nil, "", fmt.Errorf("%s is not serail", it.nodeStatus.TemplateName())
	}

	serialTemplate, err := template.ParseSerialTemplate(parentTemplate)
	if err != nil {
		return nil, "", err
	}

	childrenTemplateNames := serialTemplate.SerialChildrenList()

	if it.treeNode.Children().Length() >= len(childrenTemplateNames) {
		// TODO: unexpected situation, warn log
		return nil, "", errors.NewNoMoreTemplateInSerialTemplateError(op, it.workflowSpec.Name(), it.treeNode.TemplateName(), it.treeNode.Name())
	}

	targetName := childrenTemplateNames[it.treeNode.Children().Length()]
	targetTemplate, err := it.workflowSpec.FetchTemplateByName(targetName)
	if err != nil {
		return nil, "", err
	}

	return []template.Template{targetTemplate}, it.nodeStatus.Name(), nil
}
