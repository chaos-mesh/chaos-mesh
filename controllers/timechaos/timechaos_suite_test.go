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

package timechaos

import (
	"context"
	"errors"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	. "github.com/chaos-mesh/chaos-mesh/controllers/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func TestTimechaos(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"TimeChaos Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	Expect(v1.AddToScheme(scheme.Scheme)).To(Succeed())

	close(done)
}, 60)

var _ = AfterSuite(func() {
})

var _ = Describe("TimeChaos", func() {
	Context("TimeChaos", func() {
		podObjects, pods := GenerateNPods("p", 1, PodArg{})

		duration := "invalid_duration"

		timechaos := v1alpha1.TimeChaos{
			TypeMeta: metav1.TypeMeta{
				Kind:       "TimeChaos",
				APIVersion: "v1",
			},
			Spec: v1alpha1.TimeChaosSpec{
				Mode:       v1alpha1.AllPodMode,
				Value:      "0",
				Selector:   v1alpha1.SelectorSpec{Namespaces: []string{metav1.NamespaceDefault}},
				TimeOffset: "0s0ns",
				Duration:   &duration,
				Scheduler:  nil,
			},
		}

		r := endpoint{
			Context: ctx.Context{
				Client:        fake.NewFakeClientWithScheme(scheme.Scheme, podObjects...),
				EventRecorder: &record.FakeRecorder{},
				Log:           ctrl.Log.WithName("controllers").WithName("TimeChaos"),
			},
		}

		It("TimeChaos Apply", func() {
			defer mock.With("MockSelectAndFilterPods", func() []v1.Pod {
				return pods
			})()
			defer mock.With("MockChaosDaemonClient", &MockChaosDaemonClient{})()

			err := r.Apply(context.TODO(), ctrl.Request{}, &timechaos)

			Expect(err).ToNot(HaveOccurred())
		})

		It("TimeChaos Apply Error", func() {
			defer mock.With("MockSelectAndFilterPods", func() []v1.Pod {
				return pods
			})()
			defer mock.With("MockChaosDaemonClient", &MockChaosDaemonClient{})()
			defer mock.With("MockSetTimeOffsetError", errors.New("SetTimeOffsetError"))()

			err := r.Apply(context.TODO(), ctrl.Request{}, &timechaos)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("SetTimeOffsetError"))

		})

		It("TimeChaos Recover", func() {
			defer mock.With("MockSelectAndFilterPods", func() []v1.Pod {
				return pods
			})()
			defer mock.With("MockChaosDaemonClient", &MockChaosDaemonClient{})()

			err := r.Recover(context.TODO(), ctrl.Request{}, &timechaos)
			Expect(err).ToNot(HaveOccurred())
		})

		It("TimeChaos Recover Error", func() {
			defer mock.With("MockSelectAndFilterPods", func() []v1.Pod {
				return pods
			})()
			defer mock.With("MockChaosDaemonClient", &MockChaosDaemonClient{})()
			defer mock.With("MockRecoverTimeOffsetError", errors.New("RecoverTimeOffsetError"))()

			err := r.Apply(context.TODO(), ctrl.Request{}, &timechaos)
			Expect(err).ToNot(HaveOccurred())

			err = r.Recover(context.TODO(), ctrl.Request{}, &timechaos)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("RecoverTimeOffsetError"))
		})
	})
})
