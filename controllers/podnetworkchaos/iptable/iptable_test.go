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

package iptable

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func TestToChains(t *testing.T) {
	g := NewWithT(t)

	chains, err := ToChains([]v1alpha1.RawIptables{
		{
			Name:      "INPUT/full",
			Direction: v1alpha1.Input,
			IPSets:    []string{"tgt"},
		},
		{
			Name:          "OUTPUT/partial",
			Direction:     v1alpha1.Output,
			Device:        "eth1",
			DropPercent:   "40",
			ExcludeIPSets: []string{"local"},
		},
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(chains).To(HaveLen(2))

	g.Expect(chains[0].Name).To(Equal("INPUT/full"))
	g.Expect(chains[0].Direction).To(Equal(pb.Chain_INPUT))
	g.Expect(chains[0].Ipsets).To(Equal([]string{"tgt"}))
	g.Expect(chains[0].Target).To(Equal("DROP"))
	g.Expect(chains[0].Probability).To(BeZero())

	g.Expect(chains[1].Direction).To(Equal(pb.Chain_OUTPUT))
	g.Expect(chains[1].Device).To(Equal("eth1"))
	g.Expect(chains[1].Probability).To(BeNumerically("~", 0.4, 1e-9))
	g.Expect(chains[1].ExcludeIpsets).To(Equal([]string{"local"}))
}

func TestToChainsRejectsInvalidRules(t *testing.T) {
	g := NewWithT(t)

	_, err := ToChains([]v1alpha1.RawIptables{{Name: "x", Direction: "sideways"}})
	g.Expect(err).To(MatchError(ContainSubstring("unknown direction")))

	_, err = ToChains([]v1alpha1.RawIptables{{Name: "x", Direction: v1alpha1.Input, DropPercent: "forty"}})
	g.Expect(err).To(MatchError(ContainSubstring("parse dropPercent")))
}
