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
	"context"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-controller-manager/provider"
	. "github.com/chaos-mesh/chaos-mesh/controllers/test"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func setHostNetwork(objs []runtime.Object) {
	for _, obj := range objs {
		if pod, ok := obj.(*v1.Pod); ok {
			pod.Spec.HostNetwork = true
		}
	}
}

func TestHostNetworkOption(t *testing.T) {
	defer mock.With("MockChaosDaemonClient", &MockChaosDaemonClient{})()
	RegisterTestingT(t)

	testCases := []struct {
		name                     string
		enableHostNetworkTesting bool
		errorEvaluation          func(err string)
	}{
		{
			name:                     "host networking testing disabled (default)",
			enableHostNetworkTesting: false,
			errorEvaluation: func(err string) {
				Expect(err).To(ContainSubstring("It's dangerous to inject network chaos on a pod"))
			},
		},
		{
			name:                     "host networking testing enabled",
			enableHostNetworkTesting: true,
			errorEvaluation: func(err string) {
				Expect(err).To(Equal(""))
			},
		},
	}

	for _, testCase := range testCases {

		objs, _ := GenerateNPods("p", 1, PodArg{})

		setHostNetwork(objs)

		chaos := &v1alpha1.PodNetworkChaos{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodNetworkChaos",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:  metav1.NamespaceDefault,
				Name:       "p0",
				Generation: 1,
			},
			Spec: v1alpha1.PodNetworkChaosSpec{},
		}
		objs = append(objs, chaos)

		fakeClient := fake.NewFakeClientWithScheme(provider.NewScheme(), objs...)

		recorder := recorder.NewDebugRecorder()
		h := &Reconciler{
			Client:                  fakeClient,
			Recorder:                recorder,
			Log:                     zap.New(zap.UseDevMode(true)),
			AllowHostNetworkTesting: testCase.enableHostNetworkTesting,
		}

		_, err := h.Reconcile(ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: metav1.NamespaceDefault,
				Name:      "p0",
			},
		})
		Expect(err).To(BeNil())

		fakeClient.Get(context.Background(), types.NamespacedName{
			Namespace: metav1.NamespaceDefault,
			Name:      "p0",
		}, chaos)

		testCase.errorEvaluation(chaos.Status.FailedMessage)
	}
}

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
