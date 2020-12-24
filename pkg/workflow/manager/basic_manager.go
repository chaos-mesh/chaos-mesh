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

	"go.uber.org/atomic"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/scheduler"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

type basicManager struct {
	name string
	repo WorkflowRepo
	// A composite trigger.
	multiplexTrigger trigger.Trigger

	operableTrigger trigger.OperableTrigger
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
			// TODO: warn log
			continue
		}
		err = it.consume(ctx, event)
		if err != nil {
			// TODO: warn log
			continue
		}
	}
}

func (it *basicManager) acquire(ctx context.Context) (trigger.Event, error) {
	return it.multiplexTrigger.Acquire(ctx)
}

func (it *basicManager) consume(ctx context.Context, event trigger.Event) error {
	switch event.GetEventType() {
	case trigger.WorkflowCreated:
		workflowName := event.GetWorkflowName()
		workflow, status, err := it.repo.FetchWorkflow(workflowName)
		if err != nil {
			return err
		}
		basicScheduler := scheduler.NewBasicScheduler(workflow, status)
		templatesToRun, parentNode, err := basicScheduler.ScheduleNext(ctx)
		nodeNames, err := it.repo.CreateNodes(workflowName, templatesToRun, parentNode)
		for _, nodeName := range nodeNames {
			err := it.operableTrigger.Notify(trigger.NewEvent(workflowName, nodeName, trigger.NodeCreated))
			if err != nil {
				// TODO: warn log
				return err
			}
		}
	case trigger.WorkflowFinished:

	case trigger.NodeCreated:
		workflowName := event.GetWorkflowName()
		nodeName := event.GetNodeName()
		workflowSpec, workflowStatus, err := it.repo.FetchWorkflow(workflowName)
		if err != nil {
			return err
		}

		node, err := workflowStatus.FetchNodeByName(nodeName)
		if err != nil {
			return err
		}
		targetTemplate, err := workflowSpec.FetchTemplateByName(node.GetTemplateName())
		if err != nil {
			return err
		}

		switch targetTemplate.GetTemplateType() {
		case template.Serial:
			err := it.repo.UpdateNodesToWaitingForSchedule(workflowName, nodeName)
			if err != nil {
				return nil
			}
		default:
			return fmt.Errorf("unsupproted template type %s", targetTemplate.GetTemplateType())
		}

		err = it.repo.UpdateNodesToRunning(workflowName, nodeName)
		if err != nil {
			return err
		}

	case trigger.NodeFinished:

	case trigger.NodeWaitingForSchedule:
		workflowName := event.GetWorkflowName()
		nodeName := event.GetNodeName()
		workflowSpec, workflowStatus, err := it.repo.FetchWorkflow(workflowName)
		if err != nil {
			return err
		}
		node, err := workflowStatus.FetchNodeByName(nodeName)
		if err != nil {
			return err
		}
		targetTemplate, err := workflowSpec.FetchTemplateByName(node.GetTemplateName())
		if err != nil {
			return err
		}
		switch targetTemplate.GetTemplateType() {
		case template.Serial:
			basicScheduler := scheduler.NewBasicScheduler(workflowSpec, workflowStatus)
			newTemplateNeedSchedule, err := basicScheduler.ScheduleNextWithinParent(ctx, nodeName)
			err = it.repo.UpdateNodesToWaitingForChild(workflowName, nodeName)
			if err != nil {
				return nil
			}
			newNodeNames, err := it.repo.CreateNodes(workflowName, newTemplateNeedSchedule, nodeName)
			if err != nil {
				return nil
			}
			for _, newNodeName := range newNodeNames {
				err := it.operableTrigger.Notify(trigger.NewEvent(workflowName, newNodeName, trigger.NodeCreated))
				if err != nil {
					// TODO: warn log
					return err
				}
			}
		default:
			return fmt.Errorf("unsupproted template type %s", targetTemplate.GetTemplateType())
		}

	default:
		return fmt.Errorf("unsupported event type: %s", event.GetEventType())
	}

	// default logic
	return nil
}
