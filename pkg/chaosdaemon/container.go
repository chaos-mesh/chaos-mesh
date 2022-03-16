// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package chaosdaemon

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

// ContainerKill kills container according to container id in the req
func (s *DaemonServer) ContainerKill(ctx context.Context, req *pb.ContainerRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)

	log.Info("Container Kill", "request", req)

	action := req.Action.Action
	if action != pb.ContainerAction_KILL {
		err := errors.Errorf("container action is %s , not kill", action)
		log.Error(err, "container action is not expected")
		return nil, err
	}

	err := s.crClient.ContainerKillByContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while killing container")
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *DaemonServer) ContainerGetPid(ctx context.Context, req *pb.ContainerRequest) (*pb.ContainerResponse, error) {
	log := s.getLoggerFromContext(ctx)

	log.Info("container GetPid", "request", req)

	action := req.Action.Action
	if action != pb.ContainerAction_GETPID {
		err := errors.Errorf("container action is %s , not getpid", action)
		log.Error(err, "container action is not expected")
		return nil, err
	}

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting pid from container")
		return nil, err
	}

	return &pb.ContainerResponse{Pid: pid}, nil
}
