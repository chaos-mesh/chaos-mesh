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
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	controllerconfig "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
)

func TestBootstrapPartialAllowlist(t *testing.T) {
	objects := []types.Object{
		{Name: "podchaos", Object: &v1alpha1.PodChaos{}},
		{Name: "networkchaos", Object: &v1alpha1.NetworkChaos{}},
		{Name: "stresschaos", Object: &v1alpha1.StressChaos{}},
	}
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{v1alpha1.SchemeGroupVersion})
	for _, obj := range objects {
		gvks, _, err := scheme.ObjectKinds(obj.Object)
		require.NoError(t, err)
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
	}{
		{name: "disabled first kind", enabled: []string{"networkchaos-remote-apply"}},
		{name: "disabled middle kind", enabled: []string{"podchaos-remote-apply", "stresschaos-remote-apply"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config.ControllerCfg.EnabledControllers = tc.enabled
			localClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			// Bootstrap uses fixed controller names, including across repeated test runs.
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
			require.NoError(t, err)
			observedManager := &allowlistRegistrationManager{Manager: mgr}
			require.NoError(t, Bootstrap(Params{
				Mgr:    observedManager,
				Client: localClient,
				Logger: logr.Discard(),
				Objs:   objects,
			}))
			require.Equal(t, len(tc.enabled), observedManager.registrations,
				"all enabled kinds must register even after a disabled kind")
		})
	}
}

type allowlistRegistrationManager struct {
	ctrl.Manager
	registrations int
}

func (m *allowlistRegistrationManager) Add(runnable manager.Runnable) error {
	if err := m.Manager.Add(runnable); err != nil {
		return err
	}
	m.registrations++
	return nil
}
