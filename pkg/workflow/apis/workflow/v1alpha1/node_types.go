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

import "github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"

type Node struct {
	Name         string         `json:"name"`
	// TODO: construct node tree with only children
	ParentNode   string         `json:"parent_node"`
	NodePhase    node.NodePhase `json:"node_phase"`
	TemplateName string         `json:"template_name"`
	// Children is necessary for representing the order when replicated child template references by parent template.
	// +optional
	Children []string `json:"children,omitempty"`
}
