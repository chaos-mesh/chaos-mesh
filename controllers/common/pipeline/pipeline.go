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

	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type Pipeline struct {
	object          *types.Object
	mgr             ctrl.Manager
	client          client.Client
	logger          logr.Logger
	recorderBuilder *recorder.RecorderBuilder

	steps []reconcile.Reconciler
}

type PipelineStep func(pipeline *Pipeline) reconcile.Reconciler

func NewPipeline(object *types.Object, mgr ctrl.Manager, client client.Client, logger logr.Logger, recorderBuilder *recorder.RecorderBuilder) *Pipeline {
	return &Pipeline{
		object: object, mgr: mgr, client: client, logger: logger, recorderBuilder: recorderBuilder,
	}
}

func (p *Pipeline) GetObject() *types.Object {
	return p.object
}

func (p *Pipeline) GetManager() ctrl.Manager {
	return p.mgr
}

func (p *Pipeline) GetClient() client.Client {
	return p.client
}

func (p *Pipeline) GetLogger() logr.Logger {
	return p.logger
}

func (p *Pipeline) GetRecordBuilder() *recorder.RecorderBuilder {
	return p.recorderBuilder
}

func (p *Pipeline) AddStep(step PipelineStep) {
	reconciler := step(p)
	if reconciler == nil {
		return
	}
	p.steps = append(p.steps, reconciler)
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
