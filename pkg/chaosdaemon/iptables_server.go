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
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

const (
	iptablesCmd = "iptables"

	iptablesChainAlreadyExistErr = "iptables: Chain already exists."
)

func (s *DaemonServer) SetIptablesChains(ctx context.Context, req *pb.IptablesChainsRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)
	log.Info("Set iptables chains", "request", req)

	pid, err := s.netNSPid(ctx, log, req.HostNs, req.ContainerId, req.PodUid)
	if err != nil {
		return nil, err
	}

	iptables := buildIptablesClient(ctx, req.EnterNS || req.HostNs, pid)
	iptables.hostNS = req.HostNs
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

// hostNetNSPid is init's PID; chaos-daemon runs with hostPID, so its network
// namespace is the host's.
const hostNetNSPid uint32 = 1

// netNSPid returns a PID whose network namespace is the one to operate in.
func (s *DaemonServer) netNSPid(ctx context.Context, log logr.Logger, hostNS bool, containerID, podUID string) (uint32, error) {
	if hostNS {
		return hostNetNSPid, nil
	}

	pid, err := s.crClient.GetPidFromContainerID(ctx, containerID)
	if err != nil {
		log.Info("container PID unavailable, falling back to sandbox", "error", err.Error())

		pid, err = s.crClient.GetSandboxPidFromPodUID(ctx, podUID)
		if err != nil {
			log.Error(err, "error while getting PID")
			return 0, err
		}
	}

	return pid, nil
}

type iptablesClient struct {
	ctx     context.Context
	enterNS bool
	pid     uint32
	hostNS  bool
}

type iptablesChain struct {
	Name  string
	Rules []string
}

func buildIptablesClient(ctx context.Context, enterNS bool, pid uint32) iptablesClient {
	return iptablesClient{
		ctx:     ctx,
		enterNS: enterNS,
		pid:     pid,
	}
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
	rules, err := buildIptablesRules(chain, iptables.hostNS)
	if err != nil {
		return err
	}

	err = iptables.createNewChain(&iptablesChain{
		Name:  chain.Name,
		Rules: rules,
	})
	if err != nil {
		return err
	}

	switch chain.Direction {
	case pb.Chain_INPUT:
		return iptables.ensureRule(&iptablesChain{
			Name: "CHAOS-INPUT",
		}, "-A CHAOS-INPUT -j "+chain.Name)
	case pb.Chain_OUTPUT:
		return iptables.ensureRule(&iptablesChain{
			Name: "CHAOS-OUTPUT",
		}, "-A CHAOS-OUTPUT -j "+chain.Name)
	default:
		return errors.Errorf("unknown direction %d", chain.Direction)
	}
}

// buildIptablesRules returns the rules of a chain: RETURN rules for excluded
// ipsets and exempt ports first, then the (optionally probabilistic) target rules.
func buildIptablesRules(chain *pb.Chain, hostNS bool) ([]string, error) {
	var matchPart, excludeMatchPart, interfaceMatcher string
	if chain.Direction == pb.Chain_INPUT {
		matchPart = "src,dst"
		excludeMatchPart = "src"
		interfaceMatcher = "-i"
	} else if chain.Direction == pb.Chain_OUTPUT {
		matchPart = "dst,dst"
		excludeMatchPart = "dst"
		interfaceMatcher = "-o"
	} else {
		return nil, errors.Errorf("unknown chain direction %d", chain.Direction)
	}

	// The host namespace has one interface per ENI, so match everything but loopback.
	deviceMatch := fmt.Sprintf("%s %s", interfaceMatcher, chain.Device)
	if chain.Device == "" {
		if hostNS {
			deviceMatch = fmt.Sprintf("! %s lo", interfaceMatcher)
		} else {
			deviceMatch = fmt.Sprintf("%s %s", interfaceMatcher, defaultDevice)
		}
	}

	rules := []string{}

	for _, ipset := range chain.ExcludeIpsets {
		rules = append(rules, fmt.Sprintf("-A %s %s -m set --match-set %s %s -j RETURN -w 5",
			chain.Name, deviceMatch, ipset, excludeMatchPart))
	}

	if len(chain.ExemptTcpPorts) > 0 {
		for _, portFlag := range []string{"--destination-ports", "--source-ports"} {
			rules = append(rules, fmt.Sprintf("-A %s %s --protocol tcp -m multiport %s %s -j RETURN -w 5",
				chain.Name, deviceMatch, portFlag, chain.ExemptTcpPorts))
		}
	}

	statistic := ""
	if chain.Probability > 0 && chain.Probability < 1 {
		statistic = fmt.Sprintf(" -m statistic --mode random --probability %s",
			strconv.FormatFloat(chain.Probability, 'f', -1, 64))
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

	if len(chain.Ipsets) == 0 {
		rules = append(rules, strings.TrimSpace(fmt.Sprintf("-A %s %s%s -j %s -w 5 %s", chain.Name, deviceMatch, statistic, chain.Target, protocolAndPort)))
	}

	for _, ipset := range chain.Ipsets {
		rules = append(rules, strings.TrimSpace(fmt.Sprintf("-A %s %s -m set --match-set %s %s%s -j %s -w 5 %s",
			chain.Name, deviceMatch, ipset, matchPart, statistic, chain.Target, protocolAndPort)))
	}

	return rules, nil
}

func (iptables *iptablesClient) initializeEnv() error {
	for _, direction := range []string{"INPUT", "OUTPUT"} {
		chainName := "CHAOS-" + direction

		err := iptables.createNewChain(&iptablesChain{
			Name:  chainName,
			Rules: []string{},
		})
		if err != nil {
			return err
		}

		iptables.ensureRule(&iptablesChain{
			Name:  direction,
			Rules: []string{},
		}, "-A "+direction+" -j "+chainName)
	}

	return nil
}

// createNewChain will cover existing chain
func (iptables *iptablesClient) createNewChain(chain *iptablesChain) error {
	processBuilder := bpm.DefaultProcessBuilder(iptablesCmd, "-w", "-N", chain.Name).SetContext(iptables.ctx)
	if iptables.enterNS {
		processBuilder = processBuilder.SetNS(iptables.pid, bpm.NetNS)
	}
	cmd := processBuilder.Build(iptables.ctx)
	out, err := cmd.CombinedOutput()

	if (err == nil && len(out) == 0) ||
		(err != nil && strings.Contains(string(out), iptablesChainAlreadyExistErr)) {
		// Successfully create a new chain
		return iptables.deleteAndWriteRules(chain)
	}

	return util.EncodeOutputToError(out, err)
}

// deleteAndWriteRules will remove all existing function in the chain
// and replace with the new settings
func (iptables *iptablesClient) deleteAndWriteRules(chain *iptablesChain) error {

	// This chain should already exist
	err := iptables.flushIptablesChain(chain)
	if err != nil {
		return err
	}

	for _, rule := range chain.Rules {
		err := iptables.ensureRule(chain, rule)
		if err != nil {
			return err
		}
	}

	return nil
}

func (iptables *iptablesClient) ensureRule(chain *iptablesChain, rule string) error {
	processBuilder := bpm.DefaultProcessBuilder(iptablesCmd, "-w", "-S", chain.Name).SetContext(iptables.ctx)
	if iptables.enterNS {
		processBuilder = processBuilder.SetNS(iptables.pid, bpm.NetNS)
	}
	cmd := processBuilder.Build(iptables.ctx)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return util.EncodeOutputToError(out, err)
	}

	if strings.Contains(string(out), rule) {
		// The required rule already exist in chain
		return nil
	}

	// TODO: lock on every container but not on chaos-daemon's `/run/xtables.lock`
	processBuilder = bpm.DefaultProcessBuilder(iptablesCmd, strings.Split("-w "+rule, " ")...).SetContext(iptables.ctx)
	if iptables.enterNS {
		processBuilder = processBuilder.SetNS(iptables.pid, bpm.NetNS)
	}
	cmd = processBuilder.Build(iptables.ctx)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return util.EncodeOutputToError(out, err)
	}

	return nil
}

func (iptables *iptablesClient) flushIptablesChain(chain *iptablesChain) error {
	processBuilder := bpm.DefaultProcessBuilder(iptablesCmd, "-w", "-F", chain.Name).SetContext(iptables.ctx)
	if iptables.enterNS {
		processBuilder = processBuilder.SetNS(iptables.pid, bpm.NetNS)
	}
	cmd := processBuilder.Build(iptables.ctx)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return util.EncodeOutputToError(out, err)
	}

	return nil
}
