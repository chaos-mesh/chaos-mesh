package node

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
)

type mockNode struct {
	nodeName       string
	phase          node.NodePhase
	parentNodeName string
	templateName   string
	templateType   template.TemplateType
}

func NewMockNode() *mockNode {
	return &mockNode{}
}

func (it *mockNode) SetNodeName(nodeName string) {
	it.nodeName = nodeName
}

func (it *mockNode) SetPhase(phase node.NodePhase) {
	it.phase = phase
}

func (it *mockNode) SetParentNodeName(parentNodeName string) {
	it.parentNodeName = parentNodeName
}

func (it *mockNode) SetTemplateName(templateName string) {
	it.templateName = templateName
}

func (it *mockNode) SetTemplateType(templateType template.TemplateType) {
	it.templateType = templateType
}

func (it *mockNode) GetName() string {
	return it.nodeName
}

func (it *mockNode) GetNodePhase() node.NodePhase {
	return it.phase
}

func (it *mockNode) GetParentNodeName() string {
	return it.parentNodeName
}

func (it *mockNode) GetTemplateName() string {
	return it.templateName
}

func (it *mockNode) GetTemplateType() template.TemplateType {
	return it.templateType
}
