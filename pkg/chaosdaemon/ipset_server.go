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
	pb "github.com/pingcap/chaos-operator/pkg/chaosdaemon/pb"
)

func (s *Server) FlushIpSet(ctx context.Context, req *pb.IpSetRequest) (*empty.Empty, error) {
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	nsPath := GenNetnsPath(pid)

	// TODO: lock every ipset when working on it

	set := req.Ipset

	name := set.Name

	{
		// TODO: Hash and get a stable short name (as ipset name cannot be longer than 31 byte)
		cmd := withNetNS(ctx, nsPath, "ipset", "create", name+"old", "hash:ip")
		out, err := cmd.CombinedOutput()
		if err != nil {
			output := string(out)
			if !strings.Contains(output, "set with the same name already exists") {
				log.Error(err, "ipset create error", "command", fmt.Sprintf("ipset create %s hash:ip", name+"old"), "output", output)
				return nil, err
			}

			cmd := withNetNS(ctx, nsPath, "ipset", "flush", name+"old")
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Error(err, "ipset flush error", "command", fmt.Sprintf("ipset flush %s", name+"old"), "output", string(out))
				return nil, err
			}
		}
	}

	for _, ip := range set.Ips {
		cmd := withNetNS(ctx, nsPath, "ipset", "add", name+"old", ip)
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
		cmd := withNetNS(ctx, nsPath, "ipset", "rename", name+"old", name)
		out, err := cmd.CombinedOutput()
		if err != nil {
			output := string(out)
			if !strings.Contains(output, "a set with the new name already exists") {
				log.Error(err, "rename ipset failed", "command", fmt.Sprintf("ipset rename %s %s", name+"old", name), "output", output)
				return nil, err
			}

			cmd := withNetNS(ctx, nsPath, "ipset", "swap", name+"old", name)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Error(err, "swap ipset failed", "output", string(out))
				return nil, err
			}
		}
	}

	return &empty.Empty{}, nil
}
