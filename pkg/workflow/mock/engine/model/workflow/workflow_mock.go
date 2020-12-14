package workflow

import (
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
)

type mockWorkflowSpec struct {
	name      string
	entry     string
	templates template.Templates
}

func NewMockWorkflowSpec() *mockWorkflowSpec {
	return &mockWorkflowSpec{}
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

func (it *mockWorkflowSpec) GetName() string {
	return it.name
}

func (it *mockWorkflowSpec) GetEntry() string {
	return it.entry
}

func (it *mockWorkflowSpec) GetTemplates() template.Templates {
	return it.templates
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
