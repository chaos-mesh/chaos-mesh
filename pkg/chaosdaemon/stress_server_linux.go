// Copyright 2020 Chaos Mesh Authors.
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
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/containerd/cgroups"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

var (
	// Possible cgroup subsystems
	cgroupSubsys = []string{"cpu", "memory", "systemd", "net_cls",
		"net_prio", "freezer", "blkio", "perf_event", "devices",
		"cpuset", "cpuacct", "pids", "hugetlb"}
)

func (s *DaemonServer) ExecStressors(ctx context.Context,
	req *pb.ExecStressRequest) (*pb.ExecStressResponse, error) {
	log.Info("Executing stressors", "request", req)
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.Target)
	if err != nil {
		return nil, err
	}
	path := pidPath(int(pid))
	id, err := s.crClient.FormatContainerID(ctx, req.Target)
	if err != nil {
		return nil, err
	}
	cgroup, err := findValidCgroup(path, id)
	if err != nil {
		return nil, err
	}
	if req.Scope == pb.ExecStressRequest_POD {
		cgroup, _ = filepath.Split(cgroup)
	}
	control, err := cgroups.Load(cgroups.V1, cgroups.StaticPath(cgroup))
	if err != nil {
		return nil, err
	}

	processBuilder := bpm.DefaultProcessBuilder("stress-ng", strings.Fields(req.Stressors)...).
		EnablePause()
	if !req.WithoutNS {
		processBuilder = processBuilder.SetNS(pid, bpm.PidNS)
	}
	cmd := processBuilder.Build()

	err = s.backgroundProcessManager.StartProcess(cmd)
	if err != nil {
		return nil, err
	}
	log.Info("Start process successfully")

	procState, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return nil, err
	}
	ct, err := procState.CreateTime()

	if err = control.Add(cgroups.Process{Pid: cmd.Process.Pid}); err != nil {
		if kerr := cmd.Process.Kill(); kerr != nil {
			log.Error(kerr, "kill stressors failed", "request", req)
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

		comm, err := ReadCommName(cmd.Process.Pid)
		if err != nil {
			return nil, err
		}
		if comm != "pause\n" {
			log.Info("pause has been resumed", "comm", comm)
			break
		}
		log.Info("the process hasn't resumed, step into the following loop", "comm", comm)
	}

	return &pb.ExecStressResponse{
		Instance:  strconv.Itoa(cmd.Process.Pid),
		StartTime: ct,
	}, nil
}

var errFinished = "os: process already finished"

func (s *DaemonServer) CancelStressors(ctx context.Context,
	req *pb.CancelStressRequest) (*empty.Empty, error) {
	pid, err := strconv.Atoi(req.Instance)
	if err != nil {
		return nil, err
	}
	log.Info("Canceling stressors", "request", req)

	err = s.backgroundProcessManager.KillBackgroundProcess(ctx, pid, req.StartTime)
	if err != nil {
		return nil, err
	}
	log.Info("killing stressor successfully")
	return &empty.Empty{}, nil
}

func findValidCgroup(path cgroups.Path, target string) (string, error) {
	for _, subsys := range cgroupSubsys {
		p, err := path(cgroups.Name(subsys))
		if err != nil {
			log.Error(err, "Failed to retrieve the cgroup path", "subsystem", subsys, "target", target)
			continue
		}
		if strings.Contains(p, target) {
			return p, nil
		}
	}
	return "", fmt.Errorf("never found valid cgroup for %s", target)
}

// pidPath will return the correct cgroup paths for an existing process running inside a cgroup
// This is commonly used for the Load function to restore an existing container.
//
// Note: it is migrated from cgroups.pidPath since it will find mountinfo incorrectly inside
// the daemonset. Hope we can fix it in official cgroups repo to solve it.
func pidPath(pid int) cgroups.Path {
	p := fmt.Sprintf("/proc/%d/cgroup", pid)
	paths, err := parseCgroupFile(p)
	if err != nil {
		return errorPath(errors.Wrapf(err, "parse cgroup file %s", p))
	}
	return existingPath(paths, pid, "")
}

func errorPath(err error) cgroups.Path {
	return func(_ cgroups.Name) (string, error) {
		return "", err
	}
}

func existingPath(paths map[string]string, pid int, suffix string) cgroups.Path {
	// localize the paths based on the root mount dest for nested cgroups
	for n, p := range paths {
		dest, err := getCgroupDestination(pid, string(n))
		if err != nil {
			return errorPath(err)
		}
		rel, err := filepath.Rel(dest, p)
		if err != nil {
			return errorPath(err)
		}
		if rel == "." {
			rel = dest
		}
		paths[n] = filepath.Join("/", rel)
	}
	return func(name cgroups.Name) (string, error) {
		root, ok := paths[string(name)]
		if !ok {
			if root, ok = paths[fmt.Sprintf("name=%s", name)]; !ok {
				return "", cgroups.ErrControllerNotActive
			}
		}
		if suffix != "" {
			return filepath.Join(root, suffix), nil
		}
		return root, nil
	}
}

func parseCgroupFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseCgroupFromReader(f)
}

func parseCgroupFromReader(r io.Reader) (map[string]string, error) {
	var (
		cgroups = make(map[string]string)
		s       = bufio.NewScanner(r)
	)
	for s.Scan() {
		var (
			text  = s.Text()
			parts = strings.SplitN(text, ":", 3)
		)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid cgroup entry: %q", text)
		}
		for _, subs := range strings.Split(parts[1], ",") {
			if subs != "" {
				cgroups[subs] = parts[2]
			}
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return cgroups, nil
}

func getCgroupDestination(pid int, subsystem string) (string, error) {
	// use the process's mount info
	p := fmt.Sprintf("/proc/%d/mountinfo", pid)
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[3], nil
			}
		}
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("never found desct for %s", subsystem)
}
