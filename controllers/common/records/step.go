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

package records

import (
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/controllers/common/pipeline"
)

func Step(ctx *pipeline.PipelineContext) reconcile.Reconciler {
	return &Reconciler{
		Impl:     ctx.Impl,
		Object:   ctx.Object.Object,
		Client:   ctx.Client,
		Reader:   ctx.Reader,
		Recorder: ctx.RecorderBuilder.Build("records"),
		Selector: ctx.Selector,
		Log:      ctx.Logger.WithName("records"),
	}
}
