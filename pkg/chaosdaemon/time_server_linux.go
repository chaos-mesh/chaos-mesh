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
	"strconv"
	"sync"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/cerr"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
	"github.com/chaos-mesh/chaos-mesh/pkg/time"
)

type TimeChaosServer struct {
	// A mutation covers both the container's PID and the task manager's
	// shared task and configuration maps.
	mu                         sync.Mutex
	podContainerNameProcessMap tasks.PodContainerNameProcessMap
	manager                    tasks.TaskManager
	logger                     logr.Logger
}

func (s *TimeChaosServer) SetPodContainerNameProcess(idName tasks.PodContainerName, sysID tasks.SysPID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.podContainerNameProcessMap.Write(idName, sysID)
}

func (s *TimeChaosServer) DelPodContainerNameProcess(idName tasks.PodContainerName) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.podContainerNameProcessMap.Delete(idName)
}

func timeOffsetTaskID(uid tasks.TaskID, id tasks.PodContainerName) tasks.TaskID {
	// The controller sends the same UID for all containers selected in a pod.
	return strconv.Quote(string(id)) + uid
}

func (s *TimeChaosServer) SetTimeOffset(uid tasks.TaskID, id tasks.PodContainerName, config time.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.setTimeOffset(uid, id, config)
}

func (s *TimeChaosServer) setTimeOffset(uid tasks.TaskID, id tasks.PodContainerName, config time.Config) error {
	uid = timeOffsetTaskID(uid, id)
	if _, err := s.manager.GetConfigWithUID(uid); err == nil {
		return s.manager.Update(uid, id, &config)
	}

	paras := time.ConfigCreatorParas{
		Logger:        s.logger,
		Config:        config,
		PodProcessMap: &s.podContainerNameProcessMap,
	}

	err := s.manager.Create(uid, id, &config, paras)
	if errors.Cause(err) == cerr.ErrDuplicateEntity {
		return s.manager.Apply(uid, id, &config)
	}
	return err
}

func (s *DaemonServer) SetTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	logger := s.timeChaosServer.logger

	logger.Info("Shift time", "Request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		logger.Error(err, "error while getting IsID")
		return nil, err
	}

	target := tasks.PodContainerName(req.PodContainerName)
	timeServer := &s.timeChaosServer
	timeServer.mu.Lock()
	defer timeServer.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	timeServer.podContainerNameProcessMap.Write(target, tasks.SysPID(pid))
	err = timeServer.setTimeOffset(req.Uid, target, time.NewConfig(req.Sec, req.Nsec, req.ClkIdsMask))
	if err != nil {
		if len(timeServer.manager.GetUIDsWithPID(target)) == 0 {
			timeServer.podContainerNameProcessMap.Delete(target)
		}
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

	target := tasks.PodContainerName(req.PodContainerName)
	uid := timeOffsetTaskID(req.Uid, target)
	timeServer := &s.timeChaosServer
	timeServer.mu.Lock()
	defer timeServer.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := timeServer.manager.CheckTasks(uid, target); err != nil {
		return nil, err
	}
	timeServer.podContainerNameProcessMap.Write(target, tasks.SysPID(pid))
	err = timeServer.manager.Recover(uid, target)
	if err != nil {
		logger.Error(err, "error while recovering chaos")
		return nil, err
	}

	if len(timeServer.manager.GetUIDsWithPID(target)) == 0 {
		timeServer.podContainerNameProcessMap.Delete(target)
	}

	return &empty.Empty{}, nil
}
