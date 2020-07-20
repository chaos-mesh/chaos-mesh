// Copyright 2019 PingCAP, Inc.
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
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	iptablesCmd = "iptables"

	iptablesBadRuleErr       = "Bad rule (does a matching rule exist in that chain?)."
	iptablesIPSetNotExistErr = "doesn't exist."
)

func (s *daemonServer) FlushIptablesChains(ctx context.Context, req *pb.IptablesChainsRequest) (*empty.Empty, error) {
	log.Info("Flush iptables chains", "request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	nsPath := GetNsPath(pid, netNS)

	err = flushIptablesChains(ctx, nsPath, req.Chains)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func flushIptablesChains(ctx context.Context, nsPath string, chains []*pb.Chain) error {
	for _, chain := range chains {
		err := flushIptablesChain(ctx, nsPath, chain)
		if err != nil {
			return err
		}
	}

	return nil
}

func flushIptablesChain(ctx context.Context, nsPath string, chain *pb.Chain) error {
	cmd := withNetNS(ctx, nsPath, iptablesCmd, "-N", chain.Name)
	out, err := cmd.CombinedOutput()
	if err == nil && len(out) == 0 {
		initializeChain(ctx, nsPath, chain)
	}
	if err != nil {
		if strings.Contains(string(out), "iptables: Chain already exists.") {
			cmd = withNetNS(ctx, nsPath, iptablesCmd, "-F", chain.Name)
			out, err = cmd.CombinedOutput()
			if err != nil {
				return encodeOutputToError(out, err)
			}
		} else {
			return encodeOutputToError(out, err)
		}
	}

	var matchPart string
	if chain.Direction == pb.Chain_INPUT {
		matchPart = "src"
	} else if chain.Direction == pb.Chain_OUTPUT {
		matchPart = "dst"
	} else {
		return fmt.Errorf("unknown direction %d", chain.Direction)
	}

	for _, ipset := range chain.Ipsets {
		cmd = withNetNS(ctx, nsPath, iptablesCmd, "-A", chain.Name, "-m", "set", "--match-set", ipset, matchPart, "-j", "DROP", "-w", "5")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return encodeOutputToError(out, err)
		}
	}

	return nil
}

func initializeChain(ctx context.Context, nsPath string, chain *pb.Chain) error {
	if chain.Direction == pb.Chain_INPUT {
		cmd := withNetNS(ctx, nsPath, iptablesCmd, "-A", "INPUT", "-j", chain.Name)
		_, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
	} else if chain.Direction == pb.Chain_OUTPUT {
		cmd := withNetNS(ctx, nsPath, iptablesCmd, "-A", "OUTPUT", "-j", chain.Name)
		_, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown direction %d", chain.Direction)
	}

	return nil
}

func encodeOutputToError(output []byte, err error) error {
	return fmt.Errorf("error code: %d, msg: %s", err, string(output))
}
