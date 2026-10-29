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
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type contextImpl struct {
	ctx       context.Context
	operation Operation
	phase     v1alpha1.Phase
}

func (i *contextImpl) Apply(ctx context.Context, _ int, _ []*v1alpha1.Record, _ v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	i.ctx, i.operation = ctx, Apply
	return i.phase, nil
}

func (i *contextImpl) Recover(ctx context.Context, _ int, _ []*v1alpha1.Record, _ v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	i.ctx, i.operation = ctx, Recover
	return i.phase, nil
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
