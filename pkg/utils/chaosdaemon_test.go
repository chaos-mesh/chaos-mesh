package utils

import (
	"testing"

	. "github.com/onsi/gomega"

	chaosdaemonpb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
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
	}

	for _, tc := range cases {
		m := MergeNetem(tc.a, tc.b)
		g.Expect(tc.merged).Should(Equal(m))
	}
}
