// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestConvertIstioChaosToExperiment(t *testing.T) {
	g := NewWithT(t)
	created := metav1.NewTime(time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC))
	chaos := &v1alpha1.IstioChaos{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.KindIstioChaos,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "checkout-fault",
			Namespace:         "chaos-mesh",
			UID:               types.UID("test-uid"),
			CreationTimestamp: created,
		},
		Spec: v1alpha1.IstioChaosSpec{
			Target: v1alpha1.IstioTarget{
				Namespace:      "checkout",
				VirtualService: "checkout",
			},
			Fault: v1alpha1.IstioFault{
				Abort: &v1alpha1.IstioAbort{HTTPStatus: 503, Percentage: 20},
			},
		},
	}

	experiment, err := convertInnerObjectToExperiment(chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(experiment.Kind).To(Equal(v1alpha1.KindIstioChaos))
	g.Expect(experiment.Name).To(Equal("checkout-fault"))
	g.Expect(experiment.Namespace).To(Equal("chaos-mesh"))
	g.Expect(experiment.UID).To(Equal("test-uid"))
	g.Expect(experiment.Action).To(BeEmpty())
	g.Expect(experiment.StartTime).To(Equal(created.Time))

	var stored v1alpha1.IstioChaos
	g.Expect(json.Unmarshal([]byte(experiment.Experiment), &stored)).To(Succeed())
	g.Expect(stored.Spec.Target.VirtualService).To(Equal("checkout"))
	g.Expect(stored.Spec.Fault.Abort.HTTPStatus).To(Equal(int32(503)))
}
