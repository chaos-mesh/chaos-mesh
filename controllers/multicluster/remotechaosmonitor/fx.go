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

package remotechaosmonitor

import (
	"github.com/go-logr/logr"
	"go.uber.org/fx"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
)

type Params struct {
	fx.In

	Mgr          ctrl.Manager
	ManageClient client.Client `name:"manage-client"`
	LocalClient  client.Client
	ClusterName  string `name:"cluster-name"`
	Logger       logr.Logger
	Objs         []types.Object `group:"objs"`
}

func Bootstrap(params Params) error {
	logger := params.Logger
	mgr := params.Mgr
	objs := params.Objs
	setupLog := logger.WithName("setup-remotechaosmonitor")

	for _, obj := range objs {
		name := obj.Name + "-remotechaos-monitor"

		if !config.ShouldSpawnController(name) {
			return nil
		}

		setupLog.Info("setting up controller", "resource-name", obj.Name)

		// TODO: filter out chaos controlled by remote chaos
		builder := builder.Default(mgr).
			For(obj.Object).
			Named(obj.Name + "-remotechaos-monitor")

		err := builder.Complete(New(obj.Object, params.ManageClient, params.ClusterName, params.LocalClient, params.Logger))

		if err != nil {
			return err
		}

	}
	return nil
}

var Module = fx.Options(
	fx.Invoke(Bootstrap),
)
