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

package podkill

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
)

type Impl struct {
	client.Client
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	podchaos := obj.(*v1alpha1.PodChaos)

	var pod v1.Pod
	err := impl.Get(ctx, controller.ParseNamespacedName(records[index].Id), &pod)
	if err != nil {
		// TODO: handle this error
		return v1alpha1.NotInjected, err
	}

	err = impl.Delete(ctx, &pod, &client.DeleteOptions{
		GracePeriodSeconds: &podchaos.Spec.GracePeriod, // PeriodSeconds has to be set specifically
	})
	if err != nil {
		// TODO: handle this error
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client) *Impl {
	return &Impl{
		Client: c,
	}
}
