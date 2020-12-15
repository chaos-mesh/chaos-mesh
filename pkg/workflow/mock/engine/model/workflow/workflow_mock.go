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
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
	mocktemplate "github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/template"
)

type mockWorkflowSpec struct {
	name      string
	entry     string
	templates template.Templates
}

func NewMockWorkflowSpec() *mockWorkflowSpec {
	return &mockWorkflowSpec{
		templates: mocktemplate.NewMockedTemplates(nil),
	}
}

func (it *mockWorkflowSpec) SetName(name string) {
	it.name = name
}

func (it *mockWorkflowSpec) SetEntry(entry string) {
	it.entry = entry
}

func (it *mockWorkflowSpec) SetTemplates(templates template.Templates) {
	it.templates = templates
}

func (it *mockWorkflowSpec) GetTemplates() (template.Templates, error) {
	if it.templates == nil {
		return nil, errors.NewTemplatesIsRequiredError("mockWorkflowSpec.GetTemplates", it.GetName())
	}
	return it.templates, nil
}

func (it *mockWorkflowSpec) GetName() string {
	return it.name
}

func (it *mockWorkflowSpec) GetEntry() string {
	return it.entry
}

type mockWorkflowStatus struct {
	workflowSpecName string
	nodes            []node.Node
	rootNode         node.NodeTreeNode
	phase            workflow.WorkflowPhase
}

func NewMockWorkflowStatus() *mockWorkflowStatus {
	return &mockWorkflowStatus{}
}

func (it *mockWorkflowStatus) SetWorkflowSpecName(workflowSpecName string) {
	it.workflowSpecName = workflowSpecName
}

func (it *mockWorkflowStatus) SetNodes(nodes []node.Node) {
	it.nodes = nodes
}

func (it *mockWorkflowStatus) SetRootNode(rootNode node.NodeTreeNode) {
	it.rootNode = rootNode
}

func (it *mockWorkflowStatus) SetPhase(phase workflow.WorkflowPhase) {
	it.phase = phase
}

func (it *mockWorkflowStatus) GetPhase() workflow.WorkflowPhase {
	return it.phase
}

func (it *mockWorkflowStatus) GetNodes() []node.Node {
	return it.nodes
}

func (it *mockWorkflowStatus) GetWorkflowSpecName() string {
	return it.workflowSpecName
}

func (it *mockWorkflowStatus) GetNodesTree() node.NodeTreeNode {
	return it.rootNode
}

func (it *mockWorkflowStatus) FetchNodesMap() map[string]node.Node {
	result := make(map[string]node.Node)
	for _, item := range it.nodes {
		if _, exists := result[item.GetName()]; exists {
			panic(fmt.Sprintf("node name %s already exist,", item.GetName()))
		} else {
			result[item.GetName()] = item
		}
	}
	return result
}
