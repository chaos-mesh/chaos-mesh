// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package condition

import (
	"github.com/go-logr/logr"
	"go.uber.org/fx"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
)

type Objs struct {
	fx.In

	Objs []types.Object `group:"objs"`
}

func NewController(mgr ctrl.Manager, client client.Client, logger logr.Logger, objs Objs) (types.Controller, error) {
	setupLog := logger.WithName("setup-condition")
	for _, obj := range objs.Objs {
		setupLog.Info("setting up controller", "resource-name", obj.Name)

		err := builder.Default(mgr).
			For(obj.Object).
			Named(obj.Name + "-condition").
			Complete(&Reconciler{
				Object:   obj.Object,
				Client:   client,
				Recorder: mgr.GetEventRecorderFor("condition"),
				Log:      logger.WithName("condition"),
			})
		if err != nil {
			return "", err
		}

	}

	return "condition", nil
}
