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
	"strings"

	"github.com/chaos-mesh/chaos-driver/pkg/client"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const chaosDaemonHelperCommand = "cdh"

const ioemPeriodUs = 10000

func (s *DaemonServer) ApplyBlockChaos(ctx context.Context, req *pb.ApplyBlockChaosRequest) (*pb.ApplyBlockChaosResponse, error) {
	volumeName, err := normalizeVolumeName(ctx, req.VolumePath)
	if err != nil {
		log.Error(err, "normalize volume name", "volumePath", req.VolumePath)
		return nil, err
	}

	err = enableIOEMElevator(volumeName)
	if err != nil {
		log.Error(err, "error while enabling ioem elevator", "volumeName", volumeName)
		return nil, errors.Wrapf(err, "enable ioem elevator for volume %s", volumeName)
	}

	volumePath := "/dev/" + volumeName
	if _, err := os.Stat(volumePath); err != nil {
		log.Error(err, "error while getting stat of volume", "volumePath", volumePath)
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

func normalizeVolumeName(ctx context.Context, volumePath string) (string, error) {
	volumeName, err := bpm.DefaultProcessBuilder(chaosDaemonHelperCommand, "normalize-volume-name", volumePath).SetContext(ctx).SetNS(1, bpm.PidNS).Build().Output()
	if err != nil {
		return "", errors.Wrapf(err, "normalize volume name %s", volumePath)
	}

	return strings.Trim(string(volumeName), "\n\x00"), nil
}

func enableIOEMElevator(volumeName string) error {
	schedulerPath := "/sys/block/" + volumeName + "/queue/scheduler"
	rawSchedulers, err := os.ReadFile(schedulerPath)
	if err != nil {
		return errors.Wrapf(err, "reading schedulers %s", schedulerPath)
	}

	schedulers := strings.Split(strings.Trim(string(rawSchedulers), " \x00\n"), " ")
	choosenScheduler := ""
	for _, scheduler := range schedulers {
		// TODO: record the current scheduler, and recover in the future
		// Bue it's hard to decide whether a block device is affected by chaos

		// The ioem scheduler without fault injection is only a simple FIFO,
		// which is nearly the same with none scheduler. But for HDD, the
		// none scheduler is much slower than other schedulers. For NVMe,
		// the default scheduler is none, and it's fine to keep ioem without
		// significant overhead.

		// The current scheduler should be surrounded by "[]"

		if strings.HasPrefix(scheduler, "ioem") {
			choosenScheduler = scheduler // it's either ioem or ioem-mq
		}
	}

	// it doesn't matter to pass any permission, because the file must exist
	err = os.WriteFile(schedulerPath, []byte(choosenScheduler), 0000)
	if err != nil {
		return errors.Wrapf(err, "writing %s to %s", choosenScheduler, schedulerPath)
	}

	return nil
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
