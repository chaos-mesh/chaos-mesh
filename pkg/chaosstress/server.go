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
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosstress/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	"github.com/golang/protobuf/ptypes/empty"
	uuid2 "github.com/google/uuid"
	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log = ctrl.Log.WithName("stress-server")
)

//go:generate protoc -I pb pb/chaosstress.proto --go_out=plugins=grpc:pb

type rpcServer struct {
	sync.Mutex
	tasks map[string]*exec.Cmd
}

func (r *rpcServer) ExecStressors(ctx context.Context,
	req *pb.StressRequest) (*pb.StressResponse, error) {
	log.Info("executing stressors", "request", req)
	raw, err := uuid2.NewUUID()
	if err != nil {
		return nil, err
	}
	uuid := raw.String()
	cmd := exec.Command("stress-ng", req.Stressors)
	log.Info("execute stressors", "stressors", req.Stressors)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	r.Lock()
	defer r.Unlock()
	r.tasks[uuid] = cmd
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Error(err, "stressors exit accidentally", "stressors", req.Stressors)
		}
		r.Lock()
		defer r.Unlock()
		delete(r.tasks, uuid)
	}()
	return &pb.StressResponse{Uuid: uuid}, nil
}

func (r *rpcServer) CancelStressors(ctx context.Context,
	req *pb.StressRequest) (*empty.Empty, error) {
	log.Info("canceling stressors", "request", req)
	if len(req.Uuid) == 0 {
		return nil, fmt.Errorf("missing chaos uuid")
	}
	if cmd, ok := r.tasks[req.Uuid]; ok {
		if err := cmd.Process.Kill(); err != nil {
			log.Error(err, "fail to exit stressors", "pid", cmd.Process.Pid)
			return nil, err
		}
	}
	return &empty.Empty{}, nil
}

// StartServer starts the stress server over the specified address
func StartServer(addr string) error {
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error(err, "failed to listen tcp server", "address", addr)
		os.Exit(1)
	}
	server := grpc.NewServer(grpc.UnaryInterceptor(
		utils.TimeoutServerInterceptor))
	pb.RegisterChaosStressServer(server,
		&rpcServer{
			tasks: make(map[string]*exec.Cmd),
		})
	return server.Serve(conn)
}
