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

package workflow

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
)

type WorkflowSpec interface {
	Name() string
	Entry() string
	FetchTemplateByName(templateName string) (template.Template, error)
}

type WorkflowPhase string

const (
	// It's also the initial phase of a workflow.
	Init    WorkflowPhase = "Init"
	Running WorkflowPhase = "Running"
	Succeed WorkflowPhase = "Succeed"
	Failed  WorkflowPhase = "Failed"
)

type WorkflowStatus interface {
	// func Phase returns current phase
	Phase() WorkflowPhase
	// func Nodes returns the flat array of all nodes
	Nodes() []node.Node
	// func WorkflowSpecName returns the name of WorkflowSpec
	WorkflowSpecName() string
	// func NodesTree returns the root of the node tree.
	// This tree could present the hierarchy of how nodes execute.
	NodesTree() (node.NodeTreeNode, error)

	// func NodesMap
	// Key is the name of node
	NodesMap() map[string]node.Node

	FetchNodeByName(nodeName string) (node.Node, error)
}
