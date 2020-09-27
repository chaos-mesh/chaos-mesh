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

func (s *daemonServer) SetIptablesChains(ctx context.Context, req *pb.IptablesChainsRequest) (*empty.Empty, error) {
	log.Info("Set iptables chains", "request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	nsPath := GetNsPath(pid, bpm.NetNS)

	iptables := buildIptablesClient(ctx, nsPath)

	for _, chain := range req.Chains {
		err = iptables.setIptablesChain(chain)
		if err != nil {
			log.Error(err, "error while set iptables chains")
			return nil, err
		}
	}
	return &empty.Empty{}, nil
}

func (iptables *iptablesClient) setIptablesChain(chain *pb.Chain) error {
	switch chain.Command {
	case pb.Chain_NEW:
		var rule string
		if chain.Table != "" {
			rule = fmt.Sprintf("-t %s ", chain.Table)
		}
		log.Info("Create New Chain ", chain.ChainName, rule)
		err := iptables.createNewChain(&iptablesChain{
			Name: chain.ChainName,
			Rule: rule,
		})
		if err != nil {
			log.Error(err, "error while create iptables chains")
			return err
		}
	case pb.Chain_ADD:
		rule, err := parseAddChain(chain)
		log.Info("Add New rule in ", chain.ChainName, rule)
		if err != nil {
			log.Error(err, "error while add iptables chains")
			return err
		}
		if len(chain.Ipsets) > 1 {
			for _, ipset := range chain.Ipsets {
				rule = fmt.Sprintf(rule, ipset)
				err := iptables.ensureRule(&iptablesChain{
					Name: chain.ChainName,
				}, rule)
				if err != nil {
					log.Error(err, "error while add iptables chains")
					return err
				}
			}
		} else {
			err := iptables.ensureRule(&iptablesChain{
				Name: chain.ChainName,
			}, rule)
			if err != nil {
				log.Error(err, "error while add iptables chains")
				return err
			}
		}
	case pb.Chain_DELETE:
		log.Info("Delete Chain ", chain.ChainName)
		err := iptables.flushIptablesChain(&iptablesChain{
			Name: chain.ChainName,
			Rule: "",
		})
		if err != nil {
			log.Error(err, "error while add iptables chains")
			return err
		}
		err = iptables.ensureRule(&iptablesChain{
			Name: chain.ChainName,
		}, fmt.Sprintf("-X %s ", chain.ChainName))
	default:
		return fmt.Errorf("error no command in iptables chains")
	}
	return nil
}

func parseAddChain(chain *pb.Chain) (string, error) {
	var rule string
	if chain.Table != "" {
		rule += fmt.Sprintf("-t %s ", chain.Table)
	}
	if chain.ChainName != "" {
		rule += fmt.Sprintf("-A %s -p tcp ", chain.ChainName)
	} else {
		return "", fmt.Errorf("add chain but no chain name")
	}
	if chain.SourceAddress != "" {
		rule += fmt.Sprintf("-s %s ", chain.SourceAddress)
	}
	if chain.Sport != "" {
		rule += fmt.Sprintf("--sport %s ", chain.Sport)
	}
	if chain.DestinationAddress != "" {
		rule += fmt.Sprintf("-d %s ", chain.DestinationAddress)
	}
	if chain.Dport != "" {
		rule += fmt.Sprintf("--dport %s ", chain.Dport)
	}
	if chain.ToPorts != "" {
		rule += fmt.Sprintf("--to-ports %s ", chain.ToPorts)
	}
	if chain.Probability != "" {
		rule += fmt.Sprintf("--probability %s ", chain.Dport)
	}
	if chain.MarkIndex != "" {
		rule += fmt.Sprintf("--set-mark %s ", chain.MarkIndex)
	}
	if chain.IpsetsName != "" {
		var matchPart string
		if chain.Direction == pb.Chain_INPUT {
			matchPart = "src"
		} else if chain.Direction == pb.Chain_OUTPUT {
			matchPart = "dst"
		} else {
			return "", fmt.Errorf("unknown chain direction %d", chain.Direction)
		}
		if len(chain.Ipsets) > 1 {
			// The placeholder "%%s" is just a real percent symbol , and it works
			// outside the function to generate rules of several ipsets
			rule += fmt.Sprintf("-m set --match-set %s %%s -w 5 ", matchPart)
		} else if len(chain.Ipsets) == 1 {
			rule += fmt.Sprintf("-m set --match-set %s %s -w 5 ", matchPart, chain.Ipsets[0])
		} else {
			rule += fmt.Sprintf("-m set --match-set %s -w 5 ", matchPart)
		}
	}
	if chain.Action != "" {
		rule += fmt.Sprintf("-j %s ", chain.Action)
	} else {
		return "", fmt.Errorf("add chain must have chain Action like ACCEPT or a Chain name as Action")
	}
	return rule, nil
}

type iptablesClient struct {
	ctx    context.Context
	nsPath string
}

type iptablesChain struct {
	Name string
	Rule string
}

func buildIptablesClient(ctx context.Context, nsPath string) iptablesClient {
	return iptablesClient{
		ctx,
		nsPath,
	}
}

// createNewChain will cover existing chain
func (iptables *iptablesClient) createNewChain(chain *iptablesChain) error {
	cmd := bpm.DefaultProcessBuilder(iptablesCmd, strings.Split(chain.Rule+" -w -N "+chain.Name, " ")...).SetNetNS(iptables.nsPath).SetContext(iptables.ctx).Build()
	out, err := cmd.CombinedOutput()

	if (err == nil && len(out) == 0) ||
		(err != nil && strings.Contains(string(out), iptablesChainAlreadyExistErr)) {
		// Successfully create a new chain
		err := iptables.flushIptablesChain(chain)
		if err != nil {
			return err
		}
		return nil
	}
	return encodeOutputToError(out, err)
}

func (iptables *iptablesClient) ensureRule(chain *iptablesChain, rule string) error {
	cmd := bpm.DefaultProcessBuilder(iptablesCmd, "-w", "-S", chain.Name).SetNetNS(iptables.nsPath).SetContext(iptables.ctx).Build()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(out, err)
	}

	if strings.Contains(string(out), rule) {
		// The required rule already exist in chain
		return nil
	}

	// TODO: lock on every container but not on chaos-daemon's `/run/xtables.lock`
	cmd = bpm.DefaultProcessBuilder(iptablesCmd, strings.Split("-w "+rule, " ")...).SetNetNS(iptables.nsPath).SetContext(iptables.ctx).Build()
	out, err = cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(out, err)
	}

	return nil
}

func (iptables *iptablesClient) flushIptablesChain(chain *iptablesChain) error {
	cmd := bpm.DefaultProcessBuilder(iptablesCmd, "-w", "-F", chain.Name).SetNetNS(iptables.nsPath).SetContext(iptables.ctx).Build()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(out, err)
	}

	return nil
}
