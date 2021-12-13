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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/controllers/common/pipeline"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
)

func Step(ctx *pipeline.PipelineContext) reconcile.Reconciler {
	setupLog := ctx.Logger.WithName("setup-condition")
	name := ctx.Object.Name + "-condition"
	if !config.ShouldSpawnController(name) {
		return nil
	}

	setupLog.Info("setting up controller", "name", name)

	return &Reconciler{
		Object:   ctx.Object.Object,
		Client:   ctx.Client,
		Recorder: ctx.Mgr.GetEventRecorderFor("condition"),
		Log:      ctx.Logger.WithName("condition"),
	}
}
