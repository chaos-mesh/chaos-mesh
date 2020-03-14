// Copyright 2020 PingCAP, Inc.
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

package chaosstress

import (
	"fmt"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosstress/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
)

// ClientInterface represents the ChaosStressClient, it's used to simply unit test
type ClientInterface interface {
	pb.ChaosStressClient
	Close() error
}

// GrpcChaosStressClient would act like ChaosStressClient with a Close method
type GrpcChaosStressClient struct {
	pb.ChaosStressClient
	conn *grpc.ClientConn
}

// Close closes the client
func (c *GrpcChaosStressClient) Close() error {
	return c.conn.Close()
}

func createGrpcConnection(pod *v1.Pod, port string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", pod.Status.PodIP, port),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(utils.TimeoutClientInterceptor))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// NewGrpcChaosStressClient would create a ChaosStressClient
func NewGrpcChaosStressClient(pod *v1.Pod, port string) (ClientInterface, error) {
	if cli := mock.On("MockChaosStressClient"); cli != nil {
		return cli.(ClientInterface), nil
	}
	if err := mock.On("NewChaosStressClientError"); err != nil {
		return nil, err.(error)
	}

	cc, err := createGrpcConnection(pod, port)
	if err != nil {
		return nil, err
	}
	return &GrpcChaosStressClient{
		ChaosStressClient: pb.NewChaosStressClient(cc),
		conn:              cc,
	}, nil
}
