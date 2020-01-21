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

package chaosdaemon

import (
	"fmt"
	"net"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("chaos-daemon-server")

//go:generate protoc -I pb pb/chaosdaemon.proto --go_out=plugins=grpc:pb

// Server represents an HTTP server for tc daemon
type Server struct {
	crClient ContainerRuntimeInfoClient
}

func newServer(containerRuntime string) (*Server, error) {
	crClient, err := CreateContainerRuntimeInfoClient(containerRuntime)
	if err != nil {
		return nil, err
	}

	return &Server{
		crClient: crClient,
	}, nil
}

// StartServer starts chaos-daemon.
func StartServer(host string, port int, containerRuntime string) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Error(err, "failed to listen")
		return
	}

	s := grpc.NewServer()
	chaosDaemonServer, err := newServer(containerRuntime)
	if err != nil {
		log.Error(err, "failed to create server")
		return
	}
	pb.RegisterChaosDaemonServer(s, chaosDaemonServer)

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Error(err, "failed to serve")
		return
	}
}
