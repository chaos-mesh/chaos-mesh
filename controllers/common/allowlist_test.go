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

package common

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	controllerconfig "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	chaosimpltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/pipeline"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

func TestBootstrapControllerAllowlist(t *testing.T) {
	pairs := []*chaosimpltypes.ChaosImplPair{
		{Name: "podchaos", Object: &v1alpha1.PodChaos{}},
		{Name: "networkchaos", Object: &v1alpha1.NetworkChaos{}},
		{Name: "stresschaos", Object: &v1alpha1.StressChaos{}},
	}
	g := NewWithT(t)
	scheme := runtime.NewScheme()
	g.Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{v1alpha1.SchemeGroupVersion})
	for _, pair := range pairs {
		gvks, _, err := scheme.ObjectKinds(pair.Object)
		g.Expect(err).NotTo(HaveOccurred())
		for _, gvk := range gvks {
			mapper.Add(gvk, meta.RESTScopeNamespace)
		}
	}

	enabledControllers := config.ControllerCfg.EnabledControllers
	t.Cleanup(func() {
		config.ControllerCfg.EnabledControllers = enabledControllers
	})

	for _, tc := range []struct {
		name    string
		enabled []string
		want    []string
	}{
		{
			name:    "wildcard",
			enabled: []string{"*"},
			want:    []string{"podchaos", "networkchaos", "stresschaos"},
		},
		{
			name:    "disabled first kind",
			enabled: []string{"networkchaos"},
			want:    []string{"networkchaos"},
		},
		{
			name:    "disabled middle kind",
			enabled: []string{"podchaos", "stresschaos"},
			want:    []string{"podchaos", "stresschaos"},
		},
		{name: "none"},
		{name: "unmatched", enabled: []string{"httpchaos"}},
		{name: "suffixed controller name", enabled: []string{"podchaos-records"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			config.ControllerCfg.EnabledControllers = tc.enabled
			localClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			skipNameValidation := true
			mgr, err := ctrl.NewManager(&rest.Config{Host: "https://localhost"}, ctrl.Options{
				Scheme: scheme,
				MapperProvider: func(*rest.Config, *http.Client) (meta.RESTMapper, error) {
					return mapper, nil
				},
				NewClient: func(*rest.Config, client.Options) (client.Client, error) {
					return localClient, nil
				},
				Controller:             controllerconfig.Controller{SkipNameValidation: &skipNameValidation},
				Logger:                 logr.Discard(),
				Metrics:                metricsserver.Options{BindAddress: "0"},
				HealthProbeBindAddress: "0",
			})
			g.Expect(err).NotTo(HaveOccurred())
			observedManager := &allowlistRegistrationManager{Manager: mgr}
			var completed []string
			steps := append(AllSteps(), func(ctx *pipeline.Context) reconcile.Reconciler {
				return reconcile.Func(func(context.Context, ctrl.Request) (ctrl.Result, error) {
					completed = append(completed, ctx.Object.Name)
					return ctrl.Result{}, nil
				})
			})
			g.Expect(Bootstrap(Params{
				Mgr:             observedManager,
				Client:          localClient,
				Reader:          localClient,
				Logger:          logr.Discard(),
				RecorderBuilder: recorder.NewRecorderBuilder(localClient, logr.Discard(), scheme, nil),
				Impls:           pairs,
				Steps:           steps,
			})).To(Succeed())
			g.Expect(observedManager.controllers).To(HaveLen(len(tc.want)))

			for _, controller := range observedManager.controllers {
				_, err := controller.Reconcile(context.Background(), ctrl.Request{
					NamespacedName: types.NamespacedName{Namespace: "default", Name: "missing"},
				})
				g.Expect(err).NotTo(HaveOccurred())
			}
			g.Expect(completed).To(Equal(tc.want), "every enabled kind must retain the complete pipeline")
		})
	}
}

type allowlistRegistrationManager struct {
	ctrl.Manager
	controllers []reconcile.Reconciler
}

func (m *allowlistRegistrationManager) Add(runnable manager.Runnable) error {
	if err := m.Manager.Add(runnable); err != nil {
		return err
	}
	if controller, ok := runnable.(reconcile.Reconciler); ok {
		m.controllers = append(m.controllers, controller)
	}
	return nil
}
