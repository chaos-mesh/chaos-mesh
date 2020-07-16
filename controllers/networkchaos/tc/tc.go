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

package tc

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

// AddQdisc makes grpc call to chaosdaemon to add qdisc
func AddQdisc(ctx context.Context, c client.Client, pod *v1.Pod, qdisc *pb.Qdisc) error {
	pbClient, err := utils.NewChaosDaemonClient(ctx, c, pod, common.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}
	containerID := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.AddQdisc(ctx, &pb.QdiscRequest{
		Qdisc:       qdisc,
		ContainerId: containerID,
	})

	return err
}

// AddEmatchFilter makes grpc call to chaosdaemon to add ematch filter
func AddEmatchFilter(ctx context.Context, c client.Client, pod *v1.Pod, filter *pb.EmatchFilter) error {
	pbClient, err := utils.NewChaosDaemonClient(ctx, c, pod, common.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}
	containerID := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.AddEmatchFilter(ctx, &pb.EmatchFilterRequest{
		Filter:      filter,
		ContainerId: containerID,
	})

	return err
}

// DelQdisc makes grpc to chaosdaemon to delete tc filter
func DelQdisc(ctx context.Context, c client.Client, pod *v1.Pod, filter *pb.TcFilter) error {
	pbClient, err := utils.NewChaosDaemonClient(ctx, c, pod, common.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}
	containerID := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.DelTcFilter(ctx, &pb.TcFilterRequest{
		Filter:      filter,
		ContainerId: containerID,
	})

	return err
}
