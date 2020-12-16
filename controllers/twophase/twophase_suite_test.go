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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

func TestTwoPhase(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"TwoPhase Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	Expect(addFakeToScheme(scheme.Scheme)).To(Succeed())

	close(done)
}, 60)

var _ = AfterSuite(func() {
})

var _ end.Endpoint = (*fakeEndpoint)(nil)

type fakeEndpoint struct{}

func (r fakeEndpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	if err := mock.On("MockApplyError"); err != nil {
		return err.(error)
	}
	return nil
}

func (r fakeEndpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	if err := mock.On("MockRecoverError"); err != nil {
		return err.(error)
	}
	return nil
}

var _ v1alpha1.InnerSchedulerObject = (*fakeTwoPhaseChaos)(nil)

type fakeTwoPhaseChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            v1alpha1.ChaosStatus `json:"status,omitempty"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector v1alpha1.SelectorSpec `json:"selector"`

	Deleted bool `json:"deleted"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	Scheduler *v1alpha1.SchedulerSpec `json:"scheduler,omitempty"`

	// Next time when this action will be applied again
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
}

func (in *fakeTwoPhaseChaos) GetStatus() *v1alpha1.ChaosStatus {
	return &in.Status
}

// IsDeleted returns whether this resource has been deleted
func (in *fakeTwoPhaseChaos) IsDeleted() bool {
	return in.Deleted
}

// IsPaused returns whether this resource has been paused
func (in *fakeTwoPhaseChaos) IsPaused() bool {
	return false
}

func (r fakeEndpoint) Object() v1alpha1.InnerObject {
	return &fakeTwoPhaseChaos{}
}

func (in *fakeTwoPhaseChaos) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *fakeTwoPhaseChaos) GetNextStart() time.Time {
	if in.NextStart == nil {
		return time.Time{}
	}
	return in.NextStart.Time
}

func (in *fakeTwoPhaseChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.NextStart = nil
		return
	}

	if in.NextStart == nil {
		in.NextStart = &metav1.Time{}
	}
	in.NextStart.Time = t
}

func (in *fakeTwoPhaseChaos) GetNextRecover() time.Time {
	if in.NextRecover == nil {
		return time.Time{}
	}
	return in.NextRecover.Time
}

func (in *fakeTwoPhaseChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.NextRecover = nil
		return
	}

	if in.NextRecover == nil {
		in.NextRecover = &metav1.Time{}
	}
	in.NextRecover.Time = t
}

func (in *fakeTwoPhaseChaos) GetScheduler() *v1alpha1.SchedulerSpec {
	return in.Scheduler
}

func (in *fakeTwoPhaseChaos) GetChaos() *v1alpha1.ChaosInstance {
	return nil
}

func (in *fakeTwoPhaseChaos) DeepCopyInto(out *fakeTwoPhaseChaos) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)

	in.Status.DeepCopyInto(&out.Status)
	in.Selector.DeepCopyInto(&out.Selector)

	out.Deleted = in.Deleted

	if in.Duration != nil {
		in, out := &in.Duration, &out.Duration
		*out = new(string)
		**out = **in
	}
	if in.Scheduler != nil {
		in, out := &in.Scheduler, &out.Scheduler
		*out = new(v1alpha1.SchedulerSpec)
		**out = **in
	}
	if in.NextRecover != nil {
		in, out := &in.NextRecover, &out.NextRecover
		*out = new(metav1.Time)
		**out = **in
	}
	if in.NextStart != nil {
		in, out := &in.NextStart, &out.NextStart
		*out = new(metav1.Time)
		**out = **in
	}
}

func (in *fakeTwoPhaseChaos) DeepCopy() *fakeTwoPhaseChaos {
	if in == nil {
		return nil
	}
	out := new(fakeTwoPhaseChaos)
	in.DeepCopyInto(out)
	return out
}

func (in *fakeTwoPhaseChaos) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

var (
	schemeBuilder   = runtime.NewSchemeBuilder(addKnownTypes)
	addFakeToScheme = schemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(schema.GroupVersion{Group: "", Version: "v1"},
		&fakeTwoPhaseChaos{},
	)
	return nil
}

var _ = Describe("TwoPhase", func() {
	Context("TwoPhase", func() {
		var err error

		zeroTime := time.Time{}
		var _ = zeroTime
		pastTime := time.Now().Add(-10 * time.Hour)
		futureTime := time.Now().Add(10 * time.Hour)

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

		It("TwoPhase Action", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
			}

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := Reconciler{
				Endpoint: fakeEndpoint{},
				Context: ctx.Context{
					Client: c,
					Log:    ctrl.Log.WithName("controllers").WithName("TwoPhase"),
				},
			}

			_, err = r.Reconcile(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("misdefined scheduler"))
		})

		It("TwoPhase Delete", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
				Deleted:    true,
			}

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := Reconciler{
				Endpoint: fakeEndpoint{},
				Context: ctx.Context{
					Client: c,
					Log:    ctrl.Log.WithName("controllers").WithName("TwoPhase"),
				},
			}

			_, err = r.Reconcile(req)

			Expect(err).ToNot(HaveOccurred())
			_chaos := r.Object()
			err = r.Client.Get(context.TODO(), req.NamespacedName, _chaos)
			Expect(err).ToNot(HaveOccurred())
			Expect(_chaos.(v1alpha1.InnerSchedulerObject).GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseFinished))

			defer mock.With("MockRecoverError", errors.New("RecoverError"))()

			chaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning
			err := c.Update(context.TODO(), &chaos)
			Expect(err).NotTo(HaveOccurred())

			_, err = r.Reconcile(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("RecoverError"))
		})

		It("TwoPhase ToRecover", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			}

			chaos.SetNextRecover(pastTime)
			chaos.SetNextStart(futureTime)

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := Reconciler{
				Endpoint: fakeEndpoint{},
				Context: ctx.Context{
					Client: c,
					Log:    ctrl.Log.WithName("controllers").WithName("TwoPhase"),
				},
			}

			_, err = r.Reconcile(req)

			Expect(err).ToNot(HaveOccurred())
			_chaos := r.Object()
			err = r.Client.Get(context.TODO(), req.NamespacedName, _chaos)
			Expect(err).ToNot(HaveOccurred())
			Expect(_chaos.(v1alpha1.InnerSchedulerObject).GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseWaiting))
		})

		It("TwoPhase ToRecover Error", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			}

			defer mock.With("MockRecoverError", errors.New("RecoverError"))()
			chaos.SetNextRecover(pastTime)
			chaos.SetNextStart(futureTime)
			chaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := Reconciler{
				Endpoint: fakeEndpoint{},
				Context: ctx.Context{
					Client: c,
					Log:    ctrl.Log.WithName("controllers").WithName("TwoPhase"),
				},
			}

			_, err = r.Reconcile(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("RecoverError"))
		})

		It("TwoPhase ToApply", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			}

			chaos.SetNextRecover(futureTime)
			chaos.SetNextStart(pastTime)

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := Reconciler{
				Endpoint: fakeEndpoint{},
				Context: ctx.Context{
					Client: c,
					Log:    ctrl.Log.WithName("controllers").WithName("TwoPhase"),
				},
			}

			_, err = r.Reconcile(req)

			Expect(err).ToNot(HaveOccurred())
			_chaos := r.Object()
			err = r.Client.Get(context.TODO(), req.NamespacedName, _chaos)
			Expect(err).ToNot(HaveOccurred())
			Expect(_chaos.(v1alpha1.InnerSchedulerObject).GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseRunning))
		})

		It("TwoPhase ToApplyAgain", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			}

			chaos.SetNextRecover(futureTime)
			chaos.SetNextStart(pastTime)

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := Reconciler{
				Endpoint: fakeEndpoint{},
				Context: ctx.Context{
					Client: c,
					Log:    ctrl.Log.WithName("controllers").WithName("TwoPhase"),
				},
			}

			_, err = r.Reconcile(req)

			Expect(err).ToNot(HaveOccurred())
			_chaos := r.Object()
			err = r.Client.Get(context.TODO(), req.NamespacedName, _chaos)
			Expect(err).ToNot(HaveOccurred())
			Expect(_chaos.(v1alpha1.InnerSchedulerObject).GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseRunning))

			chaos.Status.Experiment.StartTime = &metav1.Time{Time: pastTime}
			chaos.Scheduler = &v1alpha1.SchedulerSpec{Cron: "@every 20h"}
			chaos.SetNextStart(futureTime)
			_ = c.Update(context.TODO(), &chaos)

			_, err = r.Reconcile(req)
			Expect(err).ToNot(HaveOccurred())
			err = r.Client.Get(context.TODO(), req.NamespacedName, _chaos)
			Expect(err).ToNot(HaveOccurred())
			Expect(_chaos.(v1alpha1.InnerSchedulerObject).GetStatus().Experiment.Phase).To(Equal(v1alpha1.ExperimentPhaseRunning))
			d, _ := time.ParseDuration("10h")
			exp := time.Now().Add(d)
			Expect(chaos.NextStart.Time.Year()).To(Equal(exp.Year()))
			Expect(chaos.NextStart.Time.Month()).To(Equal(exp.Month()))
			Expect(chaos.NextStart.Time.Day()).To(Equal(exp.Day()))
			Expect(chaos.NextStart.Time.Hour()).To(Equal(exp.Hour()))
			Expect(chaos.NextStart.Time.Minute()).To(Equal(exp.Minute()))
			Expect(exp.Second()-chaos.NextStart.Time.Second() < 2).To(Equal(true))
		})

		It("TwoPhase ToApply Error", func() {
			chaos := fakeTwoPhaseChaos{
				TypeMeta:   typeMeta,
				ObjectMeta: objectMeta,
				Scheduler:  &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			}

			chaos.SetNextRecover(futureTime)
			chaos.SetNextStart(pastTime)

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos)

			r := Reconciler{
				Endpoint: fakeEndpoint{},
				Context: ctx.Context{
					Client: c,
					Log:    ctrl.Log.WithName("controllers").WithName("TwoPhase"),
				},
			}

			defer mock.With("MockApplyError", errors.New("ApplyError"))()

			_, err = r.Reconcile(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ApplyError"))
		})
	})
})
