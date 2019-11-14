// Copyright 2019 PingCAP, Inc.
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
	"os"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	log = ctrl.Log.WithValues("utils", "grpc")
)

func CreateGrpcConnection(ctx context.Context, c client.Client, pod *v1.Pod) (*grpc.ClientConn, error) {
	port := os.Getenv("CHAOS_DAEMON_PORT")
	if port == "" {
		port = "8080"
	}

	nodeName := pod.Spec.NodeName
	log.Info("Creating client to chaos-daemon", "node", nodeName)

	var node v1.Node
	err := c.Get(ctx, types.NamespacedName{
		Name: nodeName,
	}, &node)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", node.Status.Addresses[0].Address, port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
