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
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

const chainAlreadyExistErr = "Chain already exists."

func (s *DaemonServer) SetIptablesChains(ctx context.Context, req *pb.IptablesChainsRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)
	log.Info("Set iptables chains", "request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Info("container PID unavailable, falling back to sandbox", "error", err.Error())

		pid, err = s.crClient.GetSandboxPidFromPodUID(ctx, req.PodUid)
		if err != nil {
			log.Error(err, "error while getting PID")
			return nil, err
		}
	}

	iptables := buildIptablesClient(ctx, log, req.EnterNS, pid)
	err = iptables.initializeEnv()
	if err != nil {
		log.Error(err, "error while initializing iptables")
		return nil, err
	}

	err = iptables.setIptablesChains(req.Chains)
	if err != nil {
		log.Error(err, "error while setting iptables chains")
		return nil, err
	}

	return &empty.Empty{}, nil
}

type iptablesClient struct {
	ctx          context.Context
	log          logr.Logger
	enterNS      bool
	pid          uint32
	ip4Available bool
	ip6Available bool
}

type iptablesChain struct {
	Name  string
	Rules []string
}

func buildIptablesClient(ctx context.Context, log logr.Logger, enterNS bool, pid uint32) iptablesClient {
	c := iptablesClient{
		ctx:          ctx,
		log:          log,
		enterNS:      enterNS,
		pid:          pid,
		ip4Available: true,
		ip6Available: true,
	}
	pb := bpm.DefaultProcessBuilder("ip", "-4", "route", "show").SetContext(ctx)
	if enterNS {
		pb = pb.SetNS(pid, bpm.NetNS)
	}
	output, err := pb.Build(ctx).CombinedOutput()
	if err != nil || len(bytes.TrimSpace(output)) == 0 {
		log.Info("IPv4 unavailable, IPv4 network chaos will be skipped",
			"error", err, "output", output)
		c.ip4Available = false
	}
	pb = bpm.DefaultProcessBuilder("ip", "-6", "route", "show").SetContext(ctx)
	if enterNS {
		pb = pb.SetNS(pid, bpm.NetNS)
	}
	output, err = pb.Build(ctx).CombinedOutput()
	if err != nil || len(bytes.TrimSpace(output)) == 0 {
		log.Info("IPv6 unavailable, IPv6 network chaos will be skipped",
			"error", err, "output", output)
		c.ip6Available = false
	}
	return c
}

func (iptables *iptablesClient) setIptablesChains(chains []*pb.Chain) error {
	for _, chain := range chains {
		err := iptables.setIptablesChain(chain)
		if err != nil {
			return err
		}
	}

	return nil
}

func (iptables *iptablesClient) setIptablesChain(chain *pb.Chain) error {
	if chain.IpVersion == pb.IpVersion_IPv4 && !iptables.ip4Available {
		iptables.log.Info("ipv4 unavailable, skipping chain", "chain", chain.Name)
		return nil
	} else if chain.IpVersion == pb.IpVersion_IPv6 && !iptables.ip6Available {
		iptables.log.Info("ipv6 unavailable, skipping chain", "chain", chain.Name)
		return nil
	} else if chain.IpVersion != pb.IpVersion_IPv4 &&
		chain.IpVersion != pb.IpVersion_IPv6 {
		return errors.Errorf("unknown ip version %d", chain.IpVersion)
	}

	var matchPart string
	var interfaceMatcher string
	if chain.Direction == pb.Chain_INPUT {
		matchPart = "src,dst"
		interfaceMatcher = "-i"
	} else if chain.Direction == pb.Chain_OUTPUT {
		matchPart = "dst,dst"
		interfaceMatcher = "-o"
	} else {
		return errors.Errorf("unknown chain direction %d", chain.Direction)
	}

	if chain.Device == "" {
		chain.Device = defaultDevice
	}

	protocolAndPort := ""
	if len(chain.Protocol) > 0 {
		protocolAndPort += fmt.Sprintf("--protocol %s", chain.Protocol)

		if len(chain.SourcePorts) > 0 {
			if strings.Contains(chain.SourcePorts, ",") {
				protocolAndPort += fmt.Sprintf(" -m multiport --source-ports %s", chain.SourcePorts)
			} else {
				protocolAndPort += fmt.Sprintf(" --source-port %s", chain.SourcePorts)
			}
		}

		if len(chain.DestinationPorts) > 0 {
			if strings.Contains(chain.DestinationPorts, ",") {
				protocolAndPort += fmt.Sprintf(" -m multiport --destination-ports %s", chain.DestinationPorts)
			} else {
				protocolAndPort += fmt.Sprintf(" --destination-port %s", chain.DestinationPorts)
			}
		}

		if len(chain.TcpFlags) > 0 {
			protocolAndPort += fmt.Sprintf(" --tcp-flags %s", chain.TcpFlags)
		}
	}

	rules := []string{}
	ipv6 := chain.IpVersion == pb.IpVersion_IPv6

	if len(chain.Ipsets) == 0 {
		rules = append(rules, strings.TrimSpace(fmt.Sprintf("-A %s %s %s -j %s -w 5 %s", chain.Name, interfaceMatcher, chain.Device, chain.Target, protocolAndPort)))
	}

	for _, ipset := range chain.Ipsets {
		rules = append(rules, strings.TrimSpace(fmt.Sprintf("-A %s %s %s -m set --match-set %s %s -j %s -w 5 %s",
			chain.Name, interfaceMatcher, chain.Device, ipset, matchPart, chain.Target, protocolAndPort)))
	}
	err := iptables.createNewChain(ipv6, &iptablesChain{
		Name:  chain.Name,
		Rules: rules,
	})
	if err != nil {
		return err
	}

	family := "CHAOS"
	if ipv6 {
		family = "CHAOS6"
	}
	direction := "OUTPUT"
	if chain.Direction == pb.Chain_INPUT {
		direction = "INPUT"
	}
	chaosChain := family + "-" + direction
	return iptables.ensureRule(ipv6, &iptablesChain{Name: chaosChain}, "-A "+chaosChain+" -j "+chain.Name)
}

func (iptables *iptablesClient) initializeEnv() error {
	if iptables.ip4Available {
		for _, direction := range []string{"INPUT", "OUTPUT"} {
			chainName := "CHAOS-" + direction
			if err := iptables.createNewChain(false, &iptablesChain{Name: chainName}); err != nil {
				return err
			}
			err := iptables.ensureRule(false, &iptablesChain{Name: direction}, "-A "+direction+" -j "+chainName)
			if err != nil {
				return err
			}
		}
	}

	if iptables.ip6Available {
		for _, direction := range []string{"INPUT", "OUTPUT"} {
			chainName := "CHAOS6-" + direction
			if err := iptables.createNewChain(true, &iptablesChain{Name: chainName}); err != nil {
				return err
			}
			err := iptables.ensureRule(true, &iptablesChain{Name: direction}, "-A "+direction+" -j "+chainName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// createNewChain will cover existing chain
func (iptables *iptablesClient) createNewChain(ipv6 bool, chain *iptablesChain) error {
	cmd := iptablesCmd(ipv6)
	processBuilder := bpm.DefaultProcessBuilder(cmd, "-w", "-N", chain.Name).SetContext(iptables.ctx)
	if iptables.enterNS {
		processBuilder = processBuilder.SetNS(iptables.pid, bpm.NetNS)
	}
	cmdExec := processBuilder.Build(iptables.ctx)
	out, err := cmdExec.CombinedOutput()

	if (err == nil && len(out) == 0) ||
		(err != nil && strings.Contains(string(out), chainAlreadyExistErr)) {
		return iptables.deleteAndWriteRules(ipv6, chain)
	}

	return util.EncodeOutputToError(out, err)
}

// deleteAndWriteRules will remove all existing function in the chain
// and replace with the new settings
func (iptables *iptablesClient) deleteAndWriteRules(ipv6 bool, chain *iptablesChain) error {

	// This chain should already exist
	if err := iptables.flushIptablesChain(ipv6, chain); err != nil {
		return err
	}

	for _, rule := range chain.Rules {
		if err := iptables.ensureRule(ipv6, chain, rule); err != nil {
			return err
		}
	}

	return nil
}

func (iptables *iptablesClient) ensureRule(ipv6 bool, chain *iptablesChain, rule string) error {
	cmd := iptablesCmd(ipv6)
	processBuilder := bpm.DefaultProcessBuilder(cmd, "-w", "-S", chain.Name).SetContext(iptables.ctx)
	if iptables.enterNS {
		processBuilder = processBuilder.SetNS(iptables.pid, bpm.NetNS)
	}
	cmdExec := processBuilder.Build(iptables.ctx)
	out, err := cmdExec.CombinedOutput()
	if err != nil {
		return util.EncodeOutputToError(out, err)
	}

	if strings.Contains(string(out), rule) {
		// The required rule already exist in chain
		return nil
	}

	// TODO: lock on every container but not on chaos-daemon's `/run/xtables.lock`
	processBuilder = bpm.DefaultProcessBuilder(cmd, strings.Split("-w "+rule, " ")...).SetContext(iptables.ctx)
	if iptables.enterNS {
		processBuilder = processBuilder.SetNS(iptables.pid, bpm.NetNS)
	}
	cmdExec = processBuilder.Build(iptables.ctx)
	out, err = cmdExec.CombinedOutput()
	if err != nil {
		return util.EncodeOutputToError(out, err)
	}

	return nil
}

func (iptables *iptablesClient) flushIptablesChain(ipv6 bool, chain *iptablesChain) error {
	cmd := iptablesCmd(ipv6)
	processBuilder := bpm.DefaultProcessBuilder(cmd, "-w", "-F", chain.Name).SetContext(iptables.ctx)
	if iptables.enterNS {
		processBuilder = processBuilder.SetNS(iptables.pid, bpm.NetNS)
	}
	cmdExec := processBuilder.Build(iptables.ctx)
	out, err := cmdExec.CombinedOutput()
	if err != nil {
		return util.EncodeOutputToError(out, err)
	}

	return nil
}

func iptablesCmd(ipv6 bool) string {
	if ipv6 {
		return "ip6tables"
	}
	return "iptables"
}
