// Copyright 2020 Chaos Mesh Authors.
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

package utils

import (
	"context"
	"math"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chaosdaemon "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// ChaosDaemonClientInterface represents the ChaosDaemonClient, it's used to simply unit test
type ChaosDaemonClientInterface interface {
	chaosdaemon.ChaosDaemonClient
	Close() error
}

// GrpcChaosDaemonClient would act like chaosdaemon.ChaosDaemonClient with a Close method
type GrpcChaosDaemonClient struct {
	chaosdaemon.ChaosDaemonClient
	conn *grpc.ClientConn
}

func (c *GrpcChaosDaemonClient) Close() error {
	return c.conn.Close()
}

// NewChaosDaemonClient would create ChaosDaemonClient
func NewChaosDaemonClient(ctx context.Context, c client.Client, pod *v1.Pod, port int) (ChaosDaemonClientInterface, error) {
	if cli := mock.On("MockChaosDaemonClient"); cli != nil {
		return cli.(ChaosDaemonClientInterface), nil
	}
	if err := mock.On("NewChaosDaemonClientError"); err != nil {
		return nil, err.(error)
	}

	cc, err := CreateGrpcConnection(ctx, c, pod, port)
	if err != nil {
		return nil, err
	}
	return &GrpcChaosDaemonClient{
		ChaosDaemonClient: chaosdaemon.NewChaosDaemonClient(cc),
		conn:              cc,
	}, nil
}

// NewChaosDaemonClientLocally would create ChaosDaemonClient in localhost
func NewChaosDaemonClientLocally(port int) (ChaosDaemonClientInterface, error) {
	if cli := mock.On("MockChaosDaemonClient"); cli != nil {
		return cli.(ChaosDaemonClientInterface), nil
	}
	if err := mock.On("NewChaosDaemonClientError"); err != nil {
		return nil, err.(error)
	}

	cc, err := CreateGrpcConnectionWithAddress(port, "localhost")
	if err != nil {
		return nil, err
	}
	return &GrpcChaosDaemonClient{
		ChaosDaemonClient: chaosdaemon.NewChaosDaemonClient(cc),
		conn:              cc,
	}, nil
}

// MergeNetem merges two Netem protos into a new one.
// REMEMBER to assign the return value, i.e. merged = utils.MergeNetm(merged, em)
// For each field it takes the bigger value of the two.
// Its main use case is merging netem of different types, e.g. delay and loss.
// It returns nil if both inputs are nil.
// Otherwise it returns a new Netem with merged values.
func MergeNetem(a, b *chaosdaemon.Netem) *chaosdaemon.Netem {
	if a == nil && b == nil {
		return nil
	}
	// NOTE: because proto getters check nil, we are good here even if one of them is nil.
	// But we just assign empty value to make IDE and linters happy.
	if a == nil {
		a = &chaosdaemon.Netem{}
	}
	if b == nil {
		b = &chaosdaemon.Netem{}
	}
	return &chaosdaemon.Netem{
		Time:          maxu32(a.GetTime(), b.GetTime()),
		Jitter:        maxu32(a.GetJitter(), b.GetJitter()),
		DelayCorr:     maxf32(a.GetDelayCorr(), b.GetDelayCorr()),
		Limit:         maxu32(a.GetLimit(), b.GetLimit()),
		Loss:          maxf32(a.GetLoss(), b.GetLoss()),
		LossCorr:      maxf32(a.GetLossCorr(), b.GetLossCorr()),
		Gap:           maxu32(a.GetGap(), b.GetGap()),
		Duplicate:     maxf32(a.GetDuplicate(), b.GetDuplicate()),
		DuplicateCorr: maxf32(a.GetDuplicateCorr(), b.GetDuplicateCorr()),
		Reorder:       maxf32(a.GetReorder(), b.GetReorder()),
		ReorderCorr:   maxf32(a.GetReorderCorr(), b.GetReorderCorr()),
		Corrupt:       maxf32(a.GetCorrupt(), b.GetCorrupt()),
		CorruptCorr:   maxf32(a.GetCorruptCorr(), b.GetCorruptCorr()),
	}
}

func maxu32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func maxf32(a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}
