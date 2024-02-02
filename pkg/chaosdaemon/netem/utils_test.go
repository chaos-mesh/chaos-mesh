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

package netem

import (
	"testing"

	. "github.com/onsi/gomega"

	chaosdaemonpb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func TestMergeNetem(t *testing.T) {
	g := NewGomegaWithT(t)

	cases := []struct {
		a      *chaosdaemonpb.Netem
		b      *chaosdaemonpb.Netem
		merged *chaosdaemonpb.Netem
	}{
		{nil, nil, nil}, // nil, nil -> nil
		{
			// no conflict
			&chaosdaemonpb.Netem{Loss: 25},
			&chaosdaemonpb.Netem{DelayCorr: 90},
			&chaosdaemonpb.Netem{Loss: 25, DelayCorr: 90},
		},
		{
			// pick the max
			&chaosdaemonpb.Netem{Loss: 25, DelayCorr: 100.2},
			&chaosdaemonpb.Netem{DelayCorr: 90},
			&chaosdaemonpb.Netem{Loss: 25, DelayCorr: 100.2},
		},
		{
			&chaosdaemonpb.Netem{Time: 100},
			&chaosdaemonpb.Netem{Rate: "800kbit"},
			&chaosdaemonpb.Netem{Time: 100, Rate: "800kbit"},
		},
	}

	for _, tc := range cases {
		m := MergeNetem(tc.a, tc.b)
		g.Expect(tc.merged).Should(Equal(m))
	}
}
