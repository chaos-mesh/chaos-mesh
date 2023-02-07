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

package blockchaos

import (
	"context"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client
	Log logr.Logger

	decoder *utils.ContainerRecordDecoder
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("blockchaos apply", "record", records[index])

	_, _, volumePath, err := controller.ParseNamespacedNameContainerVolumePath(records[index].Id)
	if err != nil {
		return v1alpha1.NotInjected, errors.Wrapf(err, "parse container and volumePath %s", records[index].Id)
	}

	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index], obj)
	pbClient := decodedContainer.PbClient
	containerId := decodedContainer.ContainerId
	if pbClient != nil {
		defer pbClient.Close()
	}
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	blockchaos := obj.(*v1alpha1.BlockChaos)
	if blockchaos.Status.InjectionIds == nil {
		blockchaos.Status.InjectionIds = make(map[string]int)
	}
	_, ok := blockchaos.Status.InjectionIds[records[index].Id]
	if ok {
		impl.Log.Info("the blockchaos has already been injected")
		return v1alpha1.Injected, nil
	}

	var res *pb.ApplyBlockChaosResponse
	if blockchaos.Spec.Action == v1alpha1.BlockDelay {
		delay, err := time.ParseDuration(blockchaos.Spec.Delay.Latency)
		if err != nil {
			return v1alpha1.NotInjected, errors.Wrapf(err, "parse latency: %s", blockchaos.Spec.Delay.Latency)
		}

		corr, err := strconv.ParseFloat(blockchaos.Spec.Delay.Correlation, 64)
		if err != nil {
			return v1alpha1.NotInjected, errors.Wrapf(err, "parse corr: %s", blockchaos.Spec.Delay.Correlation)
		}

		jitter, err := time.ParseDuration(blockchaos.Spec.Delay.Jitter)
		if err != nil {
			return v1alpha1.NotInjected, errors.Wrapf(err, "parse jitter: %s", blockchaos.Spec.Delay.Jitter)
		}

		res, err = pbClient.ApplyBlockChaos(ctx, &pb.ApplyBlockChaosRequest{
			ContainerId: containerId,
			VolumePath:  volumePath,
			Action:      pb.ApplyBlockChaosRequest_Delay,
			Delay: &pb.BlockDelaySpec{
				Delay:       delay.Nanoseconds(),
				Correlation: corr,
				Jitter:      jitter.Nanoseconds(),
			},
			EnterNS: true,
		})

		if err != nil {
			return v1alpha1.NotInjected, err
		}
	} else {
		return v1alpha1.NotInjected, utils.ErrUnknownAction
	}

	blockchaos.Status.InjectionIds[records[index].Id] = int(res.InjectionId)

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("blockchaos recover", "record", records[index])

	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index], obj)
	pbClient := decodedContainer.PbClient
	if pbClient != nil {
		defer pbClient.Close()
	}
	if err != nil {
		if errors.Is(err, utils.ErrContainerNotFound) {
			// pretend the disappeared container has been recovered
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.Injected, err
	}

	blockchaos := obj.(*v1alpha1.BlockChaos)
	if blockchaos.Status.InjectionIds == nil {
		blockchaos.Status.InjectionIds = make(map[string]int)
	}
	injection_id, ok := blockchaos.Status.InjectionIds[records[index].Id]
	if !ok {
		impl.Log.Info("the blockchaos has already been recovered")
		return v1alpha1.NotInjected, nil
	}

	if _, err = pbClient.RecoverBlockChaos(ctx, &pb.RecoverBlockChaosRequest{
		InjectionId: int32(injection_id),
	}); err != nil {
		// TODO: check whether the error still exists
		return v1alpha1.Injected, err
	}
	delete(blockchaos.Status.InjectionIds, records[index].Id)
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger, decoder *utils.ContainerRecordDecoder) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:   "blockchaos",
		Object: &v1alpha1.BlockChaos{},
		Impl: &Impl{
			Client:  c,
			Log:     log.WithName("blockchaos"),
			decoder: decoder,
		},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
