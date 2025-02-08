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
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/util"
)

const (
	ruleNotExist             = "Cannot delete qdisc with handle of zero."
	ruleNotExistLowerVersion = "RTNETLINK answers: No such file or directory"

	defaultDevice = "eth0"
)

func generateQdiscArgs(action string, qdisc *pb.Qdisc) ([]string, error) {
	if qdisc == nil {
		return nil, errors.New("qdisc is required")
	}

	if qdisc.Type == "" {
		return nil, errors.New("qdisc.Type is required")
	}

	args := []string{"qdisc", action, "dev", "eth0"}

	if qdisc.Parent == nil {
		args = append(args, "root")
	} else if qdisc.Parent.Major == 1 && qdisc.Parent.Minor == 0 {
		args = append(args, "root")
	} else {
		args = append(args, "parent", fmt.Sprintf("%d:%d", qdisc.Parent.Major, qdisc.Parent.Minor))
	}

	if qdisc.Handle == nil {
		args = append(args, "handle", fmt.Sprintf("%d:%d", 1, 0))
	} else {
		args = append(args, "handle", fmt.Sprintf("%d:%d", qdisc.Handle.Major, qdisc.Handle.Minor))
	}

	args = append(args, qdisc.Type)

	if qdisc.Args != nil {
		args = append(args, qdisc.Args...)
	}

	return args, nil
}

func getAllInterfaces(ctx context.Context, log logr.Logger, pid uint32, enterNS bool) ([]string, error) {
	var ifaces []string
	if enterNS {
		ipOutput, err := bpm.DefaultProcessBuilder("ip", "-j", "addr", "show").SetNS(pid, bpm.NetNS).SetContext(ctx).Build(ctx).CombinedOutput()
		if err != nil {
			return []string{}, err
		}
		var data []map[string]interface{}

		err = json.Unmarshal(ipOutput, &data)
		if err != nil {
			return []string{}, err
		}
		for _, iface := range data {
			name, ok := iface["ifname"]
			if !ok {
				return []string{}, errors.New("fail to read ifname from ip -j addr show")
			}
			ifaces = append(ifaces, name.(string))
		}
		log.Info("get interfaces from ip command", "ifaces", ifaces)
	} else {
		interfaces, err := net.Interfaces()
		if err != nil {
			return []string{}, errors.New("fail to read ifname from net.Interfaces()")
		}
		for _, iface := range interfaces {
			ifaces = append(ifaces, iface.Name)
		}
		log.Info("get interfaces from net.Interfaces()", "ifaces", ifaces)
	}

	return ifaces, nil
}

func (s *DaemonServer) SetTcs(ctx context.Context, in *pb.TcsRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)
	log.Info("handling tc request", "tcs", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	tcCli := buildTcClient(ctx, log, in.EnterNS, pid)

	ifaces, err := getAllInterfaces(ctx, log, pid, in.EnterNS)
	if err != nil {
		log.Error(err, "error while getting interfaces")
		return nil, err
	}
	for _, iface := range ifaces {
		err = tcCli.flush(iface)
		if err != nil {
			log.Error(err, "fail to flush tc rules on device", "device", iface)
		}
	}
	if err != nil {
		return &empty.Empty{}, err
	}

	for device, rules := range s.groupRulesAccordingToDevices(in.Tcs) {
		// tc rules are split into two different kinds according to whether it has filter.
		// all tc rules without filter are called `globalTc` and the tc rules with filter will be called `filterTc`.
		// the `globalTc` rules will be piped one by one from root, and the last `globalTc` will be connected with a PRIO
		// qdisc, which has `3 + len(filterTc)` bands. Then the 4.. bands will be connected to `filterTc` and a filter will
		// be setuped to flow packet from PRIO qdisc to it.

		// for example, four tc rules:
		// - NETEM: 50ms latency without filter
		// - NETEM: 100ms latency without filter
		// - NETEM: 50ms latency with filter ipset A
		// - NETEM: 100ms latency with filter ipset B
		// will generate tc rules:
		//	tc qdisc del dev eth0 root
		//  tc qdisc add dev eth0 root handle 1: netem delay 50000
		//  tc qdisc add dev eth0 parent 1: handle 2: netem delay 100000
		//  tc qdisc add dev eth0 parent 2: handle 3: prio bands 5 priomap 1 2 2 2 1 2 0 0 1 1 1 1 1 1 1 1
		//  tc qdisc add dev eth0 parent 3:1 handle 4: sfq
		//  tc qdisc add dev eth0 parent 3:2 handle 5: sfq
		//  tc qdisc add dev eth0 parent 3:3 handle 6: sfq
		//  tc qdisc add dev eth0 parent 3:4 handle 7: netem delay 50000
		//  iptables -A TC-TABLES-0 -o eth0 -m set --match-set A dst -j CLASSIFY --set-class 3:4 -w 5
		//  tc qdisc add dev eth0 parent 3:5 handle 8: netem delay 100000
		//  iptables -A TC-TABLES-1 -o eth0 -m set --match-set B dst -j CLASSIFY --set-class 3:5 -w 5

		globalTc := []*pb.Tc{}
		filterTc := make(map[string][]*pb.Tc)

		for _, tc := range rules {
			filter := abstractTcFilter(tc)
			if len(filter) > 0 {
				filterTc[filter] = append(filterTc[filter], tc)
				continue
			}
			globalTc = append(globalTc, tc)
		}

		if len(globalTc) > 0 {
			if err := s.setGlobalTcs(log, tcCli, globalTc, device); err != nil {
				log.Error(err, "error while setting global tc")
				return &empty.Empty{}, err
			}
		}

		if len(filterTc) > 0 {
			iptablesCli := buildIptablesClient(ctx, in.EnterNS, pid)
			if err := s.setFilterTcs(log, tcCli, iptablesCli, filterTc, device, len(globalTc)); err != nil {
				log.Error(err, "error while setting filter tc")
				return &empty.Empty{}, err
			}
		}
	}

	return &empty.Empty{}, nil
}

func (s *DaemonServer) groupRulesAccordingToDevices(tcs []*pb.Tc) map[string][]*pb.Tc {
	rules := make(map[string][]*pb.Tc)
	for _, tc := range tcs {
		if tc.Device == "" {
			tc.Device = defaultDevice
		}
		rules[tc.Device] = append(rules[tc.Device], tc)
	}
	return rules
}

func (s *DaemonServer) setGlobalTcs(log logr.Logger, cli tcClient, tcs []*pb.Tc, device string) error {
	for index, tc := range tcs {
		parentArg := "root"
		if index > 0 {
			parentArg = fmt.Sprintf("parent %d:", index)
		}

		handleArg := fmt.Sprintf("handle %d:", index+1)

		err := cli.addTc(device, parentArg, handleArg, tc)
		if err != nil {
			log.Error(err, "error while adding tc")
			return err
		}
	}

	return nil
}

func (s *DaemonServer) setFilterTcs(
	log logr.Logger,
	tcCli tcClient,
	iptablesCli iptablesClient,
	filterTc map[string][]*pb.Tc,
	device string,
	baseIndex int,
) error {
	parent := baseIndex
	band := 3 + len(filterTc) // 3 handlers for normal sfq on prio qdisc
	if err := tcCli.addPrio(device, parent, band); err != nil {
		log.Error(err, "error while adding prio")
		return err
	}

	parent++
	index := 0
	currentHandler := parent + 3 // 3 handlers for sfq on prio qdisc

	// iptables chain has been initialized by previous grpc request to set iptables
	// and iptables rules are recovered by previous call too, so there is no need
	// to remove these rules here
	chains := []*pb.Chain{}
	for _, tcs := range filterTc {
		for i, tc := range tcs {
			parentArg := fmt.Sprintf("parent %d:%d", parent, index+4)
			if i > 0 {
				parentArg = fmt.Sprintf("parent %d:", currentHandler)
			}

			currentHandler++
			handleArg := fmt.Sprintf("handle %d:", currentHandler)

			err := tcCli.addTc(device, parentArg, handleArg, tc)
			if err != nil {
				log.Error(err, "error while adding tc")
				return err
			}
		}

		ch := &pb.Chain{
			Name:      fmt.Sprintf("TC-TABLES-%d", index),
			Direction: pb.Chain_OUTPUT,
			Target:    fmt.Sprintf("CLASSIFY --set-class %d:%d", parent, index+4),
			Device:    device,
		}

		tc := tcs[0]
		if len(tc.Ipset) > 0 {
			ch.Ipsets = []string{tc.Ipset}
		}

		ch.Protocol = tc.Protocol
		ch.SourcePorts = tc.SourcePort
		ch.DestinationPorts = tc.EgressPort

		chains = append(chains, ch)

		index++
	}
	if err := iptablesCli.setIptablesChains(chains); err != nil {
		log.Error(err, "error while setting iptables")
		return err
	}

	return nil
}

type tcClient struct {
	ctx     context.Context
	log     logr.Logger
	enterNS bool
	pid     uint32
}

func buildTcClient(ctx context.Context, log logr.Logger, enterNS bool, pid uint32) tcClient {
	return tcClient{
		ctx,
		log,
		enterNS,
		pid,
	}
}

func (c *tcClient) flush(device string) error {
	processBuilder := bpm.DefaultProcessBuilder("tc", "qdisc", "del", "dev", device, "root").SetContext(c.ctx)
	if c.enterNS {
		processBuilder = processBuilder.SetNS(c.pid, bpm.NetNS)
	}
	cmd := processBuilder.Build(c.ctx)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if (!strings.Contains(string(output), ruleNotExistLowerVersion)) && (!strings.Contains(string(output), ruleNotExist)) {
			return util.EncodeOutputToError(output, err)
		}
	}
	return nil
}

func (c *tcClient) addTc(device string, parentArg string, handleArg string, tc *pb.Tc) error {
	c.log.Info("add tc", "tc", tc)

	if tc.Type == pb.Tc_BANDWIDTH {

		if tc.Tbf == nil {
			return errors.New("tbf is nil while type is BANDWIDTH")
		}
		err := c.addTbf(device, parentArg, handleArg, tc.Tbf)
		if err != nil {
			return err
		}

	} else if tc.Type == pb.Tc_NETEM {

		if tc.Netem == nil {
			return errors.New("netem is nil while type is NETEM")
		}
		err := c.addNetem(device, parentArg, handleArg, tc.Netem)
		if err != nil {
			return err
		}

	} else {
		return errors.New("unknown tc qdisc type")
	}

	return nil
}

func (c *tcClient) addPrio(device string, parent int, band int) error {
	c.log.Info("adding prio", "parent", parent)

	parentArg := "root"
	if parent > 0 {
		parentArg = fmt.Sprintf("parent %d:", parent)
	}
	args := fmt.Sprintf("qdisc add dev %s %s handle %d: prio bands %d priomap 1 2 2 2 1 2 0 0 1 1 1 1 1 1 1 1", device, parentArg, parent+1, band)

	processBuilder := bpm.DefaultProcessBuilder("tc", strings.Split(args, " ")...).SetContext(c.ctx)
	if c.enterNS {
		processBuilder = processBuilder.SetNS(c.pid, bpm.NetNS)
	}
	cmd := processBuilder.Build(c.ctx)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return util.EncodeOutputToError(output, err)
	}

	for index := 1; index <= 3; index++ {
		args := fmt.Sprintf("qdisc add dev %s parent %d:%d handle %d: sfq", device, parent+1, index, parent+1+index)

		processBuilder := bpm.DefaultProcessBuilder("tc", strings.Split(args, " ")...).SetContext(c.ctx)
		if c.enterNS {
			processBuilder = processBuilder.SetNS(c.pid, bpm.NetNS)
		}
		cmd := processBuilder.Build(c.ctx)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return util.EncodeOutputToError(output, err)
		}
	}

	return nil
}

func (c *tcClient) addNetem(device string, parent string, handle string, netem *pb.Netem) error {
	c.log.Info("adding netem", "device", device, "parent", parent, "handle", handle)

	args := fmt.Sprintf("qdisc add dev %s %s %s netem %s", device, parent, handle, convertNetemToArgs(netem))
	processBuilder := bpm.DefaultProcessBuilder("tc", strings.Split(args, " ")...).SetContext(c.ctx)
	if c.enterNS {
		processBuilder = processBuilder.SetNS(c.pid, bpm.NetNS)
	}
	cmd := processBuilder.Build(c.ctx)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return util.EncodeOutputToError(output, err)
	}
	return nil
}

func (c *tcClient) addTbf(device string, parent string, handle string, tbf *pb.Tbf) error {
	c.log.Info("adding tbf", "device", device, "parent", parent, "handle", handle)

	args := fmt.Sprintf("qdisc add dev %s %s %s tbf %s", device, parent, handle, convertTbfToArgs(tbf))
	processBuilder := bpm.DefaultProcessBuilder("tc", strings.Split(args, " ")...).SetContext(c.ctx)
	if c.enterNS {
		processBuilder = processBuilder.SetNS(c.pid, bpm.NetNS)
	}
	cmd := processBuilder.Build(c.ctx)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return util.EncodeOutputToError(output, err)
	}
	return nil
}

func convertNetemToArgs(netem *pb.Netem) string {
	args := ""
	if netem.Time > "0ms" {
		args = fmt.Sprintf("delay %s", netem.Time)
		if netem.Jitter > "0ms" {
			args = fmt.Sprintf("%s %s", args, netem.Jitter)

			if netem.DelayCorr > 0 {
				args = fmt.Sprintf("%s %f", args, netem.DelayCorr)
			}
		}

		// reordering not possible without specifying some delay
		if netem.Reorder > 0 {
			args = fmt.Sprintf("%s reorder %f", args, netem.Reorder)
			if netem.ReorderCorr > 0 {
				args = fmt.Sprintf("%s %f", args, netem.ReorderCorr)
			}

			if netem.Gap > 0 {
				args = fmt.Sprintf("%s gap %d", args, netem.Gap)
			}
		}
	}

	if netem.Limit > 0 {
		args = fmt.Sprintf("%s limit %d", args, netem.Limit)
	}

	if netem.Loss > 0 {
		args = fmt.Sprintf("%s loss %f", args, netem.Loss)
		if netem.LossCorr > 0 {
			args = fmt.Sprintf("%s %f", args, netem.LossCorr)
		}
	}

	if netem.Duplicate > 0 {
		args = fmt.Sprintf("%s duplicate %f", args, netem.Duplicate)
		if netem.DuplicateCorr > 0 {
			args = fmt.Sprintf("%s %f", args, netem.DuplicateCorr)
		}
	}

	if netem.Corrupt > 0 {
		args = fmt.Sprintf("%s corrupt %f", args, netem.Corrupt)
		if netem.CorruptCorr > 0 {
			args = fmt.Sprintf("%s %f", args, netem.CorruptCorr)
		}
	}

	if len(netem.Rate) > 0 {
		args = fmt.Sprintf("%s rate %s", args, netem.Rate)
	}

	trimedArgs := []string{}

	for _, part := range strings.Split(args, " ") {
		if len(part) > 0 {
			trimedArgs = append(trimedArgs, part)
		}
	}

	return strings.Join(trimedArgs, " ")
}

func convertTbfToArgs(tbf *pb.Tbf) string {
	args := fmt.Sprintf("rate %s burst %d", tbf.Rate, tbf.Buffer)
	if tbf.Limit > 0 {
		args = fmt.Sprintf("%s limit %d", args, tbf.Limit)
	}
	if tbf.PeakRate > 0 {
		args = fmt.Sprintf("%s peakrate %d mtu %d", args, tbf.PeakRate, tbf.MinBurst)
	}

	return args
}

func abstractTcFilter(tc *pb.Tc) string {
	filter := tc.Ipset

	if len(tc.Protocol) > 0 {
		filter += "-" + tc.Protocol
	}

	if len(tc.EgressPort) > 0 {
		filter += "-" + tc.EgressPort
	}

	if len(tc.SourcePort) > 0 {
		filter += "-" + tc.EgressPort
	}

	return filter
}
