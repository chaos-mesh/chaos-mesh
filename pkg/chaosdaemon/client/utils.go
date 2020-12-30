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

package client

import (
	"context"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chaosdaemon "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
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

	cc, err := grpcUtils.CreateGrpcConnection(ctx, c, pod, port)
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

	cc, err := grpcUtils.CreateGrpcConnectionWithAddress("localhost", port)
	if err != nil {
		return nil, err
	}
	return &GrpcChaosDaemonClient{
		ChaosDaemonClient: chaosdaemon.NewChaosDaemonClient(cc),
		conn:              cc,
	}, nil
}
