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

package trafficcontrol

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
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func TestReconciler_applyNetem(t *testing.T) {
	g := NewWithT(t)

	podObjects, pods := GenerateNPods("p", 1, PodArg{})

	v1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)

	r := endpoint{
		Context: ctx.Context{
			Client:        fake.NewFakeClientWithScheme(scheme.Scheme, podObjects...),
			EventRecorder: &record.FakeRecorder{},
			Log:           ctrl.Log.WithName("controllers").WithName("NetworkChaos"),
		},
	}

	t.Run("netem without filter", func(t *testing.T) {
		networkChaos := v1alpha1.NetworkChaos{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NetworkChaos",
				APIVersion: "v1",
			},
			Spec: v1alpha1.NetworkChaosSpec{
				Action:    "netem",
				Direction: v1alpha1.To,
				TcParameter: v1alpha1.TcParameter{
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
				TcParameter: v1alpha1.TcParameter{
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
