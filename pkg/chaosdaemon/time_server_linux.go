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
	podContainerNameProcessMap tasks.PodContainerNameProcessMap
	manager                    tasks.TaskManager

	nameLocker tasks.LockMap[tasks.PodContainerName]
	logger     logr.Logger
}

func (s *TimeChaosServer) SetPodContainerNameProcess(idName tasks.PodContainerName, sysID tasks.SysPID) {
	s.podContainerNameProcessMap.Write(idName, sysID)
}

func (s *TimeChaosServer) DelPodContainerNameProcess(idName tasks.PodContainerName) {
	s.podContainerNameProcessMap.Delete(idName)
}

func (s *TimeChaosServer) SetTimeOffset(uid tasks.TaskID, id tasks.PodContainerName, config time.Config) error {
	paras := time.ConfigCreatorParas{
		Logger:        s.logger,
		Config:        config,
		PodProcessMap: &s.podContainerNameProcessMap,
	}

	unlock := s.nameLocker.Lock(id)
	defer unlock()
	// We assume the base time skew is not sensitive with process changes which
	// means time skew will not return error when the task target pod changes container id & IsID.
	// We assume controller will never update tasks.
	// According to the above, we do not handle error from s.manager.Apply like
	// ErrDuplicateEntity(task TaskID).
	err := s.manager.Create(uid, id, &config, paras)
	if err != nil {
		if errors.Cause(err) == cerr.ErrDuplicateEntity {
			err := s.manager.Apply(uid, id, &config)
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

	s.timeChaosServer.SetPodContainerNameProcess(tasks.PodContainerName(req.PodContainerName), tasks.SysPID(pid))
	err = s.timeChaosServer.SetTimeOffset(req.Uid, tasks.PodContainerName(req.PodContainerName),
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

	nameID := tasks.PodContainerName(req.PodContainerName)

	s.timeChaosServer.SetPodContainerNameProcess(nameID, tasks.SysPID(pid))

	unlock := s.timeChaosServer.nameLocker.Lock(nameID)
	defer unlock()

	err = s.timeChaosServer.manager.Recover(req.Uid, nameID)
	if err != nil {
		logger.Error(err, "error while recovering chaos")
		return nil, err
	}

	if len(s.timeChaosServer.manager.GetUIDsWithPID(nameID)) == 0 {
		s.timeChaosServer.DelPodContainerNameProcess(nameID)
		s.timeChaosServer.nameLocker.Del(nameID)
	}

	return &empty.Empty{}, nil
}
