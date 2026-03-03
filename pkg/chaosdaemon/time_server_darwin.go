// Copyright 2022 Chaos Mesh Authors.
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

// Since TimeChaos is not unimplemented in darwin os. This file is only used for debugging, for example if your editor has gopls activated automatically.

package chaosdaemon

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
)

type TimeChaosServer struct {
	podContainerNameProcessMap tasks.PodContainerNameProcessMap
	manager                    tasks.TaskManager

	nameLocker tasks.LockMap[tasks.PodContainerName]
	logger     logr.Logger
}

func (s *TimeChaosServer) SetPodContainerNameProcess(idName tasks.PodContainerName, sysID tasks.SysPID) {
}

func (s *TimeChaosServer) DelPodContainerNameProcess(idName tasks.PodContainerName) {
}

func (s *TimeChaosServer) SetTimeOffset(uid tasks.TaskID, id tasks.PodContainerName, config interface{}) error {
	return nil
}

func (s *DaemonServer) SetTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (s *DaemonServer) RecoverTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
