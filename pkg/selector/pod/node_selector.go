package pod

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type nodeSelector struct {
	nodes []v1.Node
}

var _ generic.Selector = &nodeSelector{}

func (s *nodeSelector) AddListOption(opts client.ListOptions) client.ListOptions {
	return opts
}

func (s *nodeSelector) SetListFunc(f generic.ListFunc) generic.ListFunc {
	return f
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
func ParseNodeSelector(ctx context.Context, c client.Client, nodeNames []string, selectors map[string]string) (generic.Selector, error) {
	if len(nodeNames) == 0 && len(selectors) == 0 {
		return nil, nil
	}
	var nodes []v1.Node
	if len(nodeNames) > 0 {
		for _, name := range nodeNames {
			var node v1.Node
			if err := c.Get(ctx, types.NamespacedName{Name: name}, &node); err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
	}
	if len(selectors) > 0 {
		var nodeList v1.NodeList
		if err := c.List(ctx, &nodeList, &client.ListOptions{
			LabelSelector: labels.SelectorFromSet(selectors),
		}); err != nil {
			return nil, err
		}
		nodes = append(nodes, nodeList.Items...)
	}
	return &nodeSelector{nodes: nodes}, nil
}
