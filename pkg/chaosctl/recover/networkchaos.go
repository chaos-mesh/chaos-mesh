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

	"github.com/hasura/go-graphql-client"
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

type tcsRecover struct {
	client *ctrlclient.CtrlClient
}

func TcsRecover(client *ctrlclient.CtrlClient) Recover {
	return &tcsRecover{
		client: client,
	}
}

// Recover clean all tc qdisc rules
func (r *tcsRecover) Recover(ctx context.Context, pod *ctrlclient.PartialPod) error {
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

	var devices []graphql.String
	for dev := range deviceSet {
		devices = append(devices, graphql.String(dev))
	}

	if len(devices) == 0 {
		printStep("all tc rules are cleaned up")
		return nil
	} else {
		printStep(fmt.Sprintf("cleaning tc rules for device %v", devices))
	}

	var mutation struct {
		Pod struct {
			CleanTcs []string `graphql:"cleanTcs(devices: $devices)"`
		} `graphql:"pod(ns: $ns, name: $name)"`
	}

	err := r.client.QueryClient.Mutate(ctx, &mutation, map[string]interface{}{
		"devices": devices,
		"ns":      graphql.String(pod.Namespace),
		"name":    graphql.String(pod.Name),
	})

	if err != nil {
		return errors.Wrapf(err, "cleaned tc rules for device %v", devices)
	}

	if len(mutation.Pod.CleanTcs) != 0 {
		printStep(fmt.Sprintf("tc rules on device %s are cleaned up", strings.Join(mutation.Pod.CleanTcs, ",")))
	}

	return nil
}

type iptablesRecover struct {
	client *ctrlclient.CtrlClient
}

func IptablesRecover(client *ctrlclient.CtrlClient) Recover {
	return &iptablesRecover{
		client: client,
	}
}

// Recover clean all tables rules in chains CHAOS-INPUT and CHAOS-OUTPUT
func (r *iptablesRecover) Recover(ctx context.Context, pod *ctrlclient.PartialPod) error {
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

	var chains []graphql.String
	for chain, ok := range chainSet {
		if ok {
			chains = append(chains, graphql.String(chain))
		}
	}

	if len(chains) == 0 {
		printStep("all iptables rules are cleaned up")
		return nil
	} else {
		printStep(fmt.Sprintf("cleaning iptables rules for chains %v", chains))
	}

	var mutation struct {
		Pod struct {
			CleanIptables []string `graphql:"cleanIptables(chains: $chains)"`
		} `graphql:"pod(ns: $ns, name: $name)"`
	}

	err := r.client.QueryClient.Mutate(ctx, &mutation, map[string]interface{}{
		"chains": chains,
		"ns":     graphql.String(pod.Namespace),
		"name":   graphql.String(pod.Name),
	})

	if err != nil {
		return errors.Wrapf(err, "cleaned iptables rules for chains %v", chains)
	}

	if len(mutation.Pod.CleanIptables) != 0 {
		printStep(fmt.Sprintf("iptables rules in chains %s are cleaned up", strings.Join(mutation.Pod.CleanIptables, ",")))
	}

	return nil
}
