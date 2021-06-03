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

package ipset

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

var log = ctrl.Log.WithName("ipset")

// BuildIPSet builds an ipset with provided pod ip list
func BuildIPSet(pods []v1.Pod, externalCidrs []string, networkchaos *v1alpha1.NetworkChaos, namePostFix string, source string) v1alpha1.RawIPSet {
	name := GenerateIPSetName(networkchaos, namePostFix)
	cidrs := externalCidrs

	for _, pod := range pods {
		if len(pod.Status.PodIP) > 0 {
			cidrs = append(cidrs, netutils.IPToCidr(pod.Status.PodIP))
		}
	}

	return v1alpha1.RawIPSet{
		Name:  name,
		Cidrs: cidrs,
		RawRuleSource: v1alpha1.RawRuleSource{
			Source: source,
		},
	}
}

// GenerateIPSetName generates name for ipset
func GenerateIPSetName(networkchaos *v1alpha1.NetworkChaos, namePostFix string) string {
	return netutils.CompressName(networkchaos.Name, 27, namePostFix)
}

// FlushIPSets makes grpc calls to chaosdaemon to save ipset
func FlushIPSets(ctx context.Context, builder *chaosdaemon.ChaosDaemonClientBuilder, pod *v1.Pod, ipsets []*pb.IPSet) error {
	pbClient, err := builder.Build(ctx, pod)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
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

	return fmt.Errorf("unable to flush ip sets for pod %s", pod.Name)
}
