package netem

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"

	chaosdaemonpb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
)

func TestMergenetem(t *testing.T) {

	t.Run("empty", func(t *testing.T) {
		spec := v1alpha1.NetworkChaosSpec{
			Action: "netem",
		}
		_, err := mergeNetem(spec)
		if err == nil {
			t.Errorf("expect invalid spec failed with message %s but got nil", invalidNetemSpecMsg)
		}
		if err != nil && err.Error() != invalidNetemSpecMsg {
			t.Errorf("expect merge failed with message %s but got %v", invalidNetemSpecMsg, err)
		}
	})

	t.Run("delay loss", func(t *testing.T) {
		g := NewGomegaWithT(t)

		spec := v1alpha1.NetworkChaosSpec{
			Action: "netem",
			Delay: &v1alpha1.DelaySpec{
				Latency:     "90ms",
				Correlation: "25",
				Jitter:      "90ms",
			},
			Loss: &v1alpha1.LossSpec{
				Loss:        "25",
				Correlation: "25",
			},
		}
		m, err := mergeNetem(spec)
		g.Expect(err).ShouldNot(HaveOccurred())
		em := &chaosdaemonpb.Netem{
			Time:      90000,
			Jitter:    90000,
			DelayCorr: 25,
			Loss:      25,
			LossCorr:  25,
		}
		g.Expect(m).Should(Equal(em))
	})
}
