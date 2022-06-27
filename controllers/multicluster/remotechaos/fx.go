// Copyright 2022 Chaos Mesh Authors.
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
	"github.com/go-logr/logr"
	"go.uber.org/fx"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/multicluster/clusterregistry"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
)

type Params struct {
	fx.In

	Mgr      ctrl.Manager
	Client   client.Client
	Logger   logr.Logger
	Objs     []types.Object `group:"objs"`
	Registry *clusterregistry.RemoteClusterRegistry
}

func Bootstrap(params Params) error {
	logger := params.Logger
	mgr := params.Mgr
	objs := params.Objs
	client := params.Client
	registry := params.Registry
	setupLog := logger.WithName("setup-remotechaos")

	for _, obj := range objs {
		name := obj.Name + "-records"
		if !config.ShouldSpawnController(name) {
			return nil
		}

		setupLog.Info("setting up controller", "resource-name", obj.Name)

		builder := builder.Default(mgr).
			For(obj.Object).
			Named(obj.Name + "-remotechaos").
			WithEventFilter(predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					obj, ok := e.Object.DeepCopyObject().(v1alpha1.RemoteObject)
					if !ok {
						return false
					}

					if obj.GetRemoteCluster().ClusterName == "" {
						return false
					}

					return true
				},
			})
		err := builder.Complete(&Reconciler{
			Client: client,
			Log:    logger.WithName("remotechaos"),

			registry: registry,
		})

		if err != nil {
			return err
		}

	}

	return nil
}
