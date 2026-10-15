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
//

package remotechaos

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/multicluster/clusterregistry"
)

func TestReconcileRetriesWhenRemoteManagerMissing(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	experiment := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "experiment"}, Spec: v1alpha1.PodChaosSpec{RemoteCluster: "not-yet-registered"}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(experiment).Build()
	r := Reconciler{Client: c, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}, registry: clusterregistry.New(logr.Discard(), c)}
	result, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Namespace: experiment.Namespace, Name: experiment.Name}})
	require.True(t, err != nil || result.Requeue || result.RequeueAfter > 0, "missing manager must retry without requiring another experiment edit")
}

func TestReconcileRetriesManagementReadFailure(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	c := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
		return errors.New("temporary management API failure")
	}}).Build()
	r := Reconciler{Client: c, Log: logr.Discard(), Object: &v1alpha1.PodChaos{}}
	result, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "test", Name: "experiment"}})
	require.True(t, err != nil || result.Requeue || result.RequeueAfter > 0, "management read failure must retry")
}
