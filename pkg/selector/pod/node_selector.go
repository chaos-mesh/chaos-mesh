// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package pod

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

const nodeSelectorName = "node"

type nodeSelector struct {
	nodes []v1.Node
	// empty avoids the situation that v1alpha1.PodSelectorSpec.NodeSelectors does not match any nodes, which should not skip the selector check
	// empty is true only when v1alpha1.PodSelectorSpec.Nodes and v1alpha1.PodSelectorSpec.NodeSelectors are both empty
	empty bool
}

var _ generic.Selector = &nodeSelector{}

func (s *nodeSelector) ListOption() client.ListOption {
	return nil
}

func (s *nodeSelector) ListFunc(_ client.Reader) generic.ListFunc {
	return nil
}

func (s *nodeSelector) Match(obj client.Object) bool {
	if s.empty {
		return true
	}

	pod := obj.(*v1.Pod)
	for _, node := range s.nodes {
		if node.Name == pod.Spec.NodeName {
			return true
		}
	}
	return false
}

// if both setting Nodes and NodeSelectors, the node list will be combined.
func newNodeSelector(ctx context.Context, c client.Client, spec v1alpha1.PodSelectorSpec) (generic.Selector, error) {
	if len(spec.Nodes) == 0 && len(spec.NodeSelectors) == 0 {
		return &nodeSelector{empty: true}, nil
	}
	var nodes []v1.Node
	if len(spec.Nodes) > 0 {
		for _, name := range spec.Nodes {
			var node v1.Node
			if err := c.Get(ctx, types.NamespacedName{Name: name}, &node); err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
	}
	if len(spec.NodeSelectors) > 0 {
		var nodeList v1.NodeList
		if err := c.List(ctx, &nodeList, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(spec.NodeSelectors),
		}); err != nil {
			return nil, err
		}
		nodes = append(nodes, nodeList.Items...)
	}
	return &nodeSelector{nodes: nodes}, nil
}
