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
	"os/exec"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

const (
	chainAlreadyExistErr = "Chain already exists."
)

var filterTableUnavailableSubstrings = []string{
	"Table does not exist",
	"Address family not supported by protocol",
	"Protocol not supported",
}

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

	iptables := iptablesClient{ctx: ctx, enterNS: req.EnterNS, pid: pid}
	if err := iptables.reconcileChains(req.Chains); err != nil {
		log.Error(err, "error while reconciling iptables chains")
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

func isFirewallFamilyUnavailable(out []byte, err error) bool {
	if errors.Is(err, exec.ErrNotFound) {
		return true
	}
	msg := string(out)
	for _, pattern := range filterTableUnavailableSubstrings {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

// probeFamily checks whether the filter table for version is usable.
// It returns (unavailErr, nil) when the kernel/binary does not support that family,
// and (nil, fatalErr) for unexpected probe errors such as permission failures.
func (iptables *iptablesClient) probeFamily(version pb.IpVersion) (unavailErr error, fatalErr error) {
	if version != pb.IpVersion_IPv4 && version != pb.IpVersion_IPv6 {
		return nil, errors.Errorf("unknown ip version %d", version)
	}
	ipv6 := version == pb.IpVersion_IPv6
	out, err := iptables.run(ipv6, "-t", "filter", "-S")
	if err == nil {
		return nil, nil
	}
	wrapped := util.EncodeOutputToError(out, err)
	if !isFirewallFamilyUnavailable(out, err) {
		return nil, wrapped
	}
	return wrapped, nil
}

func (iptables *iptablesClient) reconcileChains(chains []*pb.Chain) error {
	for _, chain := range chains {
		if chain.IpVersion != pb.IpVersion_IPv4 && chain.IpVersion != pb.IpVersion_IPv6 {
			return errors.Errorf("unknown ip version %d", chain.IpVersion)
		}
	}

	versions := []pb.IpVersion{pb.IpVersion_IPv4, pb.IpVersion_IPv6}
	unavailable := make(map[pb.IpVersion]error)
	for _, version := range versions {
		unavailErr, fatalErr := iptables.probeFamily(version)
		if fatalErr != nil {
			return fatalErr
		}
		if unavailErr != nil {
			unavailable[version] = unavailErr
		}
	}

	for _, chain := range chains {
		if err := unavailable[chain.IpVersion]; err != nil && !chain.Optional {
			return err
		}
	}

	if len(chains) > 0 && len(unavailable) == len(versions) {
		return unavailable[chains[0].IpVersion]
	}

	for _, version := range versions {
		if unavailable[version] == nil {
			if err := iptables.initializeEnv(version == pb.IpVersion_IPv6); err != nil {
				return err
			}
		}
	}

	for _, chain := range chains {
		if unavailable[chain.IpVersion] == nil {
			if err := iptables.setIptablesChain(chain); err != nil {
				return err
			}
		}
	}

	return nil
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
	if chain.IpVersion != pb.IpVersion_IPv4 && chain.IpVersion != pb.IpVersion_IPv6 {
		return errors.Errorf("unknown ip version %d", chain.IpVersion)
	}

	rules, err := buildIptablesRules(chain)
	if err != nil {
		return err
	}

	ipv6 := chain.IpVersion == pb.IpVersion_IPv6
	if err := iptables.createNewChain(ipv6, &iptablesChain{
		Name:  chain.Name,
		Rules: rules,
	}); err != nil {
		return err
	}

	direction := "OUTPUT"
	if chain.Direction == pb.Chain_INPUT {
		direction = "INPUT"
	}
	chaosChain := chaosChainName(ipv6, direction)
	return iptables.ensureRule(ipv6, &iptablesChain{Name: chaosChain}, "-A "+chaosChain+" -j "+chain.Name)
}

func buildIptablesRules(chain *pb.Chain) ([]string, error) {
	var matchPart string
	var interfaceMatcher string
	if chain.Direction == pb.Chain_INPUT {
		matchPart = "src,dst"
		interfaceMatcher = "-i"
	} else if chain.Direction == pb.Chain_OUTPUT {
		matchPart = "dst,dst"
		interfaceMatcher = "-o"
	} else {
		return nil, errors.Errorf("unknown chain direction %d", chain.Direction)
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

	rules := make([]string, 0, 2*len(chain.Ipsets)+2)

	if len(chain.Ipsets) == 0 {
		rule := fmt.Sprintf("-A %s %s %s", chain.Name, interfaceMatcher, chain.Device)
		if protocolAndPort != "" {
			rule += " " + protocolAndPort
		}
		rules = append(rules, fmt.Sprintf("%s -j %s -w 5", rule, chain.Target))
		if chain.FallbackTarget != "" {
			rules = append(rules, fmt.Sprintf("-A %s %s %s -j %s -w 5", chain.Name, interfaceMatcher, chain.Device, chain.FallbackTarget))
		}
	}

	for _, ipset := range chain.Ipsets {
		rule := fmt.Sprintf("-A %s %s %s -m set --match-set %s %s", chain.Name, interfaceMatcher, chain.Device, ipset, matchPart)
		if protocolAndPort != "" {
			rule += " " + protocolAndPort
		}
		rules = append(rules, fmt.Sprintf("%s -j %s -w 5", rule, chain.Target))
		if chain.FallbackTarget != "" {
			rules = append(rules, fmt.Sprintf("-A %s %s %s -m set --match-set %s %s -j %s -w 5",
				chain.Name, interfaceMatcher, chain.Device, ipset, matchPart, chain.FallbackTarget))
		}
	}

	return rules, nil
}

func (iptables *iptablesClient) initializeEnv(ipv6 bool) error {
	for _, direction := range []string{"INPUT", "OUTPUT"} {
		chainName := chaosChainName(ipv6, direction)
		if err := iptables.createNewChain(ipv6, &iptablesChain{Name: chainName}); err != nil {
			return err
		}
		if err := iptables.ensureRule(ipv6, &iptablesChain{Name: direction}, "-A "+direction+" -j "+chainName); err != nil {
			return err
		}
	}
	return nil
}

func chaosChainName(ipv6 bool, direction string) string {
	if ipv6 {
		return "CHAOS6-" + direction
	}
	return "CHAOS-" + direction
}

func (iptables *iptablesClient) run(ipv6 bool, args ...string) ([]byte, error) {
	processBuilder := bpm.DefaultProcessBuilder(iptablesCmd(ipv6), append([]string{"-w"}, args...)...).SetContext(iptables.ctx)
	if iptables.enterNS {
		processBuilder = processBuilder.SetNS(iptables.pid, bpm.NetNS)
	}
	return processBuilder.Build(iptables.ctx).CombinedOutput()
}

// createNewChain will cover existing chain
func (iptables *iptablesClient) createNewChain(ipv6 bool, chain *iptablesChain) error {
	out, err := iptables.run(ipv6, "-N", chain.Name)

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
	out, err := iptables.run(ipv6, "-S", chain.Name)
	if err != nil {
		return util.EncodeOutputToError(out, err)
	}

	if strings.Contains(string(out), rule) {
		// The required rule already exist in chain
		return nil
	}

	// TODO: lock on every container but not on chaos-daemon's `/run/xtables.lock`
	out, err = iptables.run(ipv6, strings.Split(rule, " ")...)
	if err != nil {
		return util.EncodeOutputToError(out, err)
	}

	return nil
}

func (iptables *iptablesClient) flushIptablesChain(ipv6 bool, chain *iptablesChain) error {
	out, err := iptables.run(ipv6, "-F", chain.Name)
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
