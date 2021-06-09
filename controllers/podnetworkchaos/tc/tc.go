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

package tc

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

var log = ctrl.Log.WithName("tc")

// SetTcs makes grpc call to chaosdaemon to flush traffic control rules
func SetTcs(ctx context.Context, builder *chaosdaemon.ChaosDaemonClientBuilder, pod *v1.Pod, tcs []*pb.Tc) error {
	pbClient, err := builder.Build(ctx, pod)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	log.Info("Settings Tcs...")
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerName := containerStatus.Name
		containerID := containerStatus.ContainerID
		log.Info("attempting to set tcs", "containerName", containerName, "containerID", containerID)

		_, err = pbClient.SetTcs(ctx, &pb.TcsRequest{
			Tcs:         tcs,
			ContainerId: containerID,
			// Prevent tcs is empty, used to clean up tc rules
			Device:  "eth0",
			EnterNS: true,
		})

		if err != nil {
			log.Error(err, fmt.Sprintf("error while setting tcs for container %s, id %s", containerName, containerID))
		} else {
			log.Info("Successfully set tcs")
			return nil
		}
	}

	return fmt.Errorf("unable to set tcs for pod %s", pod.Name)
}
