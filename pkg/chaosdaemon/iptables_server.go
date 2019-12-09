package chaosdaemon

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/pingcap/chaos-operator/pkg/chaosdaemon/pb"
)

const iptablesCmd = "iptables"

func (s *Server) FlushIptables(ctx context.Context, req *pb.IpTablesRequest) (*empty.Empty, error) {
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	nsPath := GenNetnsPath(pid)

	rule := req.Rule

	format := ""

	switch rule.Direction {
	case pb.Rule_INPUT:
		format = "%s INPUT -m set --match-set %s src -j DROP -w 5"
	case pb.Rule_OUTPUT:
		format = "%s OUTPUT -m set --match-set %s dst -j DROP -w 5"
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

	if rule.Action == pb.Rule_DELETE {
		output := ""

		for !strings.Contains(output, "Bad rule (does a matching rule exist in that chain?).") { // delete until all equal rules are deleted
			cmd := withNetNS(ctx, nsPath, iptablesCmd, strings.Split(command, " ")...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				output = string(out)
			}
		}
	} else {
		cmd := withNetNS(ctx, nsPath, iptablesCmd, strings.Split(command, " ")...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			output := string(out)
			log.Info("run command failed", "command", fmt.Sprintf("%s %s", iptablesCmd, command), "stdout", output)
			return nil, err
		}
	}

	return &empty.Empty{}, nil
}
