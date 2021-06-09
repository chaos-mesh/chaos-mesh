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

package containerkill

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

type Impl struct {
	client.Client

	Log logr.Logger

	decoder *utils.ContianerRecordDecoder
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index])
	pbClient := decodedContainer.PbClient
	containerId := decodedContainer.ContainerId
	if pbClient != nil {
		defer pbClient.Close()
	}
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	if _, err = pbClient.ContainerKill(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_KILL,
		},
		ContainerId: containerId,
	}); err != nil {
		impl.Log.Error(err, "kill container error", "containerID", containerId)
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger, decoder *utils.ContianerRecordDecoder) *Impl {
	return &Impl{
		Client:  c,
		Log:     log.WithName("containerkill"),
		decoder: decoder,
	}
}
