// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package podnetworkchaos

import (
	"testing"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	. "github.com/onsi/gomega"
)

func TestMergenetem(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		spec := v1alpha1.TcParameter{}
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

		spec := v1alpha1.TcParameter{
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
		em := &pb.Netem{
			Time:      90000,
			Jitter:    90000,
			DelayCorr: 25,
			Loss:      25,
			LossCorr:  25,
		}
		g.Expect(m).Should(Equal(em))
	})
}
