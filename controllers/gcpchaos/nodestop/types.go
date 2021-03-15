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
	"encoding/base64"
	"errors"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

const (
	GcpFinalizer = "gcp-Finalizer"
)

type endpoint struct {
	ctx.Context
}

// GetComputeService is used to get the GCP compute Service.
func GetComputeService(ctx context.Context, cli client.Client, gcpchaos *v1alpha1.GcpChaos) (*compute.Service, error) {
	if gcpchaos.Spec.SecretName != nil {
		secret := &v1.Secret{}
		err := cli.Get(ctx, types.NamespacedName{
			Name:      *gcpchaos.Spec.SecretName,
			Namespace: gcpchaos.Namespace,
		}, secret)
		if err != nil {
			return nil, err
		}

		decodeBytes, err := base64.StdEncoding.DecodeString(string(secret.Data["service_account"]))
		if err != nil {
			return nil, err
		}
		computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(decodeBytes))
		if err != nil {
			return nil, err
		}
		return computeService, nil
	} else {
		computeService, err := compute.NewService(ctx)
		if err != nil {
			return nil, err
		}
		return computeService, nil
	}
}

func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	gcpchaos, ok := chaos.(*v1alpha1.GcpChaos)
	if !ok {
		err := errors.New("chaos is not gcpchaos")
		e.Log.Error(err, "chaos is not GcpChaos", "chaos", chaos)
		return err
	}
	computeService, err := GetComputeService(ctx, e.Client, gcpchaos)
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
	computeService, err := GetComputeService(ctx, e.Client, gcpchaos)
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
