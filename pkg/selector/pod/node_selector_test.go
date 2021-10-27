package pod

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"testing"

	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func TestFilterPodByNode(t *testing.T) {
	g := NewGomegaWithT(t)

	pods := []v1.Pod{
		NewPod(PodArg{Name: "p1", Namespace: "n1", Nodename: "node1"}),
		NewPod(PodArg{Name: "p2", Namespace: "n2", Nodename: "node1"}),
		NewPod(PodArg{Name: "p3", Namespace: "n2", Nodename: "node2"}),
		NewPod(PodArg{Name: "p4", Namespace: "n4", Nodename: "node3"}),
	}

	nodes := []v1.Node{
		NewNode("node1", map[string]string{"disktype": "ssd", "zone": "az1"}),
		NewNode("node2", map[string]string{"disktype": "hdd", "zone": "az1"}),
	}

	node1AndNode2Selector := &nodeSelector{nodes: nodes}
	emptyNodeSelector := &nodeSelector{empty: true}
	noNodeSelector := &nodeSelector{}

	tcs := []struct {
		name     string
		pod      v1.Pod
		selector generic.Selector
		match    bool
	}{
		{
			name:     "filter pods from node1 and node2",
			pod:      pods[0],
			selector: node1AndNode2Selector,
			match:    true,
		}, {
			name:     "filter pods from node1 and node2",
			pod:      pods[2],
			selector: node1AndNode2Selector,
			match:    true,
		}, {
			name:     "filter pods from node1 and node2",
			pod:      pods[3],
			selector: node1AndNode2Selector,
			match:    false,
		}, {
			name:     "empty filter",
			pod:      pods[0],
			selector: emptyNodeSelector,
			match:    true,
		}, {
			name:     "filter no nodes",
			pod:      pods[0],
			selector: noNodeSelector,
			match:    false,
		},
	}

	for _, tc := range tcs {
		g.Expect(tc.selector.Match(&tc.pod)).To(Equal(tc.match), tc.name)
	}
}
