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

package chaosdaemon

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
)

// ContainerKill kills container according to container id in the req
func (s *daemonServer) ContainerKill(ctx context.Context, req *pb.ContainerRequest) (*empty.Empty, error) {
	log.Info("container Kill","request", req)

	err := s.crClient.ContainerKillByContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while killing container")
		return nil, err
	}

	return &empty.Empty{}, nil
}
