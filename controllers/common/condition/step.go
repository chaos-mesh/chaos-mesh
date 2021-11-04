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

func Step(pipeline *pipeline.Pipeline) reconcile.Reconciler {
	setupLog := pipeline.GetLogger().WithName("setup-condition")
	name := pipeline.GetObject().Name + "-condition"
	if !config.ShouldSpawnController(name) {
		return nil
	}

	setupLog.Info("setting up controller", "name", name)

	return &Reconciler{
		Object:   pipeline.GetObject().Object,
		Client:   pipeline.GetClient(),
		Recorder: pipeline.GetManager().GetEventRecorderFor("condition"),
		Log:      pipeline.GetLogger().WithName("condition"),
	}
}
