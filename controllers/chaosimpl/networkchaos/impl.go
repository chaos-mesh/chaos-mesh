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

package networkchaos

import (
	"context"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/partition"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/trafficcontrol"
)

type Impl struct {
	trafficcontrol *trafficcontrol.Impl
	partition      *partition.Impl
}

func (i Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	networkchaos := obj.(*v1alpha1.NetworkChaos)

	switch networkchaos.Spec.Action {
	case v1alpha1.BandwidthAction, v1alpha1.NetemAction, v1alpha1.DelayAction, v1alpha1.LossAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction:
		return i.trafficcontrol.Apply(ctx, index, records, obj)
	case v1alpha1.PartitionAction:
		return i.partition.Apply(ctx, index, records, obj)
	default:
		return v1alpha1.NotInjected, errors.Errorf("action %s is not expected", networkchaos.Spec.Action)
	}
}

func (i Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	networkchaos := obj.(*v1alpha1.NetworkChaos)

	switch networkchaos.Spec.Action {
	case v1alpha1.BandwidthAction, v1alpha1.NetemAction, v1alpha1.DelayAction, v1alpha1.LossAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction:
		return i.trafficcontrol.Recover(ctx, index, records, obj)
	case v1alpha1.PartitionAction:
		return i.partition.Recover(ctx, index, records, obj)
	default:
		return v1alpha1.NotInjected, errors.Errorf("action %s is not expected", networkchaos.Spec.Action)
	}
}

func NewImpl(trafficcontrol *trafficcontrol.Impl, partition *partition.Impl) *Impl {
	return &Impl{
		trafficcontrol,
		partition,
	}
}
