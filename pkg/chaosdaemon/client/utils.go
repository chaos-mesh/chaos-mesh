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
	"google.golang.org/grpc"

	chaosdaemon "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
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

func New(conn *grpc.ClientConn) ChaosDaemonClientInterface {
	return &GrpcChaosDaemonClient{
		ChaosDaemonClient: chaosdaemon.NewChaosDaemonClient(conn),
		conn:              conn,
	}
}
