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

package nodestop

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	gcp "github.com/chaos-mesh/chaos-mesh/controllers/gcpchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

const (
	GcpFinalizer = "gcp-finalizer"
)

type endpoint struct {
	ctx.Context
}

func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	gcpchaos, ok := chaos.(*v1alpha1.GcpChaos)
	if !ok {
		err := errors.New("chaos is not gcpchaos")
		e.Log.Error(err, "chaos is not GcpChaos", "chaos", chaos)
		return err
	}
	computeService, err := gcp.GetComputeService(ctx, e.Client, gcpchaos)
	if err != nil {
		e.Log.Error(err, "fail to get the compute service")
		return err
	}
	gcpchaos.Finalizers = []string{GcpFinalizer}
	_, err = computeService.Instances.Stop(gcpchaos.Spec.Project, gcpchaos.Spec.Zone, gcpchaos.Spec.Instance).Do()
	if err != nil {
		gcpchaos.Finalizers = make([]string, 0)
		e.Log.Error(err, "fail to stop the instance")
		return err
	}

	return nil
}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	gcpchaos, ok := chaos.(*v1alpha1.GcpChaos)
	if !ok {
		err := errors.New("chaos is not gcpchaos")
		e.Log.Error(err, "chaos is not GcpChaos", "chaos", chaos)
		return err
	}
	gcpchaos.Finalizers = make([]string, 0)
	computeService, err := gcp.GetComputeService(ctx, e.Client, gcpchaos)
	if err != nil {
		e.Log.Error(err, "fail to get the compute service")
		return err
	}
	_, err = computeService.Instances.Start(gcpchaos.Spec.Project, gcpchaos.Spec.Zone, gcpchaos.Spec.Instance).Do()
	if err != nil {
		e.Log.Error(err, "fail to stop the instance")
		return err
	}
	return nil
}

func (e *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.GcpChaos{}
}

func init() {
	router.Register("gcpchaos", &v1alpha1.GcpChaos{}, func(obj runtime.Object) bool {
		chaos, ok := obj.(*v1alpha1.GcpChaos)
		if !ok {
			return false
		}

		return chaos.Spec.Action == v1alpha1.NodeStop
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
