// Copyright 2019 Chaos Mesh Authors.
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
	ipsetExistErr        = "set with the same name already exists"
	ipExistErr           = "it's already added"
	ipsetNewNameExistErr = "a set with the new name already exists"
)

func (s *daemonServer) FlushIpSet(ctx context.Context, req *pb.IpSetRequest) (*empty.Empty, error) {
	log.Info("flush ipset", "request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	nsPath := GetNsPath(pid, netNS)

	// TODO: lock every ipset when working on it

	set := req.Ipset
	name := set.Name

	// If the ipset already exists, the ipset will be renamed to this temp name.
	tmpName := fmt.Sprintf("%sold", name)

	// the ipset while existing iptables rules are using them can not be deleted,.
	// so we creates an temp ipset and swap it with existing one.
	if err := s.createIPSet(ctx, nsPath, tmpName); err != nil {
		return nil, err
	}

	// add ips to the temp ipset
	if err := s.addCIDRsToIPSet(ctx, nsPath, tmpName, set.Cidrs); err != nil {
		return nil, err
	}

	// rename the temp ipset with the target name of ipset if the taget ipset not exists,
	// otherwise swap  them with each other.
	if err := s.renameIPSet(ctx, nsPath, tmpName, name); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *daemonServer) createIPSet(ctx context.Context, nsPath string, name string) error {
	// ipset name cannot be longer than 31 bytes
	if len(name) > 31 {
		name = name[:31]
	}

	cmd := withNetNS(ctx, nsPath, "ipset", "create", name, "hash:net")

	log.Info("create ipset", "command", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if !strings.Contains(output, ipsetExistErr) {
			log.Error(err, "ipset create error", "command", cmd.String(), "output", output)
			return err
		}

		cmd := withNetNS(ctx, nsPath, "ipset", "flush", name)

		log.Info("flush ipset", "command", cmd.String())

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(err, "ipset flush error", "command", cmd.String(), "output", string(out))
			return err
		}
	}

	return nil
}

func (s *daemonServer) addCIDRsToIPSet(ctx context.Context, nsPath string, name string, cidrs []string) error {
	for _, cidr := range cidrs {
		cmd := withNetNS(ctx, nsPath, "ipset", "add", name, cidr)

		log.Info("add CIDR to ipset", "command", cmd.String())

		out, err := cmd.CombinedOutput()
		if err != nil {
			output := string(out)
			if !strings.Contains(output, ipExistErr) {
				log.Error(err, "ipset add error", "command", cmd.String(), "output", output)
				return err
			}
		}
	}

	return nil
}

func (s *daemonServer) renameIPSet(ctx context.Context, nsPath string, oldName string, newName string) error {
	cmd := withNetNS(ctx, nsPath, "ipset", "rename", oldName, newName)

	log.Info("rename ipset", "command", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if !strings.Contains(output, ipsetNewNameExistErr) {
			log.Error(err, "rename ipset failed", "command", cmd.String(), "output", output)
			return err
		}

		// swap the old ipset and the new ipset if the new ipset already exist.
		cmd := withNetNS(ctx, nsPath, "ipset", "swap", oldName, newName)

		log.Info("swap ipset", "command", cmd.String())

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(err, "swap ipset failed", "command", cmd.String(), "output", string(out))
			return err
		}
	}
	return nil
}
