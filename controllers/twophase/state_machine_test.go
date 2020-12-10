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

package twophase

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
)

func TestStateMachine(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Twophase StateMachine Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = Describe("TwoPhase StateMachine", func() {
	Context("TwoPhase", func() {

		statuses := []v1alpha1.ExperimentPhase{
			v1alpha1.ExperimentPhaseFailed,
			v1alpha1.ExperimentPhaseFinished,
			v1alpha1.ExperimentPhasePaused,
			v1alpha1.ExperimentPhaseRunning,
			v1alpha1.ExperimentPhaseWaiting,
		}

		It("StateMachine Target Finish", func() {
			defer mock.With("MockApplyError", errors.New("ApplyError"))()
			defer mock.With("MockRecoverError", errors.New("RecoverError"))()

			for _, status := range statuses {
				now := time.Now()
				sm := setupStateMachineWithStatus(status)

				updated, err := sm.run(context.TODO(), v1alpha1.ExperimentPhaseFinished, now)

				if status == v1alpha1.ExperimentPhaseRunning {
					Expect(updated).To(Equal(true))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("RecoverError"))
					Expect(sm.Chaos.GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseRunning))
					Expect(sm.Chaos.GetStatus().FailedMessage).To(ContainSubstring("RecoverError"))

					return
				} else if status != v1alpha1.ExperimentPhaseFinished {
					Expect(updated).To(Equal(true))
				} else {
					Expect(updated).To(Equal(false))
				}

				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("StateMachine Target Paused", func() {
			defer mock.With("MockApplyError", errors.New("ApplyError"))()
			defer mock.With("MockRecoverError", errors.New("RecoverError"))()

			for _, status := range statuses {
				now := time.Now()
				sm := setupStateMachineWithStatus(status)

				updated, err := sm.run(context.TODO(), v1alpha1.ExperimentPhasePaused, now)

				if status == v1alpha1.ExperimentPhaseRunning {
					Expect(updated).To(Equal(true))
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("RecoverError"))
					Expect(sm.Chaos.GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseRunning))
					Expect(sm.Chaos.GetStatus().FailedMessage).To(ContainSubstring("RecoverError"))

					return
				} else if status == v1alpha1.ExperimentPhaseFinished {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("turn from"))

					return
				} else if status != v1alpha1.ExperimentPhasePaused {
					Expect(updated).To(Equal(true))
				} else {
					Expect(updated).To(Equal(false))
				}

				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("StateMachine Target Running", func() {
			defer mock.With("MockApplyError", errors.New("ApplyError"))()
			defer mock.With("MockRecoverError", errors.New("RecoverError"))()

			for _, status := range statuses {
				now := time.Now()
				sm := setupStateMachineWithStatus(status)

				updated, err := sm.run(context.TODO(), v1alpha1.ExperimentPhaseRunning, now)

				if status == v1alpha1.ExperimentPhaseFinished {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("turn from"))
				} else if status == v1alpha1.ExperimentPhaseRunning {
					Expect(updated).To(Equal(false))
					Expect(err).ToNot(HaveOccurred())
				} else if status != v1alpha1.ExperimentPhasePaused {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("ApplyError"))
				}
			}
		})

		It("Pause", func() {
			// duration 15min, scheduler @every 20m
			// Then it should be running in 13:10-13:25, 13:30-13:45, 13:50-14:05, 14:10-14:25
			// Pause will only erase part of it
			now, err := time.Parse(time.RFC3339, "2020-12-07T13:10:00+00:00")
			Expect(err).ToNot(HaveOccurred())
			sm := setupStateMachineWithStatus(v1alpha1.ExperimentPhaseUninitialized)

			updated, err := sm.run(context.TODO(), v1alpha1.ExperimentPhaseRunning, now)
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal(true))
			Expect(sm.Chaos.GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseRunning))

			now = now.Add(time.Minute) // 13:11
			updated, err = sm.run(context.TODO(), v1alpha1.ExperimentPhasePaused, now)
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal(true))
			Expect(sm.Chaos.GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhasePaused))

			// should apply
			now, err = time.Parse(time.RFC3339, "2020-12-07T13:55:00+00:00")
			Expect(err).ToNot(HaveOccurred())
			updated, err = sm.run(context.TODO(), v1alpha1.ExperimentPhaseRunning, now)
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal(true))
			Expect(sm.Chaos.GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseRunning))

			now = now.Add(time.Minute) // 13:56
			updated, err = sm.run(context.TODO(), v1alpha1.ExperimentPhasePaused, now)
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal(true))
			Expect(sm.Chaos.GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhasePaused))

			now, err = time.Parse(time.RFC3339, "2020-12-07T14:06:00+00:00")
			Expect(err).ToNot(HaveOccurred())
			updated, err = sm.run(context.TODO(), v1alpha1.ExperimentPhaseRunning, now)
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal(true))
			Expect(sm.Chaos.GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseWaiting))

			now, err = time.Parse(time.RFC3339, "2020-12-07T14:11:00+00:00")
			Expect(err).ToNot(HaveOccurred())
			updated, err = sm.run(context.TODO(), v1alpha1.ExperimentPhaseRunning, now)
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(Equal(true))
			Expect(sm.Chaos.GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseRunning))
		})
	})

})

func setupStateMachineWithStatus(status v1alpha1.ExperimentPhase) *chaosStateMachine {
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "fakechaos-name",
			Namespace: metav1.NamespaceDefault,
		},
	}
	typeMeta := metav1.TypeMeta{
		Kind:       "PodChaos",
		APIVersion: "v1",
	}
	objectMeta := metav1.ObjectMeta{
		Namespace: metav1.NamespaceDefault,
		Name:      "fakechaos-name",
	}

	chaos := fakeTwoPhaseChaos{
		TypeMeta:   typeMeta,
		ObjectMeta: objectMeta,
	}

	chaos.Status.Experiment.Phase = status
	duration := "15m"
	chaos.Duration = &duration
	chaos.Scheduler = &v1alpha1.SchedulerSpec{
		Cron: "@every 20m",
	}

	c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

	r := Reconciler{
		Endpoint: fakeEndpoint{},
		Context: ctx.Context{
			Client: c,
			Log:    ctrl.Log.WithName("controllers").WithName("TwoPhase"),
		},
	}

	sm := chaosStateMachine{
		Chaos:      &chaos,
		Req:        req,
		Reconciler: &r,
	}

	return &sm
}
