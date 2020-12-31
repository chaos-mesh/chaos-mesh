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

package trigger

type Event interface {
	GetWorkflowName() string
	GetNodeName() string
	GetEventType() EventType
}

type EventType string

const (
	WorkflowCreated  EventType = "WorkflowCreated"
	WorkflowFinished EventType = "WorkflowFinished"

	NodeCreated             EventType = "NodeCreated"
	NodeFinished            EventType = "NodeFinished"
	NodeHoldingAwake        EventType = "NodeHoldingAwake"
	NodePickChildToSchedule EventType = "NodePickChildToSchedule"

	NodeChaosInjectSucceed EventType = "NodeChaosInjectSucceed"
	NodeChaosInjectFailed  EventType = "NodeChaosInjectFailed"

	// TODO: error handling
	NodeUnexpectedFailed EventType = "NodeUnexpectedFailed"

	ChildNodeSucceed EventType = "ChildNodeSucceed"
	ChildNodeFailed  EventType = "ChildNodeFailed"

	// TODO: support abort
	NodeAbort EventType = "NodeAbort"
)

type basicEvent struct {
	workflowName string
	nodeName     string
	eventType    EventType
}

func (it *basicEvent) GetWorkflowName() string {
	return it.workflowName
}

func (it *basicEvent) GetNodeName() string {
	return it.nodeName
}

func (it *basicEvent) GetEventType() EventType {
	return it.eventType
}

func NewEvent(workflowName string, nodeName string, eventType EventType) Event {
	return &basicEvent{
		workflowName: workflowName,
		nodeName:     nodeName,
		eventType:    eventType,
	}
}
