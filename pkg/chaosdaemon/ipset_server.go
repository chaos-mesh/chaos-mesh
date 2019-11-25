package chaosdaemon

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/vishvananda/netns"

	pb "github.com/pingcap/chaos-operator/pkg/chaosdaemon/pb"
)

func (s *Server) FlushIpSet(ctx context.Context, req *pb.IpSetRequest) (*empty.Empty, error) {
	{
		pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
		if err != nil {
			log.Error(err, "error while getting PID")
			return nil, err
		}

		ns, err := netns.GetFromPid(int(pid))
		if err != nil {
			log.Error(err, "error while finding network namespace", "pid", pid)
			return nil, err
		}

		s.networkNamespaceLock.Lock()
		defer s.networkNamespaceLock.Unlock()
		err = netns.Set(ns)
		if err != nil {
			log.Error(err, "fail to set network namespace")
			return nil, err
		}
	}

	// TODO: lock every ipset when working on it

	set := req.Ipset

	name := set.Name

	{
		// TODO: Hash and get a stable short name (as ipset name cannot be longer than 31 byte)
		cmd := exec.CommandContext(ctx, "ipset", "create", name+"old", "hash:ip")
		out, err := cmd.CombinedOutput()
		if err != nil {
			output := string(out)
			if !strings.Contains(output, "set with the same name already exists") {
				log.Error(err, "ipset create error", "command", fmt.Sprintf("ipset create %s hash:ip", name+"old"), "output", output)
				return nil, err
			}

			cmd := exec.CommandContext(ctx, "ipset", "flush", name+"old")
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Error(err, "ipset flush error", "command", fmt.Sprintf("ipset flush %s", name+"old"), "output", string(out))
				return nil, err
			}
		}
	}

	for _, ip := range set.Ips {
		cmd := exec.CommandContext(ctx, "ipset", "add", name+"old", ip)
		out, err := cmd.CombinedOutput()
		if err != nil {
			output := string(out)
			if !strings.Contains(output, "it's already added") {
				log.Error(err, "ipset add error", "command", fmt.Sprintf("ipset add %s %s", name+"old", ip), "output", string(out))
				return nil, err
			}
		}
	}

	{
		cmd := exec.CommandContext(ctx, "ipset", "rename", name+"old", name)
		out, err := cmd.CombinedOutput()
		if err != nil {
			output := string(out)
			if !strings.Contains(output, "a set with the new name already exists") {
				log.Error(err, "rename ipset failed", "command", fmt.Sprintf("ipset rename %s %s", name+"old", name), "output", output)
				return nil, err
			}

			cmd := exec.CommandContext(ctx, "ipset", "swap", name+"old", name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Error(err, "swap ipset failed", "output", string(out))
				return nil, err
			}
		}
	}

	return &empty.Empty{}, nil
}
