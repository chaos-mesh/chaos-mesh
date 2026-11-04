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

package remotechaosmonitor

import (
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
)

func TestBootstrapMultipleClusters(t *testing.T) {
	var registered struct {
		fx.In
		Objs []types.Object `group:"objs"`
	}
	app := fx.New(types.ChaosObjects, fx.Populate(&registered), fx.NopLogger)
	require.NoError(t, app.Err())
	require.NotEmpty(t, registered.Objs)

	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{v1alpha1.SchemeGroupVersion})
	var allowlist []string
	for _, obj := range registered.Objs {
		gvks, _, err := scheme.ObjectKinds(obj.Object)
		require.NoError(t, err)
		for _, gvk := range gvks {
			mapper.Add(gvk, meta.RESTScopeNamespace)
		}
		allowlist = append(allowlist, obj.Name+"-remotechaos-monitor")
	}

	enabledControllers := config.ControllerCfg.EnabledControllers
	t.Cleanup(func() {
		config.ControllerCfg.EnabledControllers = enabledControllers
	})

	for _, tc := range []struct {
		name    string
		enabled []string
	}{
		{name: "all controllers", enabled: []string{"*"}},
		{name: "legacy allowlist", enabled: allowlist},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config.ControllerCfg.EnabledControllers = tc.enabled
			// Controller names remain registered globally across repeated test runs.
			runID := uuid.NewString()
			for _, clusterName := range []string{"cluster-a-" + runID, "cluster-b-" + runID} {
				remoteClient := fake.NewClientBuilder().WithScheme(scheme).Build()
				mgr, err := ctrl.NewManager(&rest.Config{Host: "https://localhost"}, ctrl.Options{
					Scheme: scheme,
					MapperProvider: func(*rest.Config, *http.Client) (meta.RESTMapper, error) {
						return mapper, nil
					},
					NewClient: func(*rest.Config, client.Options) (client.Client, error) {
						return remoteClient, nil
					},
					Logger:                 logr.Discard(),
					Metrics:                metricsserver.Options{BindAddress: "0"},
					HealthProbeBindAddress: "0",
				})
				require.NoError(t, err)
				observedManager := &registrationManager{Manager: mgr}

				params := Params{
					Mgr:          observedManager,
					ManageClient: fake.NewClientBuilder().WithScheme(scheme).Build(),
					LocalClient:  remoteClient,
					ClusterName:  clusterName,
					Logger:       logr.Discard(),
					Objs:         registered.Objs,
				}
				require.NoError(t, Bootstrap(params), "register controllers for %s", clusterName)
				require.Equal(t, len(registered.Objs), observedManager.registrations)
			}
		})
	}
}

type registrationManager struct {
	ctrl.Manager
	registrations int
}

func (m *registrationManager) Add(runnable manager.Runnable) error {
	if err := m.Manager.Add(runnable); err != nil {
		return err
	}
	m.registrations++
	return nil
}
