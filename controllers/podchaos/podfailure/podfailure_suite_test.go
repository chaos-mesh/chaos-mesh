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

package podfailure

import (
	"context"
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

func TestPodFailure(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"PodFailure Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	Expect(v1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(v1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())

	close(done)
}, 60)

var _ = AfterSuite(func() {
})

var _ = Describe("PodChaos", func() {
	Context("PodFailure", func() {
		objs, pods := GenerateNPods("p", 1, PodArg{ContainerStatus: v1.ContainerStatus{
			ContainerID: "fake-container-id",
			Name:        "container-name",
		}})

		podChaos := v1alpha1.PodChaos{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodChaos",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceDefault,
				Name:      "podchaos-name",
			},
			Spec: v1alpha1.PodChaosSpec{
				Selector:      v1alpha1.SelectorSpec{Namespaces: []string{metav1.NamespaceDefault}},
				Mode:          v1alpha1.OnePodMode,
				ContainerName: "container-name",
				Scheduler:     &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			},
		}

		r := endpoint{
			Context: ctx.Context{
				Client:        fake.NewFakeClientWithScheme(scheme.Scheme, objs...),
				EventRecorder: &record.FakeRecorder{},
				Log:           ctrl.Log.WithName("controllers").WithName("PodChaos"),
			},
		}

		It("PodFailure Action", func() {
			defer mock.With("MockChaosDaemonClient", &MockChaosDaemonClient{})()
			defer mock.With("MockSelectAndFilterPods", func() []v1.Pod {
				return pods
			})()

			var err error

			err = r.Apply(context.TODO(), ctrl.Request{}, &podChaos)
			Expect(err).ToNot(HaveOccurred())

			err = r.Recover(context.TODO(), ctrl.Request{}, &podChaos)
			Expect(err).ToNot(HaveOccurred())
		})

	})
})
