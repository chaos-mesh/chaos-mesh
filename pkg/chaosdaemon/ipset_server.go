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
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

const (
	ipsetExistErr        = "set with the same name already exists"
	ipExistErr           = "it's already added"
	ipsetNewNameExistErr = "a set with the new name already exists"
)

type IPSetType string

const (
	SetIPSet     IPSetType = "list:set"
	NetPortIPSet IPSetType = "hash:net,port"
	NetIPSet     IPSetType = "hash:net"
)

func (s *DaemonServer) FlushIPSets(ctx context.Context, req *pb.IPSetsRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)
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
		s.IPSetLocker.Lock(ipset.SetName)
		err := flushIPSet(ctx, log, req.EnterNS, pid, ipset)
		s.IPSetLocker.Unlock(ipset.SetName)
		if err != nil {
			return nil, err
		}
	}

	return &empty.Empty{}, nil
}

func flushIPSet(ctx context.Context, log logr.Logger, enterNS bool, pid uint32, set *pb.IPSet) error {
	setName := set.SetName
	netPortName := set.NetPortName
	netName := set.NetName

	// If IP sets already exist, existing ones will be renamed to temp names.
	tmpSetName := fmt.Sprintf("%sold", setName)
	tmpNetPortName := fmt.Sprintf("%sold", netPortName)
	tmpNetName := fmt.Sprintf("%sold", netName)

	// IP sets can't be deleted if there are iptables rules referencing them.
	// Therefore, we create new sets and swap them.
	if err := createIPSet(ctx, log, enterNS, pid, tmpSetName, SetIPSet); err != nil {
		return err
	}
	if err := createIPSet(ctx, log, enterNS, pid, tmpNetPortName, NetPortIPSet); err != nil {
		return err
	}
	if err := createIPSet(ctx, log, enterNS, pid, tmpNetName, NetIPSet); err != nil {
		return err
	}

	// Add CIDR and port pairs to corresponding IP sets.
	if err := addCIDRsToIPSet(ctx, log, enterNS, pid, tmpNetPortName, tmpNetName, set.Cidrs); err != nil {
		return err
	}

	// Rename net,port and net IP sets to target names.
	if err := renameIPSet(ctx, log, enterNS, pid, tmpNetPortName, netPortName); err != nil {
		return err
	}
	if err := renameIPSet(ctx, log, enterNS, pid, tmpNetName, netName); err != nil {
		return err
	}

	// Add them to the set IP set.
	if err := addToIPSet(ctx, log, enterNS, pid, tmpSetName, netPortName); err != nil {
		return err
	}
	if err := addToIPSet(ctx, log, enterNS, pid, tmpSetName, netName); err != nil {
		return err
	}

	// Finally, rename the set IP set.
	err := renameIPSet(ctx, log, enterNS, pid, tmpSetName, setName)

	return err
}

func createIPSet(ctx context.Context, log logr.Logger, enterNS bool, pid uint32, name string, typ IPSetType) error {
	// ipset name cannot be longer than 31 bytes
	if len(name) > 31 {
		name = name[:31]
	}

	processBuilder := bpm.DefaultProcessBuilder("ipset", "create", name, string(typ)).SetContext(ctx)
	if enterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
	}

	cmd := processBuilder.Build(ctx)
	log.Info("create ipset", "command", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if !strings.Contains(output, ipsetExistErr) {
			log.Error(err, "ipset create error", "command", cmd.String(), "output", output)
			return util.EncodeOutputToError(out, err)
		}

		processBuilder = bpm.DefaultProcessBuilder("ipset", "flush", name).SetContext(ctx)
		if enterNS {
			processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
		}

		cmd = processBuilder.Build(ctx)
		log.Info("flush ipset", "command", cmd.String())

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(err, "ipset flush error", "command", cmd.String(), "output", string(out))
			return util.EncodeOutputToError(out, err)
		}
	}

	return nil
}

func addCIDRsToIPSet(ctx context.Context, log logr.Logger, enterNS bool, pid uint32, netPortName string, netName string, cidrsAndPorts []*pb.CidrAndPort) error {
	for _, cidr := range cidrsAndPorts {
		var name string
		var value string
		if cidr.Port == 0 {
			name = netName
			value = cidr.Cidr
		} else {
			name = netPortName
			value = fmt.Sprintf("%s,%d", cidr.Cidr, cidr.Port)
		}
		if err := addToIPSet(ctx, log, enterNS, pid, name, value); err != nil {
			return err
		}
	}

	return nil
}

func addToIPSet(ctx context.Context, log logr.Logger, enterNS bool, pid uint32, name string, value string) error {
	processBuilder := bpm.DefaultProcessBuilder("ipset", "add", name, value).SetContext(ctx)
	if enterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
	}
	cmd := processBuilder.Build(ctx)
	log.Info("add to ipset", "command", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if !strings.Contains(output, ipExistErr) {
			log.Error(err, "ipset add error", "command", cmd.String(), "output", output)
			return util.EncodeOutputToError(out, err)
		}
	}

	return nil
}

func renameIPSet(ctx context.Context, log logr.Logger, enterNS bool, pid uint32, oldName string, newName string) error {
	processBuilder := bpm.DefaultProcessBuilder("ipset", "rename", oldName, newName).SetContext(ctx)
	if enterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
	}

	cmd := processBuilder.Build(ctx)
	log.Info("rename ipset", "command", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if !strings.Contains(output, ipsetNewNameExistErr) {
			log.Error(err, "rename ipset failed", "command", cmd.String(), "output", output)
			return util.EncodeOutputToError(out, err)
		}

		// swap the old ipset and the new ipset if the new ipset already exist.
		processBuilder = bpm.DefaultProcessBuilder("ipset", "swap", oldName, newName).SetContext(ctx)
		if enterNS {
			processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
		}
		cmd := processBuilder.Build(ctx)
		log.Info("swap ipset", "command", cmd.String())

		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(err, "swap ipset failed", "command", cmd.String(), "output", string(out))
			return util.EncodeOutputToError(out, err)
		}
	}
	return nil
}
