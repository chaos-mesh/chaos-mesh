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

package recover

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pingcap/errors"

	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

// match handle number and device in tc qdisc rules
// ### Rules Example:
// We can match handle number `8001` and device `vnet0` in the following rule:
// ```
// qdisc noqueue 8001: dev vnet0 root refcnt 2
// ```
var tcRegexp = regexp.MustCompile(`([0-9]+): dev (\w+)`)

type tcsRecoverer struct {
	client *ctrlclient.CtrlClient
}

func newTcsRecoverer(client *ctrlclient.CtrlClient) Recoverer {
	return &tcsRecoverer{
		client: client,
	}
}

// Recover clean all tc qdisc rules
func (r *tcsRecoverer) Recover(ctx context.Context, pod *PartialPod) error {
	deviceSet := map[string]bool{}
	for _, rule := range pod.TcQdisc {
		matches := tcRegexp.FindStringSubmatch(rule)
		if len(matches) != 3 {
			continue
		}
		handle, err := strconv.Atoi(matches[1])
		if err != nil {
			return errors.Wrapf(err, "parse tc qdisc handle: `%s`", matches[1])
		}
		if handle > 0 {
			// if handle is 0, it's unnecessary to clean
			deviceSet[matches[2]] = true
		}
	}

	var devices []string
	for dev := range deviceSet {
		devices = append(devices, dev)
	}

	if len(devices) == 0 {
		printStep("all tc rules are cleaned up")
		return nil
	}
	printStep(fmt.Sprintf("cleaning tc rules for device %v", devices))

	cleanedTcs, err := r.client.CleanTcs(ctx, pod.Namespace, pod.Name, devices)
	if err != nil {
		return err
	}

	if len(cleanedTcs) != 0 {
		printStep(fmt.Sprintf("tc rules on device %s are cleaned up", strings.Join(cleanedTcs, ",")))
	}

	return nil
}

type iptablesRecoverer struct {
	client *ctrlclient.CtrlClient
}

func newIptablesRecoverer(client *ctrlclient.CtrlClient) Recoverer {
	return &iptablesRecoverer{
		client: client,
	}
}

// Recover clean all tables rules in chains CHAOS-INPUT and CHAOS-OUTPUT
func (r *iptablesRecoverer) Recover(ctx context.Context, pod *PartialPod) error {
	chainSet := map[string]bool{
		"CHAOS-INPUT":  false,
		"CHAOS-OUTPUT": false,
	}
	for _, rule := range pod.Iptables {
		for chain := range chainSet {
			if strings.HasPrefix(rule, fmt.Sprintf("Chain %s", chain)) {
				chainSet[chain] = true
			}
		}
	}

	var chains []string
	for chain, ok := range chainSet {
		if ok {
			chains = append(chains, chain)
		}
	}

	if len(chains) == 0 {
		printStep("all iptables rules are cleaned up")
		return nil
	}
	printStep(fmt.Sprintf("cleaning iptables rules for chains %v", chains))

	cleanedChains, err := r.client.CleanIptables(ctx, pod.Namespace, pod.Name, chains)
	if err != nil {
		return err
	}

	if len(cleanedChains) != 0 {
		printStep(fmt.Sprintf("iptables rules in chains %s are cleaned up", strings.Join(cleanedChains, ",")))
	}

	return nil
}

type networkRecoverer struct {
	tcsRecoverer      Recoverer
	iptablesRecoverer Recoverer
}

func NetworkRecoverer(client *ctrlclient.CtrlClient) Recoverer {
	return &networkRecoverer{
		tcsRecoverer:      newTcsRecoverer(client),
		iptablesRecoverer: newIptablesRecoverer(client),
	}
}

func (r *networkRecoverer) Recover(ctx context.Context, pod *PartialPod) error {
	err := r.tcsRecoverer.Recover(ctx, pod)
	if err != nil {
		return errors.Wrap(err, "recover tcs rules")
	}
	err = r.iptablesRecoverer.Recover(ctx, pod)
	if err != nil {
		return errors.Wrap(err, "recover iptables rules")
	}
	return nil
}
