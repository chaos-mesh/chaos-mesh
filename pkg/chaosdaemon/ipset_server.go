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
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	ipsetExistErr        = "set with the same name already exists"
	ipExistErr           = "it's already added"
	ipsetNewNameExistErr = "a set with the new name already exists"
)

func (s *DaemonServer) FlushIPSets(ctx context.Context, req *pb.IPSetsRequest) (*empty.Empty, error) {
	log.Info("flush ipset", "request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	for _, ipset := range req.Ipsets {
		// All operations on the ipset with the same name should be serialized,
		// because ipset is not isolated with namespace in linux < 3.12

		// **Notice**: Serialization should be enough for Chaos Mesh (but no
		// need to use name to simulate isolation), because the operation on
		// the ipset with the same name should be same for NetworkChaos.
		// It's a bad solution, only for the users who don't want to upgrade
		// their linux version to 3.12 :(
		ipset := ipset
		s.IPSetLocker.Lock(ipset.Name)
		err := flushIPSet(ctx, req.EnterNS, pid, ipset)
		s.IPSetLocker.Unlock(ipset.Name)
		if err != nil {
			return nil, err
		}
	}

	return &empty.Empty{}, nil
}

func flushIPSet(ctx context.Context, enterNS bool, pid uint32, set *pb.IPSet) error {
	name := set.Name

	// If the ipset already exists, the ipset will be renamed to this temp name.
	tmpName := fmt.Sprintf("%sold", name)

	// the ipset while existing iptables rules are using them can not be deleted,.
	// so we creates an temp ipset and swap it with existing one.
	if err := createIPSet(ctx, enterNS, pid, tmpName); err != nil {
		return err
	}

	// add ips to the temp ipset
	if err := addCIDRsToIPSet(ctx, enterNS, pid, tmpName, set.Cidrs); err != nil {
		return err
	}

	// rename the temp ipset with the target name of ipset if the taget ipset not exists,
	// otherwise swap  them with each other.
	err := renameIPSet(ctx, enterNS, pid, tmpName, name)

	return err
}

func createIPSet(ctx context.Context, enterNS bool, pid uint32, name string) error {
	// ipset name cannot be longer than 31 bytes
	if len(name) > 31 {
		name = name[:31]
	}

	processBuilder := bpm.DefaultProcessBuilder("ipset", "create", name, "hash:net").SetContext(ctx)
	if enterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
	}

	cmd := processBuilder.Build()
	log.Info("create ipset", "command", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if !strings.Contains(output, ipsetExistErr) {
			log.Error(err, "ipset create error", "command", cmd.String(), "output", output)
			return encodeOutputToError(out, err)
		}

		processBuilder = bpm.DefaultProcessBuilder("ipset", "flush", name).SetContext(ctx)
		if enterNS {
			processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
		}

		cmd = processBuilder.Build()
		log.Info("flush ipset", "command", cmd.String())

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(err, "ipset flush error", "command", cmd.String(), "output", string(out))
			return encodeOutputToError(out, err)
		}
	}

	return nil
}

func addCIDRsToIPSet(ctx context.Context, enterNS bool, pid uint32, name string, cidrs []string) error {
	for _, cidr := range cidrs {
		processBuilder := bpm.DefaultProcessBuilder("ipset", "add", name, cidr).SetContext(ctx)
		if enterNS {
			processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
		}
		cmd := processBuilder.Build()
		log.Info("add CIDR to ipset", "command", cmd.String())

		out, err := cmd.CombinedOutput()
		if err != nil {
			output := string(out)
			if !strings.Contains(output, ipExistErr) {
				log.Error(err, "ipset add error", "command", cmd.String(), "output", output)
				return encodeOutputToError(out, err)
			}
		}
	}

	return nil
}

func renameIPSet(ctx context.Context, enterNS bool, pid uint32, oldName string, newName string) error {
	processBuilder := bpm.DefaultProcessBuilder("ipset", "rename", oldName, newName).SetContext(ctx)
	if enterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
	}

	cmd := processBuilder.Build()
	log.Info("rename ipset", "command", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if !strings.Contains(output, ipsetNewNameExistErr) {
			log.Error(err, "rename ipset failed", "command", cmd.String(), "output", output)
			return encodeOutputToError(out, err)
		}

		// swap the old ipset and the new ipset if the new ipset already exist.
		processBuilder = bpm.DefaultProcessBuilder("ipset", "swap", oldName, newName).SetContext(ctx)
		if enterNS {
			processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
		}
		cmd := processBuilder.Build()
		log.Info("swap ipset", "command", cmd.String())

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(err, "swap ipset failed", "command", cmd.String(), "output", string(out))
			return encodeOutputToError(out, err)
		}
	}
	return nil
}
