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

package netem

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubectl/pkg/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/test"
	chaosdaemonpb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
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

func TestReconciler_applyNetem(t *testing.T) {
	g := NewWithT(t)

	podObjects, pods := test.GenerateNPods(
		"p",
		1,
		v1.PodRunning,
		metav1.NamespaceDefault,
		nil,
		map[string]string{"l1": "l1"},
		v1.ContainerStatus{ContainerID: "fake-container-id"},
	)

	r := Reconciler{
		Client:        fake.NewFakeClientWithScheme(scheme.Scheme, podObjects...),
		EventRecorder: &record.FakeRecorder{},
		Log:           ctrl.Log.WithName("controllers").WithName("TimeChaos"),
	}

	t.Run("netem without filter", func(t *testing.T) {
		networkChaos := v1alpha1.NetworkChaos{
			TypeMeta: metav1.TypeMeta{
				Kind:       "TimeChaos",
				APIVersion: "v1",
			},
			Spec: v1alpha1.NetworkChaosSpec{
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
			},
		}

		defer mock.With("MockSelectAndFilterPods", func() []v1.Pod {
			return pods
		})()
		defer mock.With("MockChaosDaemonClient", &test.MockChaosDaemonClient{})()

		err := r.Apply(context.TODO(), ctrl.Request{}, &networkChaos)

		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("netem with targets", func(t *testing.T) {
		networkChaos := v1alpha1.NetworkChaos{
			TypeMeta: metav1.TypeMeta{
				Kind:       "TimeChaos",
				APIVersion: "v1",
			},
			Spec: v1alpha1.NetworkChaosSpec{
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
				Direction: v1alpha1.To,
				Target: &v1alpha1.Target{
					TargetSelector: v1alpha1.SelectorSpec{},
				},
				ExternalTargets: []string{"8.8.8.8", "www.google.com"},
			},
		}

		defer mock.With("MockSelectAndFilterPods", func() []v1.Pod {
			return pods
		})()
		defer mock.With("MockChaosDaemonClient", &test.MockChaosDaemonClient{})()

		err := r.Apply(context.TODO(), ctrl.Request{}, &networkChaos)

		g.Expect(err).ToNot(HaveOccurred())
	})
}
