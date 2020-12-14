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
	goerror "errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
)

func NewBasicScheduler(workflowSpec workflow.WorkflowSpec, workflowStatus workflow.WorkflowStatus) *basicScheduler {
	return &basicScheduler{workflowSpec: workflowSpec, workflowStatus: workflowStatus}
}

type basicScheduler struct {
	workflowSpec   workflow.WorkflowSpec
	workflowStatus workflow.WorkflowStatus
}

func (it *basicScheduler) ScheduleNext(ctx context.Context) ([]template.Template, string, error) {
	op := "basicScheduler.ScheduleNext"

	nodesMap := it.workflowStatus.FetchNodesMap()
	if len(nodesMap) == 0 {
		// first schedule
		templates := it.workflowSpec.GetTemplates()
		if templates == nil {
			return nil, "", errors.NewNoTemplatesError(op, it.workflowSpec.GetName())
		}
		entry := templates.GetByTemplateName(it.workflowSpec.GetEntry())
		return []template.Template{entry}, "", nil
	} else {
		var uncompletedSchedulableCompositeNode node.Node
		for _, item := range it.workflowStatus.GetNodes() {
			if item.GetNodePhase() == node.WaitingForSchedule {
				uncompletedSchedulableCompositeNode = item
				break
			}
		}

		if uncompletedSchedulableCompositeNode == nil {
			return nil, "", errors.NewNoNeedScheduleError(op, it.workflowSpec.GetName())
		}

		templates, err := it.fetchChildrenForCompositeNode(uncompletedSchedulableCompositeNode.GetName())
		return templates, uncompletedSchedulableCompositeNode.GetName(), err

	}

}

func (it *basicScheduler) fetchChildrenForCompositeNode(parentNodeName string) ([]template.Template, error) {
	op := "basicScheduler.fetchChildrenForCompositeNode"

	if parentNode, ok := it.workflowStatus.FetchNodesMap()[parentNodeName]; ok {
		if nodeTreeNode := it.workflowStatus.GetNodesTree().FetchChildNodeByNodeName(parentNodeName); nodeTreeNode != nil {
			childrenNodes := nodeTreeNode.GetChildren()
			parentTemplate := it.workflowSpec.GetTemplates().GetByTemplateName(parentNode.GetTemplateName())
			// Serial template execute its children template one-by-one, so it need found out previous one
			// is completed or not, then pick the next one.
			// Parallel and Task execute all children at once, so it should runs into here only one time.
			if parentNode.GetTemplateType() == template.Serial {
				serialTemplate, err := template.ParseSerialTemplate(parentTemplate)
				if err != nil {
					return nil, err
				}
				result, err := fetchNextTemplateFromSerial(childrenNodes, serialTemplate.GetSerialChildrenList())
				return []template.Template{result}, err

			} else if parentNode.GetTemplateType() == template.Parallel {
				parallelTemplate, err := template.ParseParallelTemplate(parentTemplate)
				if err != nil {
					return nil, err
				}
				return parallelTemplate.GetParallelChildrenList(), nil
			} else if parentNode.GetTemplateType() == template.Task {
				taskTemplate, err := template.ParseTaskTemplate(parentTemplate)
				if err != nil {
					return nil, err
				}
				taskNode, err := node.ParseTaskNode(parentNode)
				if err != nil {
					return nil, err
				}
				results, err := fetchAllAvailableTemplatesFromTask(taskNode, taskTemplate.GetAllTemplates())
				return results, err
			} else {
				return nil, errors.NewUnsupportedNodeTypeError(op, parentNode.GetName(), parentNode.GetTemplateType(), it.workflowSpec.GetName())
			}
		} else {
			return nil, errors.NewNoSuchTreeNodeError(op, parentNodeName, it.workflowSpec.GetName())
		}
	} else {
		return nil, errors.NewNoSuchNodeError(op, parentNodeName, it.workflowSpec.GetName())
	}
}

func fetchNextTemplateFromSerial(status node.NodeTreeChildren, childrenTemplates []template.Template) (template.Template, error) {
	if status.Length() == len(childrenTemplates) {
		// TODO: unexpected situation, warn log
		return nil, errors.NewNoMoreTemplateInSerialTemplateError()
	} else {
		for _, item := range childrenTemplates {
			if status.ContainsTemplate(item.GetName()) {
				continue
			} else {
				// TODO: debug logs
				return item, nil
			}
		}

		// TODO: warn logs
		return nil, errors.NewNoMoreTemplateInSerialTemplateError()
	}
}

func fetchAllAvailableTemplatesFromTask(status node.TaskNode, allChildrenTemplates []template.Template) ([]template.Template, error) {
	nameOfAvailableTemplates, err := status.FetchAvailableChildren()
	if err != nil {
		return nil, err
	}

	var result []template.Template

	for _, name := range nameOfAvailableTemplates {
		for _, item := range allChildrenTemplates {
			if item.GetName() == name {
				result = append(result, item)
				break
			}
		}
		return nil, errors.NewNoSuchTemplateError()
	}

	return result, err
}

func IsNoNeedSchedule(err error) bool {
	return goerror.Is(err, errors.ErrNoNeedSchedule)
}
