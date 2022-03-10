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

package tc

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	chaosdaemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

var log = ctrl.Log.WithName("tc")

// SetTcs makes grpc call to chaosdaemon to flush traffic control rules
func SetTcs(ctx context.Context, pbClient chaosdaemonclient.ChaosDaemonClientInterface, pod *v1.Pod, tcs []*pb.Tc) error {
	var err error

	if len(pod.Status.ContainerStatuses) == 0 {
		err = errors.Wrapf(utils.ErrContainerNotFound, "pod %s/%s has empty container status", pod.Namespace, pod.Name)

		return err
	}

	log.Info("Settings Tcs...")
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerName := containerStatus.Name
		containerID := containerStatus.ContainerID
		log.Info("attempting to set tcs", "containerName", containerName, "containerID", containerID)

		_, err = pbClient.SetTcs(ctx, &pb.TcsRequest{
			Tcs:         tcs,
			ContainerId: containerID,
			EnterNS:     true,
		})

		if err != nil {
			log.Error(err, fmt.Sprintf("error while setting tcs for container %s, id %s", containerName, containerID))
		} else {
			log.Info("Successfully set tcs")
			return nil
		}
	}

	return errors.Errorf("unable to set tcs for pod %s", pod.Name)
}
