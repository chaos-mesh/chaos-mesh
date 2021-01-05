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

package sideeffect

import (
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/actor"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

type SideEffect interface {
	GetSideEffectType() SideEffectType
}

type SideEffectType string

const (
	UpdateNodePhase SideEffectType = "UpdateNodePhase"
	CreateNewNode                  = "CreateNewNode"
	CreateActor                    = "CreateActor"
	NotifyNewEvent                 = "NotifyNewEvent"
)

type UpdateNodePhaseSideEffect struct {
	WorkflowName string
	NodeName     string
	CurrentPhase node.NodePhase
	TargetPhase  node.NodePhase
}

func NewUpdatePhaseStatusSideEffect(workflowName string, nodeName string, currentPhase node.NodePhase, targetPhase node.NodePhase) *UpdateNodePhaseSideEffect {
	return &UpdateNodePhaseSideEffect{WorkflowName: workflowName, NodeName: nodeName, CurrentPhase: currentPhase, TargetPhase: targetPhase}
}

func (it *UpdateNodePhaseSideEffect) GetSideEffectType() SideEffectType {
	return UpdateNodePhase
}

type CreateNewNodeSideEffect struct {
	WorkflowName   string
	ParentNodeName string
	NodeName       string
	TemplateName   string
	NodePhase      node.NodePhase
}

func NewCreateNewNodeSideEffect(workflowName string, parentNodeName string, nodeName string, templateName string, nodePhase node.NodePhase) *CreateNewNodeSideEffect {
	return &CreateNewNodeSideEffect{WorkflowName: workflowName, ParentNodeName: parentNodeName, NodeName: nodeName, TemplateName: templateName, NodePhase: nodePhase}
}

func (it *CreateNewNodeSideEffect) GetSideEffectType() SideEffectType {
	return CreateNewNode
}

type NotifyNewEventSideEffect struct {
	NewEvent trigger.Event
	Delay    time.Duration
}

func NewNotifyNewEventSideEffect(newEvent trigger.Event) *NotifyNewEventSideEffect {
	return &NotifyNewEventSideEffect{NewEvent: newEvent, Delay: 0}
}
func NewNotifyNewDelayEventSideEffect(newEvent trigger.Event, delay time.Duration) *NotifyNewEventSideEffect {
	return &NotifyNewEventSideEffect{NewEvent: newEvent, Delay: delay}
}

func (it *NotifyNewEventSideEffect) GetSideEffectType() SideEffectType {
	return NotifyNewEvent
}

type CreateActorEventSideEffect struct {
	actor actor.Actor
}

func (it *CreateActorEventSideEffect) GetActor() actor.Actor {
	return it.actor
}

func (it *CreateActorEventSideEffect) GetSideEffectType() SideEffectType {
	return CreateActor
}

func NewCreateActorEventSideEffect(actor actor.Actor) *CreateActorEventSideEffect {
	return &CreateActorEventSideEffect{actor: actor}
}
