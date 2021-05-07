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
	iptablesCmd = "iptables"

	iptablesChainAlreadyExistErr = "iptables: Chain already exists."
)

func (s *DaemonServer) SetIptablesChains(ctx context.Context, req *pb.IptablesChainsRequest) (*empty.Empty, error) {
	log.Info("Set iptables chains", "request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	iptables := buildIptablesClient(ctx, req.EnterNS, pid)
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
	ctx     context.Context
	enterNS bool
	pid     uint32
}

type iptablesChain struct {
	Name  string
	Rules []string
}

func buildIptablesClient(ctx context.Context, enterNS bool, pid uint32) iptablesClient {
	return iptablesClient{
		ctx,
		enterNS,
		pid,
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
	var matchPart string
	if chain.Direction == pb.Chain_INPUT {
		matchPart = "src"
	} else if chain.Direction == pb.Chain_OUTPUT {
		matchPart = "dst"
	} else {
		return fmt.Errorf("unknown chain direction %d", chain.Direction)
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

	if len(chain.Ipsets) == 0 {
		rules = append(rules, strings.TrimSpace(fmt.Sprintf("-A %s -j %s -w 5 %s", chain.Name, chain.Target, protocolAndPort)))
	}

	for _, ipset := range chain.Ipsets {
		rules = append(rules, strings.TrimSpace(fmt.Sprintf("-A %s -m set --match-set %s %s -j %s -w 5 %s",
			chain.Name, ipset, matchPart, chain.Target, protocolAndPort)))
	}
	err := iptables.createNewChain(&iptablesChain{
		Name:  chain.Name,
		Rules: rules,
	})
	if err != nil {
		return err
	}

	if chain.Direction == pb.Chain_INPUT {
		err := iptables.ensureRule(&iptablesChain{
			Name: "CHAOS-INPUT",
		}, "-A CHAOS-INPUT -j "+chain.Name)
		if err != nil {
			return err
		}
	} else if chain.Direction == pb.Chain_OUTPUT {
		iptables.ensureRule(&iptablesChain{
			Name: "CHAOS-OUTPUT",
		}, "-A CHAOS-OUTPUT -j "+chain.Name)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown direction %d", chain.Direction)
	}
	return nil
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
	cmd := processBuilder.Build()
	out, err := cmd.CombinedOutput()

	if (err == nil && len(out) == 0) ||
		(err != nil && strings.Contains(string(out), iptablesChainAlreadyExistErr)) {
		// Successfully create a new chain
		return iptables.deleteAndWriteRules(chain)
	}

	return encodeOutputToError(out, err)
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
	cmd := processBuilder.Build()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(out, err)
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
	cmd = processBuilder.Build()
	out, err = cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(out, err)
	}

	return nil
}

func (iptables *iptablesClient) flushIptablesChain(chain *iptablesChain) error {
	processBuilder := bpm.DefaultProcessBuilder(iptablesCmd, "-w", "-F", chain.Name).SetContext(iptables.ctx)
	if iptables.enterNS {
		processBuilder = processBuilder.SetNS(iptables.pid, bpm.NetNS)
	}
	cmd := processBuilder.Build()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(out, err)
	}

	return nil
}
