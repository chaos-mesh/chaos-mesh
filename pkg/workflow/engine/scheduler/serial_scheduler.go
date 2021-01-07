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

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
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
	parentTemplate, err := it.workflowSpec.FetchTemplateByName(it.nodeStatus.GetTemplateName())
	if err != nil {
		return nil, "", err
	}

	if parentTemplate.GetTemplateType() != template.Serial {
		return nil, "", fmt.Errorf("%s is not serail", it.nodeStatus.GetTemplateName())
	}

	serialTemplate, err := template.ParseSerialTemplate(parentTemplate)
	if err != nil {
		return nil, "", err
	}

	childrenTemplateNames := serialTemplate.GetSerialChildrenList()
	var childrenTemplates []template.Template
	for _, name := range childrenTemplateNames {
		item, err := it.workflowSpec.FetchTemplateByName(name)
		if err != nil {
			return nil, "", err
		}
		childrenTemplates = append(childrenTemplates, item)
	}
	if it.treeNode.GetChildren().Length() == len(childrenTemplates) {
		// TODO: unexpected situation, warn log
		return nil, "", errors.NewNoMoreTemplateInSerialTemplateError(op, it.workflowSpec.GetName(), it.treeNode.GetTemplateName(), it.treeNode.GetName())
	}

	for _, item := range childrenTemplates {
		if it.treeNode.GetChildren().ContainsTemplate(item.GetName()) {
			continue
		}
		// TODO: debug logs
		return []template.Template{item}, it.nodeStatus.GetName(), nil
	}

	return nil, it.nodeStatus.GetName(), nil
}
