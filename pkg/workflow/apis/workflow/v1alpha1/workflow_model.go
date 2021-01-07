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

package v1alpha1

import (
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
)

func (it *Workflow) GetEntry() string {
	return it.Spec.Entry
}

func (it *Workflow) FetchTemplateByName(templateName string) (template.Template, error) {
	for _, item := range it.Spec.Templates {
		if item.Name == templateName {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("workflow %s does not contains such template called %s", it.Name, templateName)
}

func (it *Workflow) GetPhase() workflow.WorkflowPhase {
	return it.Status.Phase
}

func (it *Workflow) GetNodes() []node.Node {
	var result []node.Node
	for _, item := range it.Status.Nodes {
		item := item
		result = append(result, &item)
	}
	return result
}

func (it *Workflow) GetWorkflowSpecName() string {
	return it.ObjectMeta.Name
}

func (it *Workflow) GetNodesTree() (node.NodeTreeNode, error) {
	return buildTree(*it.Status.EntryNode, it.Status.Nodes)
}

func (it *Workflow) FetchNodesMap() map[string]node.Node {
	result := make(map[string]node.Node)
	for k, v := range it.Status.Nodes {
		result[k] = &v
	}
	return result
}

func (it *Workflow) FetchNodeByName(nodeName string) (node.Node, error) {
	if item, ok := it.Status.Nodes[nodeName]; ok {
		return &item, nil
	}
	return nil, fmt.Errorf("workflow %s does not contains such node called %s", it.Name, nodeName)
}

func buildTree(root string, nodesMap map[string]Node) (node.NodeTreeNode, error) {
	rootNode, ok := nodesMap[root]
	if !ok {
		return nil, fmt.Errorf("could fetch root node called %s", root)
	}

	rootTreeNode := treeNode{
		Node:     rootNode,
		children: nil,
	}

	var childrenNodeNames []string
	for _, item := range nodesMap {
		if item.GetParentNodeName() == root {
			childrenNodeNames = append(childrenNodeNames, item.GetName())
		}
	}

	var childrenTreeNodes children
	for _, item := range childrenNodeNames {
		childTreeNode, err := buildTree(item, nodesMap)
		if err != nil {
			return nil, err
		}
		childrenTreeNodes = append(childrenTreeNodes, childTreeNode)
	}
	rootTreeNode.children = &childrenTreeNodes

	return &rootTreeNode, nil
}

type treeNode struct {
	Node
	children *children
}

func (it *treeNode) GetChildren() node.NodeTreeChildren {
	return it.children
}

func (it *treeNode) FetchNodeByName(nodeName string) (node.NodeTreeNode, error) {
	if it.Name == nodeName {
		return it, nil
	}
	for _, child := range it.children.GetAllChildrenNode() {
		resultFromChild, err := child.FetchNodeByName(nodeName)
		if err == nil && resultFromChild != nil {
			return resultFromChild, nil
		}
	}
	return nil, fmt.Errorf("no such tree node called %s", nodeName)
}

type children []node.NodeTreeNode

func (it *children) Length() int {
	return len(*it)
}

func (it *children) ContainsTemplate(templateName string) bool {
	for _, item := range *it {
		if item.GetTemplateName() == templateName {
			return true
		}
	}
	return false
}

func (it *children) GetAllChildrenNode() []node.NodeTreeNode {
	return *it
}
