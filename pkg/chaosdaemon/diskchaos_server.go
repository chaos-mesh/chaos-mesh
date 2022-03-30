package chaosdaemon

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/disk"
	"github.com/golang/protobuf/ptypes/empty"
	"sync"
)

type DiskChaosManager struct {
	FillMap    SyncMap[*disk.Fill]
	PayloadMap SyncMap[*disk.Payload]

	locker sync.RWMutex
}

type SyncMap[T any] struct {
	inner  map[string]T
	locker sync.RWMutex
}

func NewSyncMap[T any](inner map[string]T) SyncMap[T] {
	return SyncMap[T]{
		inner:  inner,
		locker: sync.RWMutex{},
	}
}

func (m *SyncMap[T]) Read(key string) T {
	m.locker.RLock()
	defer m.locker.RUnlock()
	return m.inner[key]
}

func (m *SyncMap[T]) Write(key string, value T) {
	m.locker.Lock()
	defer m.locker.Unlock()
	m.inner[key] = value
}

func (s *DaemonServer) DiskFill(ctx context.Context, req *pb.DiskFillRequest) (*empty.Empty, error) {
	logger := s.rootLogger.WithName("Disk Fill")
	f, err := disk.InitFill(
		disk.NewFillConfig(req.FillByFallocate, disk.CommonConfig{
			Path:    req.Path,
			Size:    req.Size,
			Percent: req.Percent,
			SLock:   disk.NewSpaceLock(req.SpaceLockSize),
		}), logger)
	if err != nil {
		return nil, err
	}

	s.DiskChaosManager.FillMap.Write(req.ContainerId, f)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		return nil, err
	}

	err = f.Inject(pid)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *DaemonServer) RecoverDiskFill(ctx context.Context, req *pb.DiskFillRecoverRequest) (*empty.Empty, error) {
	logger := s.rootLogger.WithName("Recover Disk Fill")
	f := s.DiskChaosManager.FillMap.Read(req.ContainerId)

	if f != nil {
		err := f.Recover()
		if err != nil {
			logger.Error(err, "recover meet with error")
		}
	} else {
		logger.Info("recover nil")
	}

	return &empty.Empty{}, nil
}

func (s *DaemonServer) DiskPayload(ctx context.Context, req *pb.DiskPayloadRequest) (*empty.Empty, error) {
	logger := s.rootLogger.WithName("Disk Payload")
	p, err := disk.InitPayload(
		disk.NewPayloadConfig(disk.PayloadAction(req.Action),
			disk.CommonConfig{
				Path:    req.Path,
				Size:    req.Size,
				Percent: req.Percent,
				SLock:   disk.NewSpaceLock(req.SpaceLockSize),
			}, disk.RuntimeConfig{
				ProcessNum:    uint8(req.ProcessNum),
				LoopExecution: false,
			}), logger)
	if err != nil {
		return nil, err
	}

	s.DiskChaosManager.PayloadMap.Write(req.ContainerId, p)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		return nil, err
	}

	err = p.Inject(pid)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *DaemonServer) RecoverDiskPayload(ctx context.Context, req *pb.DiskPayloadRecoverRequest) (*empty.Empty, error) {
	logger := s.rootLogger.WithName("Recover Disk Payload")
	p := s.DiskChaosManager.PayloadMap.Read(req.ContainerId)

	if p != nil {
		err := p.Recover()
		if err != nil {
			logger.Error(err, "recover meet with error")
		}
	} else {
		logger.Info("recover nil")
	}

	return &empty.Empty{}, nil
}
