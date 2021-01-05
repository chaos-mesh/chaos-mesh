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
	"context"
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/manager/sideeffect/resolver"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/workflowrepo"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/manager/sideeffect"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/manager/statemachine"

	"github.com/go-logr/logr"

	"go.uber.org/atomic"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

type basicManager struct {
	name   string
	repo   workflowrepo.WorkflowRepo
	logger logr.Logger

	// A composite trigger.
	multiplexTrigger  trigger.Trigger
	operableTrigger   trigger.OperableTrigger
	nodeNameGenerator node.NodeNameGenerator

	sideEffectsResolver resolver.SideEffectsResolver
}

func NewBasicManager(
	name string,
	repo workflowrepo.WorkflowRepo,
	logger logr.Logger,
	nodeNameGenerator node.NodeNameGenerator,
	sideEffectsResolver resolver.SideEffectsResolver,
	triggers ...trigger.Trigger,
) *basicManager {
	operableTrigger := trigger.NewOperableTrigger()
	allTriggers := append(triggers, operableTrigger)
	composite := trigger.NewCompositeTrigger(allTriggers...)
	return &basicManager{
		name:                name,
		repo:                repo,
		logger:              logger,
		nodeNameGenerator:   nodeNameGenerator,
		multiplexTrigger:    composite,
		operableTrigger:     operableTrigger,
		sideEffectsResolver: sideEffectsResolver,
	}
}

// TODO: constructor for basicManager

func (it *basicManager) GetName() string {
	return it.name
}

func (it *basicManager) Run(ctx context.Context) {
	working := atomic.NewBool(true)
	go func() {
		<-ctx.Done()
		working.Store(true)
	}()

	for working.Load() {
		event, err := it.acquire(ctx)
		if err != nil {
			it.logger.Error(err, "Failed to acquire new event", "manager-name", it.GetName())
			continue
		}

		// TODO: consuming in parallel
		err = it.consume(ctx, event)
		if err != nil {
			it.logger.Error(err, "Failed to consume event", "manager-name", it.GetName(), "event", event)
			continue
		}
	}
}

func (it *basicManager) acquire(ctx context.Context) (trigger.Event, error) {
	withCancel, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	event, _, err := it.multiplexTrigger.Acquire(withCancel)
	return event, err
}

func (it *basicManager) consume(ctx context.Context, event trigger.Event) error {

	if event.GetEventType() == trigger.ChildNodeSucceed && event.GetNodeName() == "" {
		return it.operableTrigger.Notify(trigger.NewEvent(event.GetNamespace(), event.GetWorkflowName(), "", trigger.WorkflowFinished))
	}

	switch event.GetEventType() {
	case trigger.WorkflowCreated:
		it.logger.V(1).Info("event: workflow created", "event", event)
		workflowName := event.GetWorkflowName()
		workflow, _, err := it.repo.FetchWorkflow(event.GetNamespace(), workflowName)
		if err != nil {
			return err
		}
		nodeName := it.nodeNameGenerator.GenerateNodeName(workflow.GetEntry())
		err = it.repo.CreateNodes(event.GetNamespace(), workflowName, "", nodeName, workflow.GetEntry())
		if err != nil {
			return err
		}
		err = it.operableTrigger.Notify(trigger.NewEvent(event.GetNamespace(), workflowName, nodeName, trigger.NodeCreated))
		if err != nil {
			return err
		}
		return nil

	case trigger.WorkflowFinished:
		// NOOP
		return nil

	case trigger.NodeCreated, trigger.NodeFinished, trigger.NodeHoldingAwake, trigger.NodePickChildToSchedule,
		trigger.NodeChaosInjected, trigger.NodeChaosCleaned,
		trigger.NodeUnexpectedFailed,
		trigger.ChildNodeSucceed, trigger.ChildNodeFailed:

		workflowName := event.GetWorkflowName()
		nodeName := event.GetNodeName()
		workflowSpec, workflowStatus, err := it.repo.FetchWorkflow(event.GetNamespace(), workflowName)
		if err != nil {
			return err
		}
		nodeStatus, err := workflowStatus.FetchNodeByName(nodeName)
		if err != nil {
			return err
		}
		targetTemplate, err := workflowSpec.FetchTemplateByName(nodeStatus.GetTemplateName())
		if err != nil {
			return err
		}

		// create state machine and make side effects
		var sideEffects []sideeffect.SideEffect

		switch targetTemplate.GetTemplateType() {
		case template.Serial:
			treeNode := workflowStatus.GetNodesTree().FetchNodeByName(nodeName)
			serialStateMachine := statemachine.NewSerialStateMachine(event.GetNamespace(), workflowSpec, nodeStatus, treeNode, it.nodeNameGenerator)
			sideEffects, err = serialStateMachine.HandleEvent(event)
			if err != nil {
				return err
			}
		case template.Suspend:
			suspendStateMachine := statemachine.NewSuspendStateMachine(event.GetNamespace(), workflowSpec, nodeStatus)
			sideEffects, err = suspendStateMachine.HandleEvent(event)
			if err != nil {
				return err
			}
		case template.NetworkChaos:
			chaosStateMachine := statemachine.NewChaosStateMachine(event.GetNamespace(), workflowSpec, nodeStatus)
			sideEffects, err = chaosStateMachine.HandleEvent(event)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported template %s", targetTemplate.GetTemplateType())
		}

		// TODO: apply side effects
		for _, sideEffectItem := range sideEffects {
			err := it.sideEffectsResolver.ResolveSideEffect(sideEffectItem)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		// TODO: replace this error
		return fmt.Errorf("unsupported event type: %s", event.GetEventType())
	}
}
