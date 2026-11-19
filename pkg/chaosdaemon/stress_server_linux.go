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
	"math"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/mem"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/cgroups"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

func (s *DaemonServer) ExecStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*pb.ExecStressResponse, error) {
	log := s.getLoggerFromContext(ctx)
	log.Info("Executing stressors", "request", req)

	// cpuStressors
	cpuProc, err := s.ExecCPUStressors(ctx, req)
	if err != nil {
		// this error would be resolved by grpc server framework, it's the top level to print it.
		s.rootLogger.Error(err, "exec cpu stressors", "containerID", req.Target, "cpuStressors", req.CpuStressors)
		return nil, err
	}

	// memoryStressor
	memoryProc, err := s.ExecMemoryStressors(ctx, req)
	if err != nil {
		if cpuProc != nil {
			if killErr := s.backgroundProcessManager.KillBackgroundProcess(ctx, cpuProc.UID); killErr != nil {
				log.Error(killErr, "stop CPU stressor after memory stress failed", "uid", cpuProc.UID)
			}
		}
		return nil, err
	}

	resp := new(pb.ExecStressResponse)
	if cpuProc != nil {
		resp.CpuInstance = strconv.Itoa(cpuProc.Pair.Pid)
		resp.CpuStartTime = cpuProc.Pair.CreateTime
		resp.CpuInstanceUid = cpuProc.UID
	}
	if memoryProc != nil {
		resp.MemoryInstance = strconv.Itoa(memoryProc.Pair.Pid)
		resp.MemoryStartTime = memoryProc.Pair.CreateTime
		resp.MemoryInstanceUid = memoryProc.UID
	}

	return resp, nil
}

func (s *DaemonServer) CancelStressors(ctx context.Context,
	req *pb.CancelStressRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)
	CPUPID, err := strconv.Atoi(req.CpuInstance)
	if req.CpuInstance != "" && err != nil {
		return nil, err
	}

	MemoryPid, err := strconv.Atoi(req.MemoryInstance)
	if req.MemoryInstance != "" && err != nil {
		return nil, err
	}

	if req.CpuInstanceUid == "" && CPUPID != 0 {
		if uid, ok := s.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: CPUPID, CreateTime: req.CpuStartTime}); ok {
			req.CpuInstanceUid = uid
		}
	}

	if req.MemoryInstanceUid == "" && MemoryPid != 0 {
		if uid, ok := s.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: MemoryPid, CreateTime: req.MemoryStartTime}); ok {
			req.MemoryInstanceUid = uid
		}
	}

	log.Info("Canceling stressors", "request", req)

	if req.CpuInstanceUid != "" {
		err = s.backgroundProcessManager.KillBackgroundProcess(ctx, req.CpuInstanceUid)
		if err != nil {
			return nil, err
		}
	}

	if req.MemoryInstanceUid != "" {
		err = s.backgroundProcessManager.KillBackgroundProcess(ctx, req.MemoryInstanceUid)
		if err != nil {
			return nil, err
		}
	}

	log.Info("killing stressor successfully")
	return &empty.Empty{}, nil
}

func (s *DaemonServer) ExecCPUStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*bpm.Process, error) {
	log := s.getLoggerFromContext(ctx)
	if req.CpuStressors == "" {
		return nil, nil
	}
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return nil, err
	}

	attachCGroup, err := cgroups.GetAttacherForPID(int(pid))
	if err != nil {
		return nil, err
	}

	processBuilder := bpm.DefaultProcessBuilder("stress-ng", strings.Fields(req.CpuStressors)...).
		EnablePause()
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS)
	}
	cmd := processBuilder.Build(ctx)

	proc, err := s.backgroundProcessManager.StartProcess(ctx, cmd)
	if err != nil {
		return nil, err
	}
	log.Info("Start stress-ng successfully", "command", cmd, "pid", proc.Pair.Pid, "uid", proc.UID)

	if err = attachCGroup.AttachProcess(proc.Pair.Pid); err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.Error(kerr, "kill stress-ng failed", "request", req)
		}
		return nil, err
	}

	for {
		// TODO: find a better way to resume pause process
		if err := cmd.Process.Signal(syscall.SIGCONT); err != nil {
			return nil, err
		}

		log.Info("send signal to resume process")
		time.Sleep(time.Millisecond)

		// TODO: integrate the resume process into the bpm
		comm, err := util.ReadCommName(cmd.Process.Pid)
		if err != nil {
			return nil, err
		}
		if comm != "pause\n" {
			log.Info("pause has been resumed", "comm", comm)
			break
		}
		log.Info("the process hasn't resumed, step into the following loop", "comm", comm)
	}

	return proc, nil
}

func (s *DaemonServer) ExecMemoryStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*bpm.Process, error) {
	log := s.getLoggerFromContext(ctx)
	if req.MemoryStressors == "" {
		return nil, nil
	}
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return nil, err
	}

	attachCGroup, err := cgroups.GetAttacherForPID(int(pid))
	if err != nil {
		return nil, err
	}

	args := strings.Fields(req.MemoryStressors)
	if err := resolveMemorySize(args, attachCGroup.TargetCGroup().MemoryLimit); err != nil {
		return nil, errors.Wrap(err, "resolve memory stress size")
	}
	processBuilder := bpm.DefaultProcessBuilder("memStress", args...).
		EnablePause()

	if req.OomScoreAdj != 0 {
		processBuilder = processBuilder.SetOOMScoreAdj(int(req.OomScoreAdj))
	}
	if req.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS)
	}
	cmd := processBuilder.Build(ctx)

	proc, err := s.backgroundProcessManager.StartProcess(ctx, cmd)
	if err != nil {
		return nil, err
	}
	log.Info("Start memStress successfully", "command", cmd, "pid", proc.Pair.Pid, "uid", proc.UID)

	if err = attachCGroup.AttachProcess(proc.Pair.Pid); err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.Error(kerr, "kill memStress failed", "request", req)
		}
		return nil, err
	}

	for {
		// TODO: find a better way to resume pause process
		if err := cmd.Process.Signal(syscall.SIGCONT); err != nil {
			return nil, err
		}

		log.Info("send signal to resume process")
		time.Sleep(time.Millisecond)
		comm, err := util.ReadCommName(proc.Pair.Pid)

		if err != nil {
			return nil, err
		}
		if comm != "pause\n" {
			log.Info("pause has been resumed", "comm", comm)
			break
		}
		log.Info("the process hasn't resumed, step into the following loop", "comm", comm)
	}

	return proc, nil
}

func resolveMemorySize(args []string, memoryLimit func() (uint64, error)) error {
	var size string
	var index int
	// memStress uses Go flags: the last size before the first positional argument wins.
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" || arg == "-" || !strings.HasPrefix(arg, "-") {
			break
		}
		name, value, inline := strings.Cut(arg, "=")
		if name == "--client" || name == "-client" {
			continue
		}
		if !inline {
			i++
			if i == len(args) {
				break
			}
			value = args[i]
		}
		if name == "--size" || name == "-size" {
			size, index = value, i
		}
	}
	value, percent := strings.CutSuffix(size, "%")
	if !percent {
		return nil
	}
	percentage, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(percentage) || math.IsInf(percentage, 0) || percentage < 0 {
		return errors.Errorf("invalid memory percentage %q", size)
	}

	total, err := memoryLimit()
	if err != nil {
		return err
	}
	if total == cgroups.UnlimitedMemory {
		info, err := mem.VirtualMemory()
		if err != nil {
			return err
		}
		if info.Total == 0 {
			return errors.New("node MemTotal unavailable")
		}
		total = info.Total
	}
	bytes := float64(total) / 100 * percentage
	if bytes >= math.Exp2(64) {
		return errors.Errorf("memory size %s of %d bytes overflows uint64", size, total)
	}
	args[index] = strings.TrimSuffix(args[index], size) + strconv.FormatUint(uint64(bytes), 10) + "B"
	return nil
}
