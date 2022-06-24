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

package server

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

type DaemonHelper struct {
	Builder *chaosdaemon.ChaosDaemonClientBuilder
}

// GetPidFromPod returns pid given containerd ID in pod
func (h *DaemonHelper) GetPidFromPod(ctx context.Context, pod *v1.Pod) (uint32, error) {
	daemonClient, err := h.Builder.Build(ctx, pod, nil)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to craete new chaos daemon client of pod(%s/%s)", pod.Namespace, pod.Name)
	}
	defer daemonClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		err = errors.Wrapf(utils.ErrContainerNotFound, "pod %s/%s has empty container status", pod.Namespace, pod.Name)
		return 0, err
	}

	res, err := daemonClient.ContainerGetPid(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_GETPID,
		},
		ContainerId: pod.Status.ContainerStatuses[0].ContainerID,
	})
	if err != nil {
		return 0, errors.Wrapf(err, "failed get pid from pod %s/%s", pod.GetNamespace(), pod.GetName())
	}
	return res.Pid, nil
}
