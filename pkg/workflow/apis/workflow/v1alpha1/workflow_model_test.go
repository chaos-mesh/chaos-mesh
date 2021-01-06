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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildTree(t *testing.T) {
	nodes := []Node{
		{
			Name:       "node0",
			ParentNode: "",
		}, {
			Name:       "node1-0",
			ParentNode: "node0",
		}, {
			Name:       "node1-1",
			ParentNode: "node0",
		}, {
			Name:       "node2-0",
			ParentNode: "node1-0",
		}, {
			Name:       "node2-1",
			ParentNode: "node1-0",
		}, {
			Name:       "node2-2",
			ParentNode: "node1-1",
		}, {
			Name:       "node2-3",
			ParentNode: "node1-1",
		},
	}

	nodesMap := make(map[string]Node)

	for _, item := range nodes {
		nodesMap[item.GetName()] = item
	}
	tree, err := buildTree("node0", nodesMap)
	assert.NoError(t, err)
	assert.Equal(t, "node0", tree.GetName())

	assert.Equal(t, 2, tree.GetChildren().Length())
	assert.Equal(t, true,
		("node1-0" == tree.GetChildren().GetAllChildrenNode()[0].GetName() && "node1-1" == tree.GetChildren().GetAllChildrenNode()[1].GetName()) ||
			("node1-1" == tree.GetChildren().GetAllChildrenNode()[0].GetName() && "node1-0" == tree.GetChildren().GetAllChildrenNode()[1].GetName()),
	)

	target, err := tree.FetchNodeByName("node0")
	assert.NoError(t, err)
	assert.NotNil(t, target)
	assert.Equal(t, "node0", target.GetName())

	target, err = tree.FetchNodeByName("node1-1")
	assert.NoError(t, err)
	assert.NotNil(t, target)
	assert.Equal(t, "node1-1", target.GetName())

	target, err = tree.FetchNodeByName("node2-3")
	assert.NoError(t, err)
	assert.NotNil(t, target)
	assert.Equal(t, "node2-3", target.GetName())

	target, err = tree.FetchNodeByName("not-exist")
	assert.Error(t, err)
}
