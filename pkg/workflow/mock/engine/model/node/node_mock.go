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

func NewMockNode(nodeName string, phase node.NodePhase, parentNodeName string, templateName string, templateType template.TemplateType) *mockNode {
	return &mockNode{nodeName: nodeName, phase: phase, parentNodeName: parentNodeName, templateName: templateName, templateType: templateType}
}

func (it *mockNode) SetPhase(phase node.NodePhase) {
	it.phase = phase
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

type mockTreeNode struct {
	nodeName     string
	templateName string
	children     *mockNodeTreeChildren
}

func NewMockTreeNode(nodeName string, templateName string, children *mockNodeTreeChildren) *mockTreeNode {
	return &mockTreeNode{nodeName: nodeName, templateName: templateName, children: children}
}

func (it *mockTreeNode) SetChildren(children *mockNodeTreeChildren) {
	it.children = children
}

func (it *mockTreeNode) GetName() string {
	return it.nodeName
}

func (it *mockTreeNode) FetchNodeByName(nodeName string) node.NodeTreeNode {
	if it.GetName() == nodeName {
		return it
	}
	for _, treeNode := range it.GetChildren().GetAllChildrenNode() {
		target := treeNode.FetchNodeByName(nodeName)
		if target != nil {
			return target
		}
	}
	return nil
}

func (it *mockTreeNode) GetTemplateName() string {
	return it.templateName
}

func (it *mockTreeNode) GetChildren() node.NodeTreeChildren {
	return it.children
}

type mockNodeTreeChildren struct {
	nodesMap map[string]node.NodeTreeNode
}

func NewMockNodeTreeChildren(nodesMap map[string]node.NodeTreeNode) *mockNodeTreeChildren {
	if nodesMap == nil {
		nodesMap = make(map[string]node.NodeTreeNode)
	}
	return &mockNodeTreeChildren{nodesMap: nodesMap}
}

func (it *mockNodeTreeChildren) Length() int {
	return len(it.nodesMap)
}

func (it *mockNodeTreeChildren) ContainsNode(nodeName string) bool {
	_, exists := it.nodesMap[nodeName]
	return exists
}

func (it *mockNodeTreeChildren) ContainsTemplate(templateName string) bool {
	for _, treeNode := range it.nodesMap {
		if treeNode.GetTemplateName() == templateName {
			return true
		}
	}
	return false
}

func (it *mockNodeTreeChildren) GetAllChildrenNode() []node.NodeTreeNode {
	result := make([]node.NodeTreeNode, 0)
	for _, treeNode := range it.nodesMap {
		result = append(result, treeNode)
	}
	return result
}
