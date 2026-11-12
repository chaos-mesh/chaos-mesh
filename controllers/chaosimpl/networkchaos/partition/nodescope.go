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

package partition

import (
	"context"
	"net"
	"sort"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/iptable"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/nodescoped"
)

const (
	localIPSetPostFix = "local"

	// The cilium host IP annotations hold the address of cilium_host, which
	// kubelet probes and other host-originated traffic to pods come from.
	ciliumHostIPAnnotation       = "network.cilium.io/ipv4-cilium-host"
	legacyCiliumHostIPAnnotation = "io.cilium.network.ipv4-cilium-host"
)

func isNodeScope(networkchaos *v1alpha1.NetworkChaos) bool {
	return networkchaos.Spec.Scope == v1alpha1.NodeScope
}

// buildLocalIPSet lists the addresses of a node and of every pod on it, so
// that traffic which stays on the node is never dropped.
func buildLocalIPSet(ctx context.Context, reader client.Reader, networkchaos *v1alpha1.NetworkChaos, nodeName string, source string) (v1alpha1.RawIPSet, error) {
	var node v1.Node
	if err := reader.Get(ctx, types.NamespacedName{Name: nodeName}, &node); err != nil {
		return v1alpha1.RawIPSet{}, errors.Wrapf(err, "get node %s", nodeName)
	}

	var pods v1.PodList
	if err := reader.List(ctx, &pods, client.MatchingFields{nodescoped.NodeNameField: nodeName}); err != nil {
		return v1alpha1.RawIPSet{}, errors.Wrapf(err, "list pods on node %s", nodeName)
	}

	return v1alpha1.RawIPSet{
		Name:      ipset.GenerateIPSetName(networkchaos, localIPSetPostFix),
		IPSetType: v1alpha1.NetIPSet,
		Cidrs:     localCidrs(&node, pods.Items),
		RawRuleSource: v1alpha1.RawRuleSource{
			Source: source,
		},
	}, nil
}

// localCidrs returns the sorted IPv4 cidrs of a node and of the pods on it
func localCidrs(node *v1.Node, pods []v1.Pod) []string {
	set := map[string]struct{}{}
	addIP := func(ip string) {
		if parsed := net.ParseIP(ip); parsed != nil && parsed.To4() != nil {
			set[netutils.IPToCidr(parsed.To4().String())] = struct{}{}
		}
	}
	addCidr := func(cidr string) {
		if _, ipNet, err := net.ParseCIDR(cidr); err == nil && ipNet.IP.To4() != nil {
			set[ipNet.String()] = struct{}{}
		}
	}

	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeInternalIP || address.Type == v1.NodeExternalIP {
			addIP(address.Address)
		}
	}
	for _, cidr := range node.Spec.PodCIDRs {
		addCidr(cidr)
	}
	addCidr(node.Spec.PodCIDR)
	addIP(node.Annotations[ciliumHostIPAnnotation])
	addIP(node.Annotations[legacyCiliumHostIPAnnotation])

	for _, pod := range pods {
		for _, podIP := range pod.Status.PodIPs {
			addIP(podIP.IP)
		}
		addIP(pod.Status.PodIP)
	}

	cidrs := make([]string, 0, len(set))
	for cidr := range set {
		cidrs = append(cidrs, cidr)
	}
	sort.Strings(cidrs)
	return cidrs
}

// appendNodeScopeDrop appends the local ipset and the partial drop chains of a
// node scope partition to a transaction.
func appendNodeScopeDrop(t interface{ Append(interface{}) error }, networkchaos *v1alpha1.NetworkChaos, local v1alpha1.RawIPSet, device string) error {
	if err := t.Append(local); err != nil {
		return err
	}

	directions := []v1alpha1.ChainDirection{}
	if networkchaos.Spec.Direction == v1alpha1.To || networkchaos.Spec.Direction == v1alpha1.Both {
		directions = append(directions, v1alpha1.Output)
	}
	if networkchaos.Spec.Direction == v1alpha1.From || networkchaos.Spec.Direction == v1alpha1.Both {
		directions = append(directions, v1alpha1.Input)
	}

	for _, direction := range directions {
		pbDirection := pb.Chain_OUTPUT
		if direction == v1alpha1.Input {
			pbDirection = pb.Chain_INPUT
		}
		err := t.Append(v1alpha1.RawIptables{
			Name:          iptable.GenerateName(pbDirection, networkchaos),
			Direction:     direction,
			Device:        device,
			DropPercent:   dropPercent(networkchaos),
			ExcludeIPSets: []string{local.Name},
			RawRuleSource: v1alpha1.RawRuleSource{
				Source: local.Source,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
