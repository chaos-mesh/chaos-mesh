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
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/cerr"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
	"github.com/chaos-mesh/chaos-mesh/pkg/time"
)

type TimeChaosServer struct {
	podProcessMap tasks.PodProcessMap
	manager       tasks.TaskManager
	logger        logr.Logger
}

func (s *TimeChaosServer) SetPodProcess(podID tasks.PodID, sysID tasks.SysPID) {
	s.podProcessMap.Write(podID, sysID)
}

func (s *TimeChaosServer) SetTimeOffset(uid tasks.TaskID, pid tasks.IsID, config time.Config) error {
	paras := time.ConfigCreatorParas{
		Logger:        s.logger,
		Config:        config,
		PodProcessMap: &s.podProcessMap,
	}

	// We assume the base time skew is not sensitive with process changes which
	// means time skew will not return error when the task target pod changes container id & IsID.
	// We assume controller will never update tasks.
	// According to the above, we do not handle error from s.manager.Apply like
	// ErrDuplicateEntity(task TaskID).
	err := s.manager.Create(uid, pid, &config, paras)
	if err != nil {
		if errors.Cause(err) == cerr.ErrDuplicateEntity {
			err := s.manager.Apply(uid, pid, &config)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (s *DaemonServer) SetTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	logger := s.timeChaosServer.logger

	logger.Info("Shift time", "Request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		logger.Error(err, "error while getting IsID")
		return nil, err
	}

	s.timeChaosServer.SetPodProcess(tasks.PodID(req.PodId), tasks.SysPID(pid))
	err = s.timeChaosServer.SetTimeOffset(req.Uid, tasks.PodID(req.PodId),
		time.NewConfig(req.Sec, req.Nsec, req.ClkIdsMask))
	if err != nil {
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

	s.timeChaosServer.SetPodProcess(tasks.PodID(req.PodId), tasks.SysPID(pid))

	err = s.timeChaosServer.manager.Recover(req.Uid, tasks.PodID(req.PodId))
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}
