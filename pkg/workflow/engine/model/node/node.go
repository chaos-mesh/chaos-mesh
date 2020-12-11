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

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
)

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
	GetName() string
	GetNodePhase() NodePhase
	GetParent() string

	GetTemplateName() string
	GetTemplateType() template.TemplateType
}

type NodeTreeNode interface {
	GetName() string
	GetChildren() NodeTreeChildren
	GetParent() NodeTreeNode
	FetchChildNodeByNodeName(nodeName string) NodeTreeNode
}

type NodeTreeChildren interface {
	Length() int
	ContainsName(nodeName string) bool
	ContainsTemplate(templateName string) bool
	GetAllChildrenNode() []NodeTreeNode
}
