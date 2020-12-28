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
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
)

func (it *Workflow) GetEntry() string {
	return it.Spec.Entry
}

func (it *Workflow) FetchTemplateByName(templateName string) (template.Template, error) {
	panic("implement me")
}

func (it *Workflow) GetPhase() workflow.WorkflowPhase {
	return it.Status.Phase
}

func (it *Workflow) GetNodes() []node.Node {
	panic("implement me")
}

func (it *Workflow) GetWorkflowSpecName() string {
	return it.ObjectMeta.Name
}

func (it *Workflow) GetNodesTree() node.NodeTreeNode {
	panic("implement me")
}

func (it *Workflow) FetchNodesMap() map[string]node.Node {
	panic("implement me")
}

func (it *Workflow) FetchNodeByName(nodeName string) (node.Node, error) {
	panic("implement me")
}
