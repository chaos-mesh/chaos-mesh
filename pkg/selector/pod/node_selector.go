package pod

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const nodeSelectorName = "node"

type nodeSelector struct {
	nodes []v1.Node
}

var _ generic.Selector = &nodeSelector{}


func (s *nodeSelector) ListOption() client.ListOption {
	return nil
}

func (s *nodeSelector) ListFunc() generic.ListFunc {
	return nil
}

func (s *nodeSelector) Match(obj client.Object) bool {
	if len(s.nodes) == 0 {
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
		return &nodeSelector{}, nil
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
