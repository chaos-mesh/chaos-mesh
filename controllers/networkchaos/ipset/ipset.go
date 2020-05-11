// Copyright 2020 PingCAP, Inc.
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
	"crypto/sha1"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/common"
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"
)

const (
	ipsetLen = 27
)

// BuildIpSet builds an ipset with provided pod ip list
func BuildIpSet(pods []v1.Pod, networkchaos *v1alpha1.NetworkChaos, namePostFix string) pb.IpSet {
	name := generateIpSetName(networkchaos, namePostFix)
	ips := make([]string, 0, len(pods))

	for _, pod := range pods {
		if len(pod.Status.PodIP) > 0 {
			ips = append(ips, pod.Status.PodIP)
		}
	}

	return pb.IpSet{
		Name: name,
		Ips:  ips,
	}
}

func generateIpSetName(networkchaos *v1alpha1.NetworkChaos, namePostFix string) string {
	originalName := networkchaos.Name

	var ipsetName string
	if len(originalName) < 6 {
		ipsetName = originalName + "_" + namePostFix
	} else {
		namePrefix := originalName[0:5]
		nameRest := originalName[5:]

		hasher := sha1.New()
		hasher.Write([]byte(nameRest))
		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))

		// keep the length does not exceed 27
		ipsetName = namePrefix + "_" + hashValue[0:ipsetLen-7-len(namePostFix)] + "_" + namePostFix
	}

	return ipsetName
}

// FlushIpSet makes grpc calls to chaosdaemon to save ipset
func FlushIpSet(ctx context.Context, c client.Client, pod *v1.Pod, ipset pb.IpSet) error {
	pbClient, err := utils.NewChaosDaemonClient(ctx, c, pod, common.Cfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.FlushIpSet(ctx, &pb.IpSetRequest{
		Ipset:       &ipset,
		ContainerId: containerID,
	})
	return err
}
