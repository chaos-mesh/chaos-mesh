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

package ipset

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	chaosdaemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

var log = ctrl.Log.WithName("ipset")

// BuildIPSets builds IP sets separated by address family.
func BuildIPSets(pods []v1.Pod, externalCidrs []netutils.Target, networkchaos *v1alpha1.NetworkChaos, namePostFix string, source string) (v4Sets []v1alpha1.RawIPSet, v6Sets []v1alpha1.RawIPSet) {
	netName := GenerateIPSetName(networkchaos, "net_"+namePostFix)
	netPortName := GenerateIPSetName(networkchaos, "netport_"+namePostFix)
	net6Name := GenerateIPSetName(networkchaos, "net6_"+namePostFix)
	netPort6Name := GenerateIPSetName(networkchaos, "netport6_"+namePostFix)

	type entries struct {
		cidrs        []string
		cidrAndPorts []v1alpha1.CidrAndPort
	}
	var ipv4, ipv6 entries
	add := func(network *net.IPNet, port uint16) {
		family := &ipv4
		if network.IP.To4() == nil {
			family = &ipv6
		}
		if port == 0 {
			family.cidrs = append(family.cidrs, network.String())
		} else {
			family.cidrAndPorts = append(family.cidrAndPorts, v1alpha1.CidrAndPort{Cidr: network.String(), Port: port})
		}
	}

	for _, target := range externalCidrs {
		add(target.Network, target.Port)
	}
	for _, pod := range pods {
		for _, ip := range podIPs(pod) {
			add(netutils.IPToCidr(ip), 0)
		}
	}

	if len(ipv4.cidrs) > 0 {
		v4Sets = append(v4Sets, v1alpha1.RawIPSet{
			Name:          netName,
			IPSetType:     v1alpha1.NetIPSet,
			Cidrs:         ipv4.cidrs,
			RawRuleSource: v1alpha1.RawRuleSource{Source: source},
		})
	}
	if len(ipv4.cidrAndPorts) > 0 {
		v4Sets = append(v4Sets, v1alpha1.RawIPSet{
			Name:          netPortName,
			IPSetType:     v1alpha1.NetPortIPSet,
			CidrAndPorts:  ipv4.cidrAndPorts,
			RawRuleSource: v1alpha1.RawRuleSource{Source: source},
		})
	}

	if len(ipv6.cidrs) > 0 {
		v6Sets = append(v6Sets, v1alpha1.RawIPSet{
			Name:          net6Name,
			IPSetType:     v1alpha1.NetIPSetV6,
			Cidrs:         ipv6.cidrs,
			RawRuleSource: v1alpha1.RawRuleSource{Source: source},
		})
	}
	if len(ipv6.cidrAndPorts) > 0 {
		v6Sets = append(v6Sets, v1alpha1.RawIPSet{
			Name:          netPort6Name,
			IPSetType:     v1alpha1.NetPortIPSetV6,
			CidrAndPorts:  ipv6.cidrAndPorts,
			RawRuleSource: v1alpha1.RawRuleSource{Source: source},
		})
	}

	return
}

// FamilyFromDNSOnly reports whether the IPv4 and IPv6 target entries came exclusively from DNS resolution.
func FamilyFromDNSOnly(pods []v1.Pod, externalCidrs []netutils.Target) (v4DNSOnly bool, v6DNSOnly bool) {
	var hasV4, hasV6 bool
	var explicitV4, explicitV6 bool

	for _, pod := range pods {
		for _, ip := range podIPs(pod) {
			if ip.To4() != nil {
				hasV4 = true
				explicitV4 = true
			} else {
				hasV6 = true
				explicitV6 = true
			}
		}
	}

	for _, target := range externalCidrs {
		if target.Network == nil {
			continue
		}
		if target.Network.IP.To4() != nil {
			hasV4 = true
			if !target.FromDNS {
				explicitV4 = true
			}
		} else {
			hasV6 = true
			if !target.FromDNS {
				explicitV6 = true
			}
		}
	}

	return hasV4 && !explicitV4, hasV6 && !explicitV6
}

func podIPs(pod v1.Pod) []net.IP {
	if len(pod.Status.PodIPs) > 0 {
		ips := make([]net.IP, 0, len(pod.Status.PodIPs))
		for _, podIP := range pod.Status.PodIPs {
			if ip := net.ParseIP(podIP.IP); ip != nil {
				ips = append(ips, ip)
			}
		}
		return ips
	}
	if ip := net.ParseIP(pod.Status.PodIP); ip != nil {
		return []net.IP{ip}
	}
	return nil
}

// BuildSetIPSet builds a list:set IP set that stores the names of the given sets.
func BuildSetIPSet(sets []v1alpha1.RawIPSet, networkchaos *v1alpha1.NetworkChaos, namePostFix string, source string) v1alpha1.RawIPSet {
	name := GenerateIPSetName(networkchaos, "set_"+namePostFix)
	setNames := []string{}

	for _, set := range sets {
		setNames = append(setNames, set.Name)
	}

	return v1alpha1.RawIPSet{
		Name:          name,
		IPSetType:     v1alpha1.SetIPSet,
		SetNames:      setNames,
		RawRuleSource: v1alpha1.RawRuleSource{Source: source},
	}
}

// GenerateIPSetName generates name for ipset
func GenerateIPSetName(networkchaos *v1alpha1.NetworkChaos, namePostFix string) string {
	return netutils.CompressName(networkchaos.Name, 27, namePostFix)
}

// FlushIPSets makes grpc calls to chaosdaemon to save ipset
func FlushIPSets(ctx context.Context, pbClient chaosdaemonclient.ChaosDaemonClientInterface, pod *v1.Pod, ipsets []*pb.IPSet) error {
	var err error

	if len(pod.Status.ContainerStatuses) == 0 {
		err = errors.Wrapf(utils.ErrContainerNotFound, "pod %s/%s has empty container status", pod.Namespace, pod.Name)
		return err
	}

	log.Info("Flushing IP Sets....")
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerID := containerStatus.ContainerID
		log.Info("attempting to flush ip set", "containerID", containerID)

		_, err = pbClient.FlushIPSets(ctx, &pb.IPSetsRequest{
			Ipsets:      ipsets,
			ContainerId: containerID,
			EnterNS:     true,
			PodUid:      string(pod.UID),
		})

		if err != nil {
			log.Error(err, fmt.Sprintf("error while flushing ip sets for containerID %s", containerID))
		} else {
			log.Info("Successfully flushed ip set")
			return nil
		}
	}

	return errors.Errorf("unable to flush ip sets for pod %s", pod.Name)
}
