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

const iptablesCmd = "iptables"

func (s *Server) FlushIptables(ctx context.Context, req *pb.IpTablesRequest) (*empty.Empty, error) {
	{
		pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
		if err != nil {
			log.Error(err, "error while getting PID")
			return nil, err
		}

		nsPath := fmt.Sprintf("%s/%d/ns/net", defaultProcPrefix, pid)
		ns, err := netns.GetFromPath(nsPath)
		if err != nil {
			log.Error(err, "error while finding network namespace", "nsPath", nsPath)
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

	rule := req.Rule

	format := ""

	switch rule.Direction {
	case pb.Rule_INPUT:
		format = "%s INPUT -m set --match-set %s src -j DROP"
	case pb.Rule_OUTPUT:
		format = "%s OUTPUT -m set --match-set %s dst -j DROP"
	default:
		return nil, fmt.Errorf("unknown rule direction")
	}

	action := ""
	switch rule.Action {
	case pb.Rule_ADD:
		action = "-A"
	case pb.Rule_DELETE:
		action = "-D"
	}

	command := fmt.Sprintf(format, action, rule.Set)

	cmd := exec.CommandContext(ctx, iptablesCmd, strings.Split(command, " ")...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if !(rule.Action == pb.Rule_DELETE) { // TODO: handle error more carefully
			log.Info("run command failed", "command", fmt.Sprintf("%s %s", iptablesCmd, command), "stdout", output)
			return nil, err
		}
	}

	return &empty.Empty{}, nil
}
