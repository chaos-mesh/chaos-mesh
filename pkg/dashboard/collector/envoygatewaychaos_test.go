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

func TestConvertEnvoyGatewayChaosToExperiment(t *testing.T) {
	g := NewWithT(t)
	created := metav1.NewTime(time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC))
	chaos := &v1alpha1.EnvoyGatewayChaos{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.KindEnvoyGatewayChaos,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "api-fault",
			Namespace:         "chaos-mesh",
			UID:               types.UID("test-uid"),
			CreationTimestamp: created,
		},
		Spec: v1alpha1.EnvoyGatewayChaosSpec{
			Target: v1alpha1.EnvoyGatewayTarget{
				Namespace: "app",
				Kind:      v1alpha1.EnvoyGatewayHTTPRoute,
				Route:     "api",
			},
			Fault: v1alpha1.EnvoyGatewayFault{
				Delay: &v1alpha1.EnvoyGatewayDelay{FixedDelay: "250ms", Percentage: 20},
			},
		},
	}

	experiment, err := convertInnerObjectToExperiment(chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(experiment.Kind).To(Equal(v1alpha1.KindEnvoyGatewayChaos))
	g.Expect(experiment.Name).To(Equal("api-fault"))
	g.Expect(experiment.Action).To(BeEmpty())
	g.Expect(experiment.StartTime).To(Equal(created.Time))

	var stored v1alpha1.EnvoyGatewayChaos
	g.Expect(json.Unmarshal([]byte(experiment.Experiment), &stored)).To(Succeed())
	g.Expect(stored.Spec.Target.Route).To(Equal("api"))
	g.Expect(stored.Spec.Fault.Delay.FixedDelay).To(Equal("250ms"))
}
