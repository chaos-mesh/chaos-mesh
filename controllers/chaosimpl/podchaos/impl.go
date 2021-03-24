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

package podchaos

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	 "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos/containerkill"
	 "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos/podfailure"
	 "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos/podkill"
	"github.com/pkg/errors"
)

type Impl struct {
	podkill *podkill.Impl
	podfailure *podfailure.Impl
	containerkill *containerkill.Impl
}

func (i Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	podchaos := obj.(*v1alpha1.PodChaos)

	switch podchaos.Spec.Action {
	case v1alpha1.PodKillAction:
		return i.podkill.Apply(ctx, index, records, obj)
	case v1alpha1.PodFailureAction:
		return i.podfailure.Apply(ctx, index, records, obj)
	case v1alpha1.ContainerKillAction:
		return i.containerkill.Apply(ctx, index, records, obj)
	default:
		return v1alpha1.NotInjected, errors.Errorf("action %s is not expected", podchaos.Spec.Action)
	}
}

func (i Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	podchaos := obj.(*v1alpha1.PodChaos)

	switch podchaos.Spec.Action {
	case v1alpha1.PodKillAction:
		return i.podkill.Recover(ctx, index, records, obj)
	case v1alpha1.PodFailureAction:
		return i.podfailure.Recover(ctx, index, records, obj)
	case v1alpha1.ContainerKillAction:
		return i.containerkill.Recover(ctx, index, records, obj)
	default:
		return v1alpha1.NotInjected, errors.Errorf("action %s is not expected", podchaos.Spec.Action)
	}
}

func NewImpl(podkill *podkill.Impl, podfailure *podfailure.Impl, containerkill *containerkill.Impl) *Impl {
	return &Impl{
		podkill,
		podfailure,
		containerkill,
	}
}

