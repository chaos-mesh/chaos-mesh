// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package records

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	controllerconfig "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
)

type contextImpl struct {
	ctx       context.Context
	operation Operation
	phase     v1alpha1.Phase
	err       error
	onCall    func(v1alpha1.InnerObject)
	calls     int
}

func (i *contextImpl) Apply(ctx context.Context, _ int, _ []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	i.ctx, i.operation = ctx, Apply
	i.calls++
	if i.onCall != nil {
		i.onCall(obj)
	}
	return i.phase, i.err
}

func (i *contextImpl) Recover(ctx context.Context, _ int, _ []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	i.ctx, i.operation = ctx, Recover
	i.calls++
	if i.onCall != nil {
		i.onCall(obj)
	}
	return i.phase, i.err
}

func TestReconcilePassesContext(t *testing.T) {
	for _, tc := range []struct {
		operation Operation
		desired   v1alpha1.DesiredPhase
		phase     v1alpha1.Phase
	}{
		{Apply, v1alpha1.RunningPhase, v1alpha1.NotInjected},
		{Recover, v1alpha1.StoppedPhase, v1alpha1.Injected},
	} {
		t.Run(string(tc.operation), func(t *testing.T) {
			scheme := runtime.NewScheme()
			if err := v1alpha1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			chaos := &v1alpha1.PhysicalMachineChaos{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}
			chaos.Status.Experiment.DesiredPhase = tc.desired
			chaos.Status.Experiment.Records = []*v1alpha1.Record{{Id: "target", Phase: tc.phase}}
			impl := &contextImpl{phase: tc.phase}
			r := Reconciler{
				Impl: impl, Object: &v1alpha1.PhysicalMachineChaos{},
				Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(chaos).Build(),
				Log:    logr.Discard(),
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(chaos)}); err != nil {
				t.Fatal(err)
			}
			if impl.operation != tc.operation || impl.ctx != ctx {
				t.Fatalf("operation = %s, context = %v; want %s with reconciliation context", impl.operation, impl.ctx, tc.operation)
			}
		})
	}
}

func recordFixture(t *testing.T, maxEvents int) (*v1alpha1.NetworkChaos, client.WithWatch) {
	t.Helper()
	g := NewWithT(t)
	scheme := runtime.NewScheme()
	g.Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	previous := controllerconfig.ControllerCfg
	controllerconfig.ControllerCfg = &config.ChaosControllerConfig{MaxEvents: maxEvents}
	t.Cleanup(func() { controllerconfig.ControllerCfg = previous })

	obj := &v1alpha1.NetworkChaos{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default", UID: "original"}}
	obj.Status.Experiment.DesiredPhase = v1alpha1.RunningPhase
	obj.Status.Experiment.Records = []*v1alpha1.Record{{Id: "default/target", Phase: v1alpha1.NotInjected}}
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj).Build()
	return obj, kube
}

func recordsReconciler(kube client.Client, impl *contextImpl) *Reconciler {
	return &Reconciler{
		Impl: impl, Object: &v1alpha1.NetworkChaos{}, Client: kube,
		Log: logr.Discard(), Recorder: recorder.NewDebugRecorder(),
	}
}

func TestReconcilePersistsRecordsAcrossAStatusConflict(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	obj, kube := recordFixture(t, 5)
	key := client.ObjectKeyFromObject(obj)
	req := ctrl.Request{NamespacedName: key}
	impl := &contextImpl{
		phase: v1alpha1.Injected,
		onCall: func(obj v1alpha1.InnerObject) {
			obj.(*v1alpha1.NetworkChaos).Status.Instances = map[string]int64{"default/target": 7}
		},
	}
	writes := 0
	r := recordsReconciler(interceptor.NewClient(kube, interceptor.Funcs{
		Get: func(got context.Context, c client.WithWatch, key client.ObjectKey, target client.Object, opts ...client.GetOption) error {
			g.Expect(got).To(Equal(ctx))
			return c.Get(got, key, target, opts...)
		},
		Update: func(got context.Context, c client.WithWatch, target client.Object, opts ...client.UpdateOption) error {
			g.Expect(got).To(Equal(ctx))
			writes++
			if writes == 1 {
				var latest v1alpha1.NetworkChaos
				g.Expect(c.Get(ctx, key, &latest)).To(Succeed())
				latest.Annotations = map[string]string{"owner": "concurrent-write"}
				latest.Status.Experiment.DesiredPhase = v1alpha1.StoppedPhase
				g.Expect(c.Update(ctx, &latest)).To(Succeed())
				err := c.Update(got, target, opts...)
				g.Expect(apierrors.IsConflict(err)).To(BeTrue())
				return err
			}
			return c.Update(got, target, opts...)
		},
	}), impl)

	g.Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
	g.Expect(writes).To(Equal(2))
	g.Expect(impl.calls).To(Equal(1), "retry must not repeat the external apply")
	g.Expect(kube.Get(ctx, key, obj)).To(Succeed())
	g.Expect(obj.Annotations).To(HaveKeyWithValue("owner", "concurrent-write"))
	g.Expect(obj.Status.Experiment.DesiredPhase).To(Equal(v1alpha1.StoppedPhase))
	g.Expect(obj.Status.Instances).To(HaveKeyWithValue("default/target", int64(7)))
	g.Expect(obj.Status.Experiment.Records).To(HaveLen(1))
	record := obj.Status.Experiment.Records[0]
	g.Expect(record.Phase).To(Equal(v1alpha1.Injected))
	g.Expect(record.InjectedCount).To(Equal(1))
	g.Expect(record.Events).To(HaveLen(1))
}

func TestReconcileDiscardsInFlightRecordsForDeletedOrReplacedExperiments(t *testing.T) {
	for _, replace := range []bool{false, true} {
		name := "deleted"
		if replace {
			name = "replaced"
		}
		t.Run(name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := t.Context()
			original, kube := recordFixture(t, 5)
			key := client.ObjectKeyFromObject(original)
			writes := 0
			impl := &contextImpl{
				phase: v1alpha1.Injected,
				onCall: func(obj v1alpha1.InnerObject) {
					obj.(*v1alpha1.NetworkChaos).Status.Instances = map[string]int64{"default/target": 7}
					g.Expect(kube.Delete(ctx, original)).To(Succeed())
					if replace {
						replacement := original.DeepCopy()
						replacement.ResourceVersion = ""
						replacement.UID = "replacement"
						replacement.Status.Experiment.Records = nil
						g.Expect(kube.Create(ctx, replacement)).To(Succeed())
					}
				},
			}
			r := recordsReconciler(interceptor.NewClient(kube, interceptor.Funcs{
				Update: func(cctx context.Context, c client.WithWatch, target client.Object, opts ...client.UpdateOption) error {
					writes++
					return c.Update(cctx, target, opts...)
				},
			}), impl)

			g.Expect(r.Reconcile(ctx, ctrl.Request{NamespacedName: key})).To(Equal(ctrl.Result{}))
			g.Expect(impl.calls).To(Equal(1))
			g.Expect(writes).To(BeZero())
			var latest v1alpha1.NetworkChaos
			if !replace {
				g.Expect(apierrors.IsNotFound(kube.Get(ctx, key, &latest))).To(BeTrue())
				return
			}
			g.Expect(kube.Get(ctx, key, &latest)).To(Succeed())
			g.Expect(string(latest.UID)).To(Equal("replacement"))
			g.Expect(latest.Status.Experiment.Records).To(BeEmpty())
			g.Expect(latest.Status.Instances).To(BeEmpty())
		})
	}
}

func TestRecoveryRetriesAndHonorsLoweredEventLimit(t *testing.T) {
	for _, limit := range []int{0, 2} {
		ctx := t.Context()
		g := NewWithT(t)
		obj, kube := recordFixture(t, limit)
		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)}
		obj.Status.Experiment.DesiredPhase = v1alpha1.StoppedPhase
		record := obj.Status.Experiment.Records[0]
		record.Phase = v1alpha1.Injected
		for range 4 {
			record.Events = append(record.Events, *v1alpha1.NewRecordEvent(v1alpha1.TypeSucceeded, v1alpha1.Apply, "", metav1.Now()))
		}
		g.Expect(kube.Update(ctx, obj)).To(Succeed())
		impl := &contextImpl{phase: v1alpha1.Injected, err: errors.New("daemon temporarily unavailable")}
		r := recordsReconciler(kube, impl)

		g.Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{Requeue: true}))
		g.Expect(kube.Get(ctx, req.NamespacedName, obj)).To(Succeed())
		record = obj.Status.Experiment.Records[0]
		g.Expect(record.Phase).To(Equal(v1alpha1.Injected))
		g.Expect(record.RecoveredCount).To(BeZero())
		g.Expect(record.Events).To(HaveLen(limit))

		impl.phase, impl.err = v1alpha1.NotInjected, nil
		g.Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
		g.Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
		g.Expect(impl.calls).To(Equal(2))
		g.Expect(kube.Get(ctx, req.NamespacedName, obj)).To(Succeed())
		record = obj.Status.Experiment.Records[0]
		g.Expect(record.Phase).To(Equal(v1alpha1.NotInjected))
		g.Expect(record.RecoveredCount).To(Equal(1))
		g.Expect(record.Events).To(HaveLen(limit))
		if limit > 0 {
			g.Expect(record.Events[0]).To(HaveField("Type", Equal(v1alpha1.TypeFailed)))
			g.Expect(record.Events[1]).To(HaveField("Type", Equal(v1alpha1.TypeSucceeded)))
		}
	}
}

func TestRecordReadErrorRequeuesWithoutApplying(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	obj, kube := recordFixture(t, 5)
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(obj)}
	failRead := true
	impl := &contextImpl{phase: v1alpha1.Injected}
	r := recordsReconciler(interceptor.NewClient(kube, interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, target client.Object, opts ...client.GetOption) error {
			if failRead {
				return errors.New("API read temporarily unavailable")
			}
			return c.Get(ctx, key, target, opts...)
		},
	}), impl)
	g.Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{Requeue: true}))
	g.Expect(impl.calls).To(BeZero())

	failRead = false
	g.Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
	g.Expect(impl.calls).To(Equal(1))
	g.Expect(kube.Get(ctx, req.NamespacedName, obj)).To(Succeed())
	g.Expect(obj.Status.Experiment.Records[0].Phase).To(Equal(v1alpha1.Injected))
}
