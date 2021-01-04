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

package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/controllers/config"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DefaultRPCTimeout specifies default timeout of RPC between controller and chaos-operator
const DefaultRPCTimeout = 60 * time.Second

// RPCTimeout specifies timeout of RPC between controller and chaos-operator
var RPCTimeout = DefaultRPCTimeout

var log = ctrl.Log.WithName("util")

// CreateGrpcConnection create a grpc connection with given port
func CreateGrpcConnection(ctx context.Context, c client.Client, pod *v1.Pod, port int) (*grpc.ClientConn, error) {
	nodeName := pod.Spec.NodeName
	log.Info("Creating client to chaos-daemon", "node", nodeName)

	ns := config.ControllerCfg.Namespace
	var endpoints v1.Endpoints
	err := c.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      "chaos-daemon",
	}, &endpoints)
	if err != nil {
		return nil, err
	}

	daemonIP := findIPOnEndpoints(&endpoints, nodeName)
	if len(daemonIP) == 0 {
		return nil, errors.Errorf("cannot find daemonIP on node %s in related Endpoints %v", nodeName, endpoints)
	}
	return CreateGrpcConnectionWithAddress(daemonIP, port)
}

// CreateGrpcConnectionWithAddress create a grpc connection with given port and address
func CreateGrpcConnectionWithAddress(address string, port int) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", address, port),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(TimeoutClientInterceptor))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func findIPOnEndpoints(e *v1.Endpoints, nodeName string) string {
	for _, subset := range e.Subsets {
		for _, addr := range subset.Addresses {
			if addr.NodeName != nil && *addr.NodeName == nodeName {
				return addr.IP
			}
		}
	}

	return ""
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
