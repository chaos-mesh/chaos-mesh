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

package pipeline

import (
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/controllers/condition"
	"github.com/chaos-mesh/chaos-mesh/controllers/desiredphase"
	"github.com/chaos-mesh/chaos-mesh/controllers/finalizers"
	"github.com/chaos-mesh/chaos-mesh/controllers/records"
)

func ConditionStep(ctx *PipelineContext) reconcile.Reconciler {
	setupLog := ctx.Logger.WithName("setup-condition")
	name := ctx.Object.Name + "-condition"
	setupLog.Info("setting up controller", "name", name)

	return &condition.Reconciler{
		Object:   ctx.Object.Object,
		Client:   ctx.Client,
		Recorder: ctx.Mgr.GetEventRecorderFor("condition"),
		Log:      ctx.Logger.WithName("condition"),
	}
}

func DesiredPhaseStep(ctx *PipelineContext) reconcile.Reconciler {
	setupLog := ctx.Logger.WithName("setup-desiredphase")
	name := ctx.Object.Name + "-desiredphase"

	setupLog.Info("setting up controller", "name", name)

	return &desiredphase.Reconciler{
		Object:   ctx.Object.Object,
		Client:   ctx.Client,
		Recorder: ctx.RecorderBuilder.Build("desiredphase"),
		Log:      ctx.Logger.WithName("desiredphase"),
	}
}

func FinalizersStep(ctx *PipelineContext) reconcile.Reconciler {
	setupLog := ctx.Logger.WithName("setup-finalizers")
	name := ctx.Object.Name + "-finalizers"

	setupLog.Info("setting up controller", "name", name)

	return &finalizers.Reconciler{
		Object:   ctx.Object.Object,
		Client:   ctx.Client,
		Recorder: ctx.RecorderBuilder.Build("finalizers"),
		Log:      ctx.Logger.WithName("finalizers"),
	}
}

func RecordsStep(ctx *PipelineContext) reconcile.Reconciler {
	return &records.Reconciler{
		Impl:     ctx.Impl,
		Object:   ctx.Object.Object,
		Client:   ctx.Client,
		Reader:   ctx.Reader,
		Recorder: ctx.RecorderBuilder.Build("records"),
		Selector: ctx.Selector,
		Log:      ctx.Logger.WithName("records"),
	}
}

func AllSteps() []PipelineStep {
	return []PipelineStep{FinalizersStep, DesiredPhaseStep, ConditionStep, RecordsStep}
}
