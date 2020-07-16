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
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func (s *daemonServer) SetNetem(ctx context.Context, in *pb.NetemRequest) (*empty.Empty, error) {
	log.Info("Set netem", "Request", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	if err := applyNetem(in.Netem, pid); err != nil {
		return nil, status.Errorf(codes.Internal, "netem apply error: %v", err)
	}

	return &empty.Empty{}, nil
}

func (s *daemonServer) DeleteNetem(ctx context.Context, in *pb.NetemRequest) (*empty.Empty, error) {
	log.Info("Delete netem", "Request", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	if err := deleteNetem(in.Netem, pid); err != nil {
		return nil, status.Errorf(codes.Internal, "netem cancel error: %v", err)
	}

	return &empty.Empty{}, nil
}
