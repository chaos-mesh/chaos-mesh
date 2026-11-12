// Copyright 2026 Chaos Mesh Authors.
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

package partition

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/podnetworkchaosmanager"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/nodescoped"
	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func TestLocalCidrs(t *testing.T) {
	g := NewWithT(t)

	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "node-a",
			Annotations: map[string]string{ciliumHostIPAnnotation: "10.0.1.1"},
		},
		Spec: v1.NodeSpec{PodCIDRs: []string{"10.0.1.0/24", "fd00::/64"}},
		Status: v1.NodeStatus{Addresses: []v1.NodeAddress{
			{Type: v1.NodeInternalIP, Address: "192.168.0.10"},
			{Type: v1.NodeInternalIP, Address: "fd00::10"},
			{Type: v1.NodeHostName, Address: "node-a"},
		}},
	}
	pods := []v1.Pod{
		{Status: v1.PodStatus{PodIP: "10.0.1.5", PodIPs: []v1.PodIP{{IP: "10.0.1.5"}, {IP: "fd00::5"}}}},
		{Status: v1.PodStatus{PodIP: "192.168.0.10"}},
		{Status: v1.PodStatus{}},
	}

	g.Expect(localCidrs(node, pods)).To(Equal([]string{
		"10.0.1.0/24",
		"10.0.1.1/32",
		"10.0.1.5/32",
		"192.168.0.10/32",
	}))
}

func newNodeScopeImpl(objects ...client.Object) (*Impl, client.Client) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		WithStatusSubresource(&v1alpha1.NodeNetworkChaos{}, &v1alpha1.PodNetworkChaos{}).
		WithIndex(&v1.Pod{}, nodescoped.NodeNameField, func(obj client.Object) []string {
			return []string{obj.(*v1.Pod).Spec.NodeName}
		}).
		Build()

	builder := podnetworkchaosmanager.NewBuilder(podnetworkchaosmanager.Params{
		Logger: logr.Discard(),
		Client: c,
		Reader: c,
		Scheme: scheme,
	})
	return NewImpl(c, builder, logr.Discard()), c
}

func TestApplyAndRecoverNodeScope(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-a"},
		Status: v1.NodeStatus{Addresses: []v1.NodeAddress{
			{Type: v1.NodeInternalIP, Address: "192.168.0.10"},
		}},
	}
	victim := NewPod(PodArg{Name: "crdb-0", Namespace: "db", Nodename: "node-a"})
	victim.Status.PodIP = "10.0.1.5"
	neighbour := NewPod(PodArg{Name: "web", Namespace: "apps", Nodename: "node-a"})
	neighbour.Status.PodIP = "10.0.1.6"

	impl, c := newNodeScopeImpl(node, &victim, &neighbour)

	networkchaos := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{Name: "gray", Namespace: "db"},
		Spec: v1alpha1.NetworkChaosSpec{
			Action:      v1alpha1.PartitionAction,
			Direction:   v1alpha1.Both,
			Partition:   &v1alpha1.PartitionSpec{DropPercent: "40"},
			Scope:       v1alpha1.NodeScope,
			IncludeHost: true,
		},
	}
	records := []*v1alpha1.Record{
		{Id: "db/crdb-0", SelectorKey: "."},
		{Id: v1alpha1.NodeRecordIDPrefix + "node-a", SelectorKey: "."},
	}

	localName := "gray_local"
	expectChains := func(chains []v1alpha1.RawIptables, device string) {
		g.Expect(chains).To(HaveLen(2))
		directions := []v1alpha1.ChainDirection{}
		for _, chain := range chains {
			directions = append(directions, chain.Direction)
			g.Expect(chain.DropPercent).To(Equal("40"))
			g.Expect(chain.ExcludeIPSets).To(Equal([]string{localName}))
			g.Expect(chain.IPSets).To(BeEmpty())
			g.Expect(chain.Device).To(Equal(device))
			g.Expect(chain.Source).To(Equal("db/gray"))
		}
		g.Expect(directions).To(ConsistOf(v1alpha1.Input, v1alpha1.Output))
	}
	expectLocal := func(ipsets []v1alpha1.RawIPSet) {
		g.Expect(ipsets).To(HaveLen(1))
		g.Expect(ipsets[0].Name).To(Equal(localName))
		g.Expect(ipsets[0].Cidrs).To(Equal([]string{"10.0.1.5/32", "10.0.1.6/32", "192.168.0.10/32"}))
	}

	phase, err := impl.Apply(ctx, 0, records, networkchaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(waitForApplySync))

	podnetworkchaos := &v1alpha1.PodNetworkChaos{}
	g.Expect(c.Get(ctx, types.NamespacedName{Namespace: "db", Name: "crdb-0"}, podnetworkchaos)).To(Succeed())
	expectLocal(podnetworkchaos.Spec.IPSets)
	expectChains(podnetworkchaos.Spec.Iptables, "")

	phase, err = impl.Apply(ctx, 1, records, networkchaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(waitForApplySync))

	nodenetworkchaos := &v1alpha1.NodeNetworkChaos{}
	g.Expect(c.Get(ctx, types.NamespacedName{Name: "node-a"}, nodenetworkchaos)).To(Succeed())
	g.Expect(nodenetworkchaos.OwnerReferences).To(HaveLen(1))
	g.Expect(nodenetworkchaos.OwnerReferences[0].Kind).To(Equal("Node"))
	expectLocal(nodenetworkchaos.Spec.IPSets)
	expectChains(nodenetworkchaos.Spec.Iptables, "")

	records[1].Phase = waitForApplySync
	nodenetworkchaos.Status.ObservedGeneration = networkchaos.Status.Instances[records[1].Id]
	g.Expect(c.Status().Update(ctx, nodenetworkchaos)).To(Succeed())
	phase, err = impl.Apply(ctx, 1, records, networkchaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))

	records[1].Phase = v1alpha1.Injected
	phase, err = impl.Recover(ctx, 1, records, networkchaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(waitForRecoverSync))

	g.Expect(c.Get(ctx, types.NamespacedName{Name: "node-a"}, nodenetworkchaos)).To(Succeed())
	g.Expect(nodenetworkchaos.Spec.IPSets).To(BeEmpty())
	g.Expect(nodenetworkchaos.Spec.Iptables).To(BeEmpty())
}

func TestRecoverNodeScopeOnDeletedNode(t *testing.T) {
	g := NewWithT(t)

	impl, _ := newNodeScopeImpl()
	networkchaos := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{Name: "gray", Namespace: "db"},
	}
	records := []*v1alpha1.Record{
		{Id: v1alpha1.NodeRecordIDPrefix + "gone", SelectorKey: ".", Phase: v1alpha1.Injected},
	}

	phase, err := impl.Recover(context.Background(), 0, records, networkchaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
}
