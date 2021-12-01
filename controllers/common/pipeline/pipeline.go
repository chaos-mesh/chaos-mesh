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

package pipeline

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	chaosimpltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

type Pipeline struct {
	steps []reconcile.Reconciler
	ctx   *PipelineContext
}

type PipelineContext struct {
	Object *types.Object
	Mgr    ctrl.Manager
	Client client.Client
	client.Reader

	Logger          logr.Logger
	RecorderBuilder *recorder.RecorderBuilder
	Impl            chaosimpltypes.ChaosImpl
	Selector        *selector.Selector
}

type PipelineStep func(ctx *PipelineContext) reconcile.Reconciler

func NewPipeline(ctx *PipelineContext) *Pipeline {
	return &Pipeline{
		ctx: ctx,
	}
}

func (p *Pipeline) AddSteps(steps ...PipelineStep) {
	for _, step := range steps {
		reconciler := step(p.ctx)
		if reconciler == nil {
			return
		}
		p.steps = append(p.steps, reconciler)
	}
}

// Reconcile the steps
func (p *Pipeline) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	result := ctrl.Result{
		Requeue: false,
	}

	for _, step := range p.steps {
		ret, err := step.Reconcile(ctx, req)
		if err != nil {
			return result, err
		}

		result.Requeue = result.Requeue || ret.Requeue
		result.RequeueAfter = minDuration(result.RequeueAfter, ret.RequeueAfter)
	}

	return result, nil
}

func minDuration(d1, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	}
	return d2
}
