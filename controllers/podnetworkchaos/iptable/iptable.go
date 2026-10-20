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

package iptable

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	chaosdaemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

var log = ctrl.Log.WithName("iptable")

// SetIptablesChains makes grpc call to chaosdaemon to flush iptable
func SetIptablesChains(ctx context.Context, pbClient chaosdaemonclient.ChaosDaemonClientInterface, pod *v1.Pod, chains []*pb.Chain) error {
	var err error

	if len(pod.Status.ContainerStatuses) == 0 {
		err = errors.Wrapf(utils.ErrContainerNotFound, "pod %s/%s has empty container status", pod.Namespace, pod.Name)

		return err
	}

	log.Info("Setting IP Tables Chains...")
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerName := containerStatus.Name
		containerID := containerStatus.ContainerID
		log.Info("attempting to set ip table chains", "containerName", containerName, "containerID", containerID)
		_, err = pbClient.SetIptablesChains(ctx, &pb.IptablesChainsRequest{
			Chains:      chains,
			ContainerId: containerID,
			EnterNS:     true,
			PodUid:      string(pod.UID),
		})

		if err != nil {
			log.Error(err, fmt.Sprintf("error while setting ip tables chains for container %s, id %s", containerName, containerID))
		} else {
			log.Info("Successfully set ip table chains")
			return nil
		}
	}

	return errors.Errorf("unable to set ip tables chains for pod %s", pod.Name)
}

// SetHostIptablesChains makes grpc call to chaosdaemon to flush iptables in the host network namespace
func SetHostIptablesChains(ctx context.Context, pbClient chaosdaemonclient.ChaosDaemonClientInterface, chains []*pb.Chain) error {
	_, err := pbClient.SetIptablesChains(ctx, &pb.IptablesChainsRequest{
		Chains: chains,
		HostNs: true,
	})
	return errors.Wrap(err, "set iptables chains in host network namespace")
}

// ToChains converts iptables rules of a PodNetworkChaos into chaos-daemon chains
func ToChains(rules []v1alpha1.RawIptables) ([]*pb.Chain, error) {
	chains := []*pb.Chain{}
	for _, rule := range rules {
		var direction pb.Chain_Direction
		switch rule.Direction {
		case v1alpha1.Input:
			direction = pb.Chain_INPUT
		case v1alpha1.Output:
			direction = pb.Chain_OUTPUT
		default:
			return nil, errors.Errorf("unknown direction %s", string(rule.Direction))
		}

		var probability float64
		if rule.DropPercent != "" {
			percent, err := strconv.ParseFloat(rule.DropPercent, 64)
			if err != nil {
				return nil, errors.Wrapf(err, "parse dropPercent of chain %s", rule.Name)
			}
			probability = percent / 100
		}

		chains = append(chains, &pb.Chain{
			Name:          rule.Name,
			Ipsets:        rule.IPSets,
			Direction:     direction,
			Target:        "DROP",
			Device:        rule.Device,
			Probability:   probability,
			ExcludeIpsets: rule.ExcludeIPSets,
		})
	}
	return chains, nil
}

// GenerateName generates chain name for network chaos
func GenerateName(direction pb.Chain_Direction, networkchaos *v1alpha1.NetworkChaos) (chainName string) {
	switch direction {
	case pb.Chain_INPUT:
		chainName = "INPUT/" + netutils.CompressName(networkchaos.Name, 21, "")
	case pb.Chain_OUTPUT:
		chainName = "OUTPUT/" + netutils.CompressName(networkchaos.Name, 20, "")
	}

	return
}
