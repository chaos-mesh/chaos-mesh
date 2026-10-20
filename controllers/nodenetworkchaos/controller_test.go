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

package nodenetworkchaos

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/test"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

type recordingDaemonClient struct {
	test.MockChaosDaemonClient

	ipsets   []*pb.IPSetsRequest
	iptables []*pb.IptablesChainsRequest
}

func (c *recordingDaemonClient) FlushIPSets(ctx context.Context, in *pb.IPSetsRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	c.ipsets = append(c.ipsets, in)
	return nil, nil
}

func (c *recordingDaemonClient) SetIptablesChains(ctx context.Context, in *pb.IptablesChainsRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	c.iptables = append(c.iptables, in)
	return nil, nil
}

func TestApplyTargetsTheHostNetworkNamespace(t *testing.T) {
	g := NewWithT(t)

	daemon := &recordingDaemonClient{}
	r := &Reconciler{ChaosDaemonPort: 31767}
	obj := &v1alpha1.NodeNetworkChaos{
		Spec: v1alpha1.NodeNetworkChaosSpec{
			IPSets: []v1alpha1.RawIPSet{
				{Name: "gray_local", IPSetType: v1alpha1.NetIPSet, Cidrs: []string{"10.0.1.5/32"}},
			},
			Iptables: []v1alpha1.RawIptables{
				{Name: "INPUT/gray", Direction: v1alpha1.Input, DropPercent: "40", ExcludeIPSets: []string{"gray_local"}},
			},
		},
	}
	obj.Name = "node-a"

	g.Expect(r.applyWith(context.Background(), daemon, obj)).To(Succeed())

	g.Expect(daemon.ipsets).To(HaveLen(1))
	g.Expect(daemon.ipsets[0].HostNs).To(BeTrue())
	g.Expect(daemon.ipsets[0].EnterNS).To(BeFalse())
	g.Expect(daemon.ipsets[0].Ipsets).To(HaveLen(1))
	g.Expect(daemon.ipsets[0].Ipsets[0].Name).To(Equal("gray_local"))

	g.Expect(daemon.iptables).To(HaveLen(1))
	g.Expect(daemon.iptables[0].HostNs).To(BeTrue())
	g.Expect(daemon.iptables[0].Chains).To(HaveLen(1))
	chain := daemon.iptables[0].Chains[0]
	g.Expect(chain.Direction).To(Equal(pb.Chain_INPUT))
	g.Expect(chain.Probability).To(BeNumerically("~", 0.4))
	g.Expect(chain.ExcludeIpsets).To(Equal([]string{"gray_local"}))
	g.Expect(chain.ExemptTcpPorts).To(Equal("10250,31767"))
}
