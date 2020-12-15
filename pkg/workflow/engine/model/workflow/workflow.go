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
	GetName() string
	GetEntry() string
	GetTemplates() (template.Templates, error)
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
	// func GetPhase returns current phase
	GetPhase() WorkflowPhase
	// func GetNodes returns the flat array of all nodes
	GetNodes() []node.Node
	// func GetWorkflowSpecName returns the name of WorkflowSpec
	GetWorkflowSpecName() string
	// func GetNodesTree returns the root of the node tree.
	// This tree could present the hierarchy of how nodes execute.
	GetNodesTree() node.NodeTreeNode

	// func FetchNodesMap
	// Key is the name of node
	FetchNodesMap() map[string]node.Node
}
