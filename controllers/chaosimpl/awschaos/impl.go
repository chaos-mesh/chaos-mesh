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

package awschaos

import (
	"context"

	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/awschaos/detachvolume"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/awschaos/ec2restart"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/awschaos/ec2stop"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type Impl struct {
	detachvolume *detachvolume.Impl
	ec2restart   *ec2restart.Impl
	ec2stop      *ec2stop.Impl
}

func (i Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	awschaos := obj.(*v1alpha1.AwsChaos)

	switch awschaos.Spec.Action {
	case v1alpha1.DetachVolume:
		return i.detachvolume.Apply(ctx, index, records, obj)
	case v1alpha1.Ec2Restart:
		return i.ec2restart.Apply(ctx, index, records, obj)
	case v1alpha1.Ec2Stop:
		return i.ec2stop.Apply(ctx, index, records, obj)
	default:
		return v1alpha1.NotInjected, errors.Errorf("action %s is not expected", awschaos.Spec.Action)
	}
}

func (i Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	awschaos := obj.(*v1alpha1.AwsChaos)

	switch awschaos.Spec.Action {
	case v1alpha1.DetachVolume:
		return i.detachvolume.Recover(ctx, index, records, obj)
	case v1alpha1.Ec2Restart:
		return i.ec2restart.Recover(ctx, index, records, obj)
	case v1alpha1.Ec2Stop:
		return i.ec2stop.Recover(ctx, index, records, obj)
	default:
		return v1alpha1.NotInjected, errors.Errorf("action %s is not expected", awschaos.Spec.Action)
	}
}

func NewImpl(detachvolume *detachvolume.Impl, ec2restart *ec2restart.Impl, ec2stop *ec2stop.Impl) *common.ChaosImplPair {
	return &common.ChaosImplPair{
		Name:   "awschaos",
		Object: &v1alpha1.AwsChaos{},
		Impl: &Impl{
			detachvolume,
			ec2restart,
			ec2stop,
		},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
	detachvolume.NewImpl,
	ec2restart.NewImpl,
	ec2stop.NewImpl)
