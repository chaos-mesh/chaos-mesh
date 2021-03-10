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

package core

type WorkflowRepository interface {
	MutateWithKubeClient() WorkflowRepository
}

// Workflow defines the root structure of a workflow.
type Workflow struct {
	UID          string         `json:"uid"`
	Name         string         `json:"name"`
	Entry        string         `json:"entry"` // the entry node name
	Status       WorkflowStatus `json:"status"`
	Topology     Topology       `json:"topology"`
	CurrentNodes []Node         `json:"current_nodes"`
}

// workflowState defines the current state of a workflow.
//
// Includes: Initializing, Running, Errored, Finished.
//
// Const definitions can be found below this type.
type WorkflowStatus string

const (
	// WorkflowInitializing represents a workflow is being initialized.
	WorkflowInitializing WorkflowStatus = "Initializing"

	// WorkflowRunning represents that a workflow is running.
	WorkflowRunning WorkflowStatus = "Running"

	// WorkflowErrored represents an error in a workflow.
	WorkflowErrored WorkflowStatus = "Errored"

	// WorkflowFinished represents that a workflow has ended.
	WorkflowFinished WorkflowStatus = "Finished"
)

// Topology describes the process of a workflow.
type Topology struct {
	Nodes []Node `json:"nodes"`
}

// Node defines the single step of a workflow.
type Node struct {
	Name     string       `json:"name"`
	Type     NodeType     `json:"type"`
	State    NodeState    `json:"state,omitempty"`
	Serial   NodeSerial   `json:"serial,omitempty"`
	Parallel NodeParallel `json:"parallel,omitempty"`
	Template string       `json:"template"`
}

// NodeSerial defines SerialNode's specific fields.
type NodeSerial struct {
	Tasks []string `json:"tasks"`
}

// NodeParallel defines ParallelNode's specific fields.
type NodeParallel struct {
	Tasks []string `json:"tasks"`
}

// NodeType defines the type of a workflow node.
//
// There will be five types can be refered as NodeType: Chaos, Serial, Parallel, Suspend, Task.
//
// Const definitions can be found below this type.
type NodeType string

const (
	// ChaosNode represents a node will perform a single Chaos Experiment.
	ChaosNode NodeType = "ChaosNode"

	// SerialNode represents a node that will perform continuous templates.
	SerialNode NodeType = "SerialNode"

	// ParallelNode represents a node that will perform parallel templates.
	ParallelNode NodeType = "ParallelNode"

	// SuspendNode represents a node that will perform wait operation.
	SuspendNode NodeType = "SuspendNode"

	// TaskNode represents a node that will perform user-defined task.
	TaskNode NodeType = "TaskNode"
)

// NodeState represents a node in different stage.
//
// It should be note that not all states are applicable to any node types.
// A Node will contains only partial defined states.
//
// Const definitions can be found below this type.
type NodeState string

const (
	// NodeInitializing represents a node is being initialized.
	NodeInitializing NodeState = "Initializing"

	// NodeWaitingForSchedule represents a node is idle and safe for next scheduling.
	//
	// Only available in: SerialNode, ParallelNode and TaskNode.
	NodeWaitingForSchedule NodeState = "WaitingForSchedule"

	// NodeWaitingForChild represents at least 1 child node is in Running or Holding state.
	//
	// Only available in: SerialNode, ParallelNode and TaskNode.
	NodeWaitingForChild NodeState = "WaitingForChild"

	// NodeRunning represents a node is doing dirty works.
	//
	// Available in: ChaosNode, SuspendNode and TaskNode.
	NodeRunning NodeState = "Running"

	// NodeEvaluating represents a node is collecting the result of a user's pod, then picks a template to execute.
	//
	// Only available: TaskNode.
	NodeEvaluating NodeState = "Evaluating"

	// NodeHolding represents that the current node is waiting for the next action.
	//
	// Only available: ChaosNode and SuspendNode.
	NodeHolding NodeState = "Holding"

	// NodeSucceed represents a node is completed.
	NodeSucceed NodeState = "Succeed"

	// NodeFailed represents a node is failed.
	NodeFailed NodeState = "Failed"
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
