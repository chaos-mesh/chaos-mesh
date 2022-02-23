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

package chaosdaemon

import (
	"context"
	"os"

	"github.com/chaos-mesh/chaos-driver/pkg/client"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const ioemPeriodUs = 10000

func (s *DaemonServer) ApplyBlockChaos(ctx context.Context, req *pb.ApplyBlockChaosRequest) (*pb.ApplyBlockChaosResponse, error) {
	volumeName, err := normalizeVolumeName(req.VolumePath)
	if err != nil {
		log.Error(err, "normalize volume name", "volumePath", req.VolumePath)
		return nil, err
	}

	// TODO: automatically modify the elevator to the `ioem` or `ioem-mq`

	volumePath := "/dev/" + volumeName
	if _, err := os.Stat(volumePath); err != nil {
		return nil, errors.Wrapf(err, "volume path %s does not exist", volumePath)
	}

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	c, err := client.New()
	if err != nil {
		log.Error(err, "create chaos-driver client")
		return nil, err
	}

	if req.Action == pb.ApplyBlockChaosRequest_Limit {
		// 1e6 / period_us * quota = IOPS, which means
		// quota = IOPS * period_us / 1e6
		quota := uint64(req.Iops * ioemPeriodUs / 1e6)
		id, err := c.InjectIOEMLimit(volumePath, 0, uint(pid), ioemPeriodUs, quota)
		if err != nil {
			log.Error(err, "inject ioem limit")
			return nil, err
		}
		return &pb.ApplyBlockChaosResponse{
			InjectionId: int32(id),
		}, nil
	} else if req.Action == pb.ApplyBlockChaosRequest_Delay {
		id, err := c.InjectIOEMDelay(volumePath, 0, uint(pid), int64(req.Delay.Delay), int64(req.Delay.Jitter), uint32(req.Delay.Correlation))
		if err != nil {
			log.Error(err, "inject ioem delay")
			return nil, err
		}
		return &pb.ApplyBlockChaosResponse{
			InjectionId: int32(id),
		}, nil
	} else {
		return nil, errors.New("unknown action")
	}
}

func (s *DaemonServer) RecoverBlockChaos(ctx context.Context, req *pb.RecoverBlockChaosRequest) (*empty.Empty, error) {
	c, err := client.New()
	if err != nil {
		log.Error(err, "create chaos-driver client")
		return nil, err
	}

	err = c.Recover(int(req.InjectionId))
	if err != nil {
		log.Error(err, "recover injection", "id", req.InjectionId)
		return nil, err
	}

	return &empty.Empty{}, nil
}
