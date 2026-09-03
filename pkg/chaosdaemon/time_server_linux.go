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

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/time"
)

type TimeChaosServer struct {
	manager *time.PersistentTimeChaos
	logger  logr.Logger
}

func newTimeChaosServer(logger logr.Logger) TimeChaosServer {
	return TimeChaosServer{
		manager: time.NewPersistentTimeChaos("/host-run/chaos-daemon/timechaos", logger),
		logger:  logger,
	}
}

func (s *DaemonServer) SetTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	logger := s.timeChaosServer.logger

	logger.Info("Shift time", "Request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		logger.Error(err, "error while getting IsID")
		return nil, err
	}

	err = s.timeChaosServer.manager.Apply(req.Uid, req.PodContainerName, req.ContainerId, int(pid),
		time.NewConfig(req.Sec, req.Nsec, req.ClkIdsMask))
	if err != nil {
		logger.Error(err, "error while applying chaos")
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *DaemonServer) RecoverTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	logger := s.timeChaosServer.logger

	logger.Info("Recover time", "Request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		logger.Error(err, "error while getting IsID")
		return nil, err
	}

	if err := s.timeChaosServer.manager.Recover(req.Uid, req.PodContainerName, req.ContainerId, int(pid)); err != nil {
		logger.Error(err, "error while recovering chaos")
		return nil, err
	}

	return &empty.Empty{}, nil
}
