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
	"context"
	"errors"
	"fmt"

	engineerrors "github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/scheduler"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/manager/sideeffect"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

type SerialStateMachine struct {
	namespace         string
	workflowSpec      workflow.WorkflowSpec
	nodeStatus        node.Node
	treeNode          node.NodeTreeNode
	nodeNameGenerator node.NodeNameGenerator
}

func NewSerialStateMachine(namespace string, workflowSpec workflow.WorkflowSpec, nodeStatus node.Node, treeNode node.NodeTreeNode, nodeNameGenerator node.NodeNameGenerator) *SerialStateMachine {
	return &SerialStateMachine{namespace: namespace, workflowSpec: workflowSpec, nodeStatus: nodeStatus, treeNode: treeNode, nodeNameGenerator: nodeNameGenerator}
}

func (it *SerialStateMachine) GetName() string {
	return "SerialStateMachine"
}

func (it *SerialStateMachine) HandleEvent(event trigger.Event) ([]sideeffect.SideEffect, error) {
	switch event.GetEventType() {
	case trigger.NodeCreated:
		if it.nodeStatus.GetNodePhase() == node.Init {
			var result []sideeffect.SideEffect
			result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.WaitingForSchedule))
			result = append(result, sideeffect.NewNotifyNewEventSideEffect(trigger.NewEvent(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), trigger.NodePickChildToSchedule)))
			return result, nil
		}
		// TODO: replace this error
		return nil, fmt.Errorf("StateMachine %s can not handle evnet %s at status %s", it.GetName(), event, it.nodeStatus)

	case trigger.NodePickChildToSchedule:
		if it.nodeStatus.GetNodePhase() == node.WaitingForSchedule {
			templates, _, err := scheduler.NewSerialScheduler(it.workflowSpec, it.nodeStatus, it.treeNode).ScheduleNext(context.TODO())
			if err != nil {
				return nil, err
			}

			// create nodes
			var newNodes []tempNode
			for _, templateItem := range templates {
				newNodes = append(newNodes, tempNode{
					nodeName:     it.nodeNameGenerator.GenerateNodeName(templateItem.GetName()),
					templateName: templateItem.GetName(),
					nodePhase:    node.Init,
				})
			}

			var result []sideeffect.SideEffect
			result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.WaitingForChild))
			for _, nodeItem := range newNodes {
				result = append(result, sideeffect.NewCreateNewNodeSideEffect(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), nodeItem.nodeName, nodeItem.templateName, nodeItem.nodePhase))
			}
			for _, nodeItem := range newNodes {
				result = append(result, sideeffect.NewNotifyNewEventSideEffect(trigger.NewEvent(it.namespace, it.workflowSpec.GetName(), nodeItem.nodeName, trigger.NodeCreated)))
			}
			return result, nil
		}
		// TODO: replace this error
		return nil, fmt.Errorf("StateMachine %s can not handle evnet %s at status %s", it.GetName(), event, it.nodeStatus)

	case trigger.ChildNodeSucceed:
		// TODO: assert current state
		_, _, err := scheduler.NewSerialScheduler(it.workflowSpec, it.nodeStatus, it.treeNode).ScheduleNext(context.TODO())
		if err != nil {
			if errors.Is(err, engineerrors.ErrNoMoreTemplateInSerialTemplate) {
				var result []sideeffect.SideEffect
				result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.Succeed))
				result = append(result, sideeffect.NewNotifyNewEventSideEffect(trigger.NewEvent(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetParentNodeName(), trigger.ChildNodeSucceed)))
				return result, nil
			}
			return nil, err
		}

		// still have schedulable child
		var result []sideeffect.SideEffect
		result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.WaitingForSchedule))
		result = append(result, sideeffect.NewNotifyNewEventSideEffect(trigger.NewEvent(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), trigger.NodePickChildToSchedule)))
		return result, nil

	case trigger.ChildNodeFailed:
		panic("unimplemented")
	default:
		// TODO: replace this error
		return nil, fmt.Errorf("StateMachine %s could not resolve event %s", it.GetName(), event)
	}
}

type tempNode struct {
	nodeName     string
	templateName string
	nodePhase    node.NodePhase
}
