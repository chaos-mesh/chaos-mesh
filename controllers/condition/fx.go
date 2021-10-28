// Copyright 2021 Chaos Mesh Authors.
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

package condition

import (
	"github.com/go-logr/logr"
	"go.uber.org/fx"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
	"github.com/chaos-mesh/chaos-mesh/pkg/metrics"
)

type Params struct {
	fx.In

	Mgr              ctrl.Manager
	Client           client.Client
	Logger           logr.Logger
	MetricsCollector *metrics.ChaosControllerManagerMetricsCollector

	Objs []types.Object `group:"objs"`
}

func Bootstrap(params Params) error {
	logger := params.Logger

	setupLog := logger.WithName("setup-condition")
	for _, obj := range params.Objs {
		name := obj.Name + "-condition"
		if !config.ShouldSpawnController(name) {
			return nil
		}

		setupLog.Info("setting up controller", "resource-name", obj.Name)

		err := builder.Default(params.Mgr).
			For(obj.Object).
			Named(name).
			Complete(&Reconciler{
				Object:           obj.Object,
				Client:           params.Client,
				Recorder:         params.Mgr.GetEventRecorderFor("condition"),
				Log:              logger.WithName("condition"),
				MetricsCollector: params.MetricsCollector,
			})
		if err != nil {
			return err
		}

	}

	return nil
}
