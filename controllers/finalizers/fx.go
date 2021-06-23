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

package finalizers

import (
	"github.com/go-logr/logr"
	"go.uber.org/fx"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type Objs struct {
	fx.In

	Objs []types.Object `group:"objs"`
}

func NewController(mgr ctrl.Manager, client client.Client, logger logr.Logger, recorderBuilder *recorder.RecorderBuilder, pairs Objs) (types.Controller, error) {
	for _, obj := range pairs.Objs {
		err := builder.Default(mgr).
			For(obj.Object).
			Named(obj.Name + "-finalizers").
			Complete(&Reconciler{
				Object:   obj.Object,
				Client:   client,
				Recorder: recorderBuilder.Build("finalizer"),
				Log:      logger.WithName("finalizers"),
			})
		if err != nil {
			return "", err
		}

	}

	return "finalizers", nil
}
