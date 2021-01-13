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
	namespace    string
	workflowSpec workflow.WorkflowSpec
	nodeStatus   node.Node
}

func NewChaosStateMachine(namespace string, workflowSpec workflow.WorkflowSpec, nodeStatus node.Node) *ChaosStateMachine {
	return &ChaosStateMachine{namespace: namespace, workflowSpec: workflowSpec, nodeStatus: nodeStatus}
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

		// TODO: multi type of chaos injection
		var chaosActor actor.Actor

		if targetTemplate.GetTemplateType() == template.NetworkChaos {
			chaosTemplate, err := template.ParseNetworkChaosTemplate(targetTemplate)
			if err != nil {
				return nil, err
			}

			networkChaosSpec := chaosTemplate.FetchNetworkChaosSpec()
			targetChaos := chaosmeshv1alph1.NetworkChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      it.nodeStatus.GetName(),
					Namespace: it.namespace,
				},
				Spec: networkChaosSpec,
			}

			chaosActor = actor.NewCreateNetworkChaosActor(targetChaos)
		} else if targetTemplate.GetTemplateType() == template.PodChaos {
			chaosTemplate, err := template.ParsePodChaosTemplate(targetTemplate)
			if err != nil {
				return nil, err
			}
			podChaosSpec := chaosTemplate.FetchPodChaos()
			targetChaos := chaosmeshv1alph1.PodChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      it.nodeStatus.GetName(),
					Namespace: it.namespace,
				},
				Spec: podChaosSpec,
			}
			chaosActor = actor.NewCreatePodChaosActor(targetChaos)
		} else {
			// TODO: replace this error
			return nil, fmt.Errorf("%s is not supported now", targetTemplate.GetTemplateType())
		}

		var result []sideeffect.SideEffect
		result = append(
			result,
			sideeffect.NewUpdatePhaseStatusSideEffect(
				it.namespace,
				it.workflowSpec.GetName(),
				it.nodeStatus.GetName(),
				it.nodeStatus.GetNodePhase(),
				node.Running,
			),
		)
		result = append(
			result,
			sideeffect.NewCreateActorEventSideEffect(chaosActor),
		)
		result = append(
			result,
			sideeffect.NewNotifyNewEventSideEffect(
				trigger.NewEvent(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), trigger.NodeChaosInjected),
			))
		return result, nil
	case trigger.NodeChaosInjected:
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
		result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.Holding))
		result = append(result, sideeffect.NewNotifyNewDelayEventSideEffect(trigger.NewEvent(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), trigger.NodeHoldingAwake), holdingDuration))
		return result, nil
	case trigger.NodeHoldingAwake:
		// TODO: assert current state
		targetTemplate, err := it.workflowSpec.FetchTemplateByName(it.nodeStatus.GetTemplateName())
		if err != nil {
			return nil, err
		}

		var chaosActor actor.Actor

		if targetTemplate.GetTemplateType() == template.NetworkChaos {
			chaosActor = actor.NewDeleteNetworkChaosActor(it.namespace, it.nodeStatus.GetName())
		} else if targetTemplate.GetTemplateType() == template.PodChaos {
			chaosActor = actor.NewDeletePodChaosActor(it.namespace, it.nodeStatus.GetName())
		} else {
			// TODO: replace this error
			return nil, fmt.Errorf("%s is not supported now", targetTemplate.GetTemplateType())
		}

		var result []sideeffect.SideEffect
		result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.Running))
		result = append(result, sideeffect.NewCreateActorEventSideEffect(chaosActor))
		result = append(result, sideeffect.NewNotifyNewEventSideEffect(trigger.NewEvent(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), trigger.NodeChaosCleaned)))
		return result, nil
	case trigger.NodeChaosCleaned:
		// TODO: assert current state
		var result []sideeffect.SideEffect
		result = append(result, sideeffect.NewUpdatePhaseStatusSideEffect(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetName(), it.nodeStatus.GetNodePhase(), node.Succeed))
		result = append(result, sideeffect.NewNotifyNewEventSideEffect(trigger.NewEvent(it.namespace, it.workflowSpec.GetName(), it.nodeStatus.GetParentNodeName(), trigger.ChildNodeSucceed)))
		return result, nil
	default:
		// TODO: replace this error
		return nil, fmt.Errorf("StateMachine %s could not resolve event %s", it.GetName(), event)
	}
}
