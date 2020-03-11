package chaosdaemon

import (
	"context"

	"github.com/pingcap/chaos-mesh/pkg/time"

	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
)

func (s *daemonServer) SetTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	log.Info("Shift time", "Request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	childPids, err := GetChildProcesses(pid)
	if err != nil {
		log.Error(err, "fail to get child processes")
	}
	allPids := append(childPids, pid)
	log.Info("all related processes found", "pids", allPids)

	for _, pid := range allPids {
		err = time.ModifyTime(int(pid), req.Sec, req.Nsec, req.ClkIdsMask)
		if err != nil {
			log.Error(err, "error while modifying time", "pid", pid)
			return nil, err
		}
	}

	return &empty.Empty{}, nil
}

func (s *daemonServer) RecoverTimeOffset(ctx context.Context, req *pb.TimeRequest) (*empty.Empty, error) {
	log.Info("Recover time", "Request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	childPids, err := GetChildProcesses(pid)
	if err != nil {
		log.Error(err, "fail to get child processes")
	}
	allPids := append(childPids, pid)
	log.Info("get all related process pids", "pids", allPids)

	for _, pid := range allPids {
		// FIXME: if the process has halted and no process with this pid exists, we will get an error.
		err = time.ModifyTime(int(pid), int64(0), int64(0), 0)
		if err != nil {
			log.Error(err, "error while recovering", "pid", pid)
			return nil, err
		}
	}

	return &empty.Empty{}, nil
}
