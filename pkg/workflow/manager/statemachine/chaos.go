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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	chaosmeshv1alph1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/actor"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/manager/sideeffect"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

type ChaosStateMachine struct {
	workflowSpec      workflow.WorkflowSpec
	nodeStatus        node.Node
	treeNode          node.NodeTreeNode
	nodeNameGenerator node.NodeNameGenerator
}

func (it *ChaosStateMachine) GetName() string {
	return "ChaosStateMachine"
}

func (it *ChaosStateMachine) HandleEvent(event trigger.Event) ([]sideeffect.SideEffect, error) {
	switch event.GetEventType() {
	case trigger.NodeCreated:
		// TODO: assert current state
		targetTemplate, err := it.workflowSpec.FetchTemplateByName(it.nodeStatus.GetTemplateName())
		if err != nil {
			return nil, err
		}
		chaosTemplate, err := template.ParseNetworkChaosTemplate(targetTemplate)
		if err != nil {
			return nil, err
		}

		networkChaosSpec := chaosTemplate.FetchNetworkChaosSpec()
		targetChaos := chaosmeshv1alph1.NetworkChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name: it.nodeStatus.GetName(),
				// FIXME: namespace is not configured
				Namespace: "",
			},
			Spec: networkChaosSpec,
		}

		var result []sideeffect.SideEffect
		result = append(
			result,
			sideeffect.NewUpdatePhaseStatusSideEffect(
				it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.Running,
			),
		)
		result = append(
			result,
			sideeffect.NewCreateActorEventSideEffect(actor.NewCreateNetworkChaosActor(targetChaos)),
		)
		result = append(
			result,
			sideeffect.NewNotifyNewEventSideEffect(
				trigger.NewEvent(it.workflowSpec.GetName(), it.nodeStatus.GetName(), trigger.NodeChaosInjectSucceed),
			))
		return result, nil
	case trigger.NodeChaosInjectSucceed:
		// TODO: assert current state
		targetTemplate, err := it.workflowSpec.FetchTemplateByName(it.nodeStatus.GetTemplateName())
		if err != nil {
			return nil, err
		}
		chaosTemplate, err := template.ParseNetworkChaosTemplate(targetTemplate)
		if err != nil {
			return nil, err
		}
		holdingDuration, err := chaosTemplate.GetDuration()
		if err != nil {
			return nil, err
		}

		var result []sideeffect.SideEffect
		result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.Holding))
		result = append(result, sideeffect.NewNotifyNewDelayEventSideEffect(trigger.NewEvent(it.workflowSpec.GetName(), it.nodeStatus.GetName(), trigger.NodeHoldingAwake), holdingDuration))
		return result, nil
	case trigger.NodeHoldingAwake:
		// TODO: assert current state
		var result []sideeffect.SideEffect
		result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.Succeed))
		result = append(result, sideeffect.NewNotifyNewEventSideEffect(trigger.NewEvent(it.workflowSpec.GetName(), it.nodeStatus.GetParentNodeName(), trigger.ChildNodeSucceed)))
		return result, nil
	default:
		// TODO: replace this error
		return nil, fmt.Errorf("StateMachine %s could not resolve event %s", it.GetName(), event)
	}
}
