// Copyright 2019 Chaos Mesh Authors.
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
	"fmt"
	"time"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DefaultRPCTimeout specifies default timeout of RPC between controller and chaos-operator
const DefaultRPCTimeout = 60 * time.Second

// RPCTimeout specifies timeout of RPC between controller and chaos-operator
var RPCTimeout = DefaultRPCTimeout

// CreateGrpcConnection create a grpc connection with given port
func CreateGrpcConnection(ctx context.Context, c client.Client, pod *v1.Pod, port int) (*grpc.ClientConn, error) {
	nodeName := pod.Spec.NodeName
	log.Info("Creating client to chaos-daemon", "node", nodeName)

	var node v1.Node
	err := c.Get(ctx, types.NamespacedName{
		Name: nodeName,
	}, &node)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.Status.Addresses[0].Address, port),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(TimeoutClientInterceptor))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// TimeoutClientInterceptor wraps the RPC with a timeout.
func TimeoutClientInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, RPCTimeout)
	defer cancel()
	return invoker(ctx, method, req, reply, cc, opts...)
}

// TimeoutServerInterceptor ensures the context is intact before handling over the
// request to application.
func TimeoutServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return handler(ctx, req)
}
