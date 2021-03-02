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

package node

type NodePhase string

const (
	Init               NodePhase = "Init"
	WaitingForSchedule NodePhase = "WaitingForSchedule"
	Running            NodePhase = "Running"
	// It means current node is not changing something, but waiting for some signal. It's still alive.
	// It's the most common state for most of Node which referenced Template contains duration or deadline.
	Holding         NodePhase = "Holding"
	Succeed         NodePhase = "Succeed"
	Failed          NodePhase = "Failed"
	WaitingForChild NodePhase = "WaitingForChild"
	Evaluating      NodePhase = "Evaluating"
)

type Node interface {
	Name() string
	NodePhase() NodePhase
	ParentNodeName() string
	TemplateName() string
}

// FIXME: remove this interface, it's should belongs to node status
type NodeTreeNode interface {
	Name() string
	TemplateName() string
	Children() NodeTreeChildren
	// Find node by name, might return itself, or find from children nodes recursively
	FetchNodeByName(nodeName string) (NodeTreeNode, error)
}

type NodeTreeChildren interface {
	Length() int
	GetAllChildrenNode() []NodeTreeNode
}
