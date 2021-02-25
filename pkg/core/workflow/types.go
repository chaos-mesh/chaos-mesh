// Copyright 2021 Chaos Mesh Authors.
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

// Package workflow includes all type definitaions about Workflow.
package workflow

// Workflow defines the root structure of a workflow.
type Workflow struct {
	UID         string         `gorm:"index:uid" json:"uid"`
	Name        string         `json:"name"`
	Entry       string         `json:"entry"` // the entry node name
	Status      workflowStatus `json:"status"`
	CurrentNode TopologyNode   `json:"current_node"`
	Topology    Topology       `json:"topology"`
}

// workflowState defines the current state of a workflow.
//
// Includes: Initializing, Running, Errored, Finished.
//
// Const definitions can be found below this type.
type workflowStatus string

const (
	// WorkflowInitializing represents a workflow is being initialized.
	WorkflowInitializing workflowStatus = "Initializing"

	// WorkflowRunning represents that a workflow is running.
	WorkflowRunning workflowStatus = "Running"

	// WorkflowErrored represents an error in a workflow.
	WorkflowErrored workflowStatus = "Errored"

	// WorkflowFinished represents that a workflow has ended.
	WorkflowFinished workflowStatus = "Finished"
)

// Topology describes the process of a workflow.
type Topology struct {
	Nodes []TopologyNode `json:"nodes"`
}

// TopologyNode defines the basic structure of a node.
type TopologyNode struct {
	Type  nodeType  `json:"type"`
	State nodeState `json:"state,omitempty"`
}

// Node defines the single step of a workflow.
type Node struct {
	TopologyNode
	Serial   NodeSerial   `json:"serial,omitempty"`
	Parallel NodeParallel `json:"parallel,omitempty"`
	Template Template     `json:"template"`
}

// NodeSerial defines SerialNode's specific fields.
type NodeSerial struct {
	Tasks []Node `json:"tasks"`
}

// NodeParallel defines ParallelNode's specific fields.
type NodeParallel struct {
	Tasks []Node `json:"tasks"`
}

// nodeType defines the type of a workflow node.
//
// There will be five types can be refered as nodeType: Chaos, Serial, Parallel, Suspend, Task.
//
// Const definitions can be found below this type.
type nodeType string

const (
	// ChaosNode represents a node will perform a single Chaos Experiment.
	ChaosNode nodeType = "ChaosNode"

	// SerialNode represents a node that will perform continuous templates.
	SerialNode nodeType = "SerialNode"

	// ParallelNode represents a node that will perform parallel templates.
	ParallelNode nodeType = "ParallelNode"

	// SuspendNode represents a node that will perform wait operation.
	SuspendNode nodeType = "SuspendNode"

	// TaskNode represents a node that will perform user-defined task.
	TaskNode nodeType = "TaskNode"
)

// nodeState represents a node in different stage.
//
// It should be note that not all states are applicable to any node types.
// A Node will contains only partial defined states.
//
// Const definitions can be found below this type.
type nodeState string

const (
	// NodeInitializing represents a node is being initialized.
	NodeInitializing nodeState = "Initializing"

	// NodeWaitingForSchedule represents a node is idle and safe for next scheduling.
	//
	// Only available in: SerialNode, ParallelNode and TaskNode.
	NodeWaitingForSchedule nodeState = "WaitingForSchedule"

	// NodeWaitingForChild represents at least 1 child node is in Running or Holding state.
	//
	// Only available in: SerialNode, ParallelNode and TaskNode.
	NodeWaitingForChild nodeState = "WaitingForChild"

	// NodeRunning represents a node is doing dirty works.
	//
	// Available in: ChaosNode, SuspendNode and TaskNode.
	NodeRunning nodeState = "Running"

	// NodeEvaluating represents a node is collecting the result of a user's pod, then picks a template to execute.
	//
	// Only available: TaskNode.
	NodeEvaluating nodeState = "Evaluating"

	// NodeHolding represents that the current node is waiting for the next action.
	//
	// Only available: ChaosNode and SuspendNode.
	NodeHolding nodeState = "Holding"

	// NodeSucceed represents a node is completed.
	NodeSucceed nodeState = "Succeed"

	// NodeFailed represents a node is failed.
	NodeFailed nodeState = "Failed"
)

// Detail defines the detail of a workflow.
type Detail struct {
	WorkflowUID string     `json:"workflow_uid"`
	Templates   []Template `json:"templates"`
}

// Template defines a complete structure of a template.
type Template struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Duration string      `json:"duration,omitempty"`
	Spec     interface{} `json:"spec"`
}
