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

package clusterregistry

import (
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/multicluster/remotechaosmonitor"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
)

func TestRemoteManagerNameReuse(t *testing.T) {
	enabled := config.ControllerCfg.EnabledControllers
	t.Cleanup(func() { config.ControllerCfg.EnabledControllers = enabled })
	config.ControllerCfg.EnabledControllers = []string{"*"}
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{v1alpha1.SchemeGroupVersion})
	mapper.Add(v1alpha1.SchemeGroupVersion.WithKind("PodChaos"), meta.RESTScopeNamespace)
	clusterName := "recreated-" + uuid.NewString()
	for generation := 0; generation < 2; generation++ {
		c := fake.NewClientBuilder().WithScheme(scheme).Build()
		options := controllerManagerOption(scheme)
		options.MapperProvider = func(*rest.Config, *http.Client) (meta.RESTMapper, error) { return mapper, nil }
		options.NewClient = func(*rest.Config, client.Options) (client.Client, error) { return c, nil }
		options.Logger = logr.Discard()
		options.HealthProbeBindAddress = "0"
		mgr, err := ctrl.NewManager(&rest.Config{Host: "https://localhost"}, *options)
		require.NoError(t, err)
		require.NoError(t, remotechaosmonitor.Bootstrap(remotechaosmonitor.Params{Mgr: mgr, ManageClient: c, LocalClient: c, ClusterName: clusterName, Logger: logr.Discard(), Objs: []types.Object{{Name: "podchaos", Object: &v1alpha1.PodChaos{}}}}), "register manager generation %d", generation)
	}
}
