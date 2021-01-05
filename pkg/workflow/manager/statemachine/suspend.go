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

package statemachine

import (
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/manager/sideeffect"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

type SuspendStateMachine struct {
	workflowSpec      workflow.WorkflowSpec
	nodeStatus        node.Node
	treeNode          node.NodeTreeNode
	nodeNameGenerator node.NodeNameGenerator
}

func NewSuspendStateMachine(workflowSpec workflow.WorkflowSpec, nodeStatus node.Node, treeNode node.NodeTreeNode, nodeNameGenerator node.NodeNameGenerator) *SuspendStateMachine {
	return &SuspendStateMachine{workflowSpec: workflowSpec, nodeStatus: nodeStatus, treeNode: treeNode, nodeNameGenerator: nodeNameGenerator}
}

func (it *SuspendStateMachine) GetName() string {
	return "SuspendStateMachine"
}

func (it *SuspendStateMachine) HandleEvent(event trigger.Event) ([]sideeffect.SideEffect, error) {
	switch event.GetEventType() {
	case trigger.NodeCreated:
		if it.nodeStatus.GetNodePhase() == node.Init {
			targetTemplate, err := it.workflowSpec.FetchTemplateByName(it.nodeStatus.GetTemplateName())
			if err != nil {
				return nil, err
			}
			suspendTemplate, err := template.ParseSuspendTemplate(targetTemplate)
			if err != nil {
				return nil, err
			}
			holdingDuration, err := suspendTemplate.GetDuration()
			if err != nil {
				return nil, err
			}

			var result []sideeffect.SideEffect
			result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.Holding))
			result = append(result, sideeffect.NewNotifyNewDelayEventSideEffect(trigger.NewEvent(it.workflowSpec.GetName(), it.nodeStatus.GetName(), trigger.NodeHoldingAwake), holdingDuration))
			return result, nil
		}
		// TODO: replace this error
		return nil, fmt.Errorf("StateMachine %s can not handle evnet %s at status %s", it.GetName(), event, it.nodeStatus)
	case trigger.NodeHoldingAwake:
		// TODO: assert current state
		var result []sideeffect.SideEffect
		result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.Succeed))
		result = append(result, sideeffect.NewNotifyNewEventSideEffect(trigger.NewEvent(it.workflowSpec.GetName(), it.nodeStatus.GetParentNodeName(), trigger.ChildNodeSucceed)))
		return result, nil
	default:
		return nil, fmt.Errorf("StateMachine %s could not resolve event %s", it.GetName(), event)
	}
}
