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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"

	daemonClient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
)

// SetTcs makes grpc call to chaosdaemon to flush traffic control rules
func SetTcs(ctx context.Context, c client.Client, pod *v1.Pod, tcs []*pb.Tc) error {
	pbClient, err := daemonClient.NewChaosDaemonClient(ctx, c, pod, config.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.SetTcs(ctx, &pb.TcsRequest{
		Tcs:         tcs,
		ContainerId: containerID,
		// Prevent tcs is empty, used to clean up tc rules
		Device:  "eth0",
		EnterNS: true,
	})
	return err
}
