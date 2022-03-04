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

// BuildIPSets builds IP sets with provided pod ip list.
// It returns the set IP set and other IP sets.
func BuildIPSets(pods []v1.Pod, externalCidrs []v1alpha1.CidrAndPort, networkchaos *v1alpha1.NetworkChaos, namePostFix string, source string) (v1alpha1.RawIPSet, []v1alpha1.RawIPSet) {
	setName := GenerateIPSetName(networkchaos, "set_"+namePostFix)
	netName := GenerateIPSetName(networkchaos, "net_"+namePostFix)
	netPortName := GenerateIPSetName(networkchaos, "netport_"+namePostFix)

	cidrs := []string{}
	cidrandPorts := []v1alpha1.CidrAndPort{}

	for _, cidr := range externalCidrs {
		if cidr.Port == 0 {
			cidrs = append(cidrs, cidr.Cidr)
		} else {
			cidrandPorts = append(cidrandPorts, cidr)
		}
	}

	for _, pod := range pods {
		if len(pod.Status.PodIP) > 0 {
			cidrs = append(cidrs, netutils.IPToCidr(pod.Status.PodIP))
		}
	}

	setIPSet := v1alpha1.RawIPSet{
		Name:      setName,
		IPSetType: v1alpha1.SetIPSet,
		SetNames:  []string{netName, netPortName},
		RawRuleSource: v1alpha1.RawRuleSource{
			Source: source,
		},
	}
	otherIPSets := []v1alpha1.RawIPSet{
		{
			Name:      netName,
			IPSetType: v1alpha1.NetIPSet,
			Cidrs:     cidrs,
			RawRuleSource: v1alpha1.RawRuleSource{
				Source: source,
			},
		},
		{
			Name:         netPortName,
			IPSetType:    v1alpha1.NetPortIPSet,
			CidrAndPorts: cidrandPorts,
			RawRuleSource: v1alpha1.RawRuleSource{
				Source: source,
			},
		},
	}

	return setIPSet, otherIPSets
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
