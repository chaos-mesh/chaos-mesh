// Copyright 2026 Chaos Mesh Authors.
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

package podnetworkchaos

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func groupFixedDelayTargets(spec v1alpha1.PodNetworkChaosSpec, tcs []*pb.Tc) {
	ipsets := make(map[string]v1alpha1.RawIPSet, len(spec.IPSets))
	for _, set := range spec.IPSets {
		ipsets[set.Name] = set
	}

	type target struct {
		device   string
		protocol string
		members  string
	}
	representatives := make(map[target]string)
	for i, rule := range spec.TrafficControls {
		if rule.Type != v1alpha1.Netem || rule.Delay == nil || rule.Delay.Reorder != nil ||
			rule.Loss != nil || rule.Duplicate != nil || rule.Corrupt != nil || rule.Rate != nil || rule.Bandwidth != nil {
			continue
		}
		// Conversion has already validated the duration.
		jitter, _ := time.ParseDuration(rule.Delay.Jitter)
		if jitter != 0 {
			continue
		}
		members, ok := ipsetTargetMembers(rule.IPSet, ipsets)
		if !ok {
			continue
		}
		device := rule.Device
		if device == "" {
			device = "eth0"
		}
		key := target{device: device, protocol: tcs[i].Protocol, members: members}
		if representative, ok := representatives[key]; ok {
			// The daemon stacks rules with the same set name in one band.
			// Keep source-owned sets intact and share only the request's target.
			tcs[i].Ipset = representative
		} else {
			representatives[key] = rule.IPSet
		}
	}
}

func ipsetTargetMembers(name string, ipsets map[string]v1alpha1.RawIPSet) (string, bool) {
	set, ok := ipsets[name]
	if !ok || set.IPSetType != v1alpha1.SetIPSet {
		return "", false
	}
	members := []string{}
	for _, childName := range set.SetNames {
		child, ok := ipsets[childName]
		if !ok {
			return "", false
		}
		switch child.IPSetType {
		case v1alpha1.NetIPSet:
			for _, cidr := range child.Cidrs {
				members = append(members, "net:"+cidr)
			}
		case v1alpha1.NetPortIPSet:
			for _, entry := range child.CidrAndPorts {
				members = append(members, "net,port:"+entry.Cidr+","+strconv.Itoa(int(entry.Port)))
			}
		default:
			return "", false
		}
	}
	slices.Sort(members)
	return strings.Join(slices.Compact(members), "\n"), true
}
