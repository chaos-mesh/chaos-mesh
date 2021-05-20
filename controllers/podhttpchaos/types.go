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

package podhttpchaos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

// Handler applys podhttpchaos
type Handler struct {
	client.Client
	Log logr.Logger
}

// Apply flushes http chaos configuration on pod
func (h *Handler) Apply(ctx context.Context, chaos *v1alpha1.PodHttpChaos) (status int32, err error) {
	h.Log.Info("updating http chaos", "pod", chaos.Namespace+"/"+chaos.Name, "spec", chaos.Spec)

	pod := &v1.Pod{}
	status = http.StatusInternalServerError

	err = h.Client.Get(ctx, types.NamespacedName{
		Name:      chaos.Name,
		Namespace: chaos.Namespace,
	}, pod)
	if err != nil {
		h.Log.Error(err, "fail to find pod")
		return
	}

	pbClient, err := chaosdaemon.NewChaosDaemonClient(ctx, h.Client, pod)
	if err != nil {
		return
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		err = fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
		return
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	rules, err := json.Marshal(chaos.Spec.Rules)
	if err != nil {
		return
	}

	input := string(rules)
	h.Log.Info("input with", "rules", input)

	proxyPorts := make([]uint32, 0, len(chaos.Spec.ProxyPorts))
	for _, port := range chaos.Spec.ProxyPorts {
		proxyPorts = append(proxyPorts, uint32(port))
	}

	res, err := pbClient.ApplyHttpChaos(ctx, &pb.ApplyHttpChaosRequest{
		Rules:       input,
		ProxyPorts:  proxyPorts,
		ContainerId: containerID,

		Instance:  chaos.Status.Pid,
		StartTime: chaos.Status.StartTime,
		EnterNS:   true,
	})
	if err != nil {
		return
	}

	status = res.StatusCode
	if status != http.StatusOK {
		err = errors.New(res.Error)
		return
	}

	chaos.Status.Pid = res.Instance
	chaos.Status.StartTime = res.StartTime
	chaos.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: pod.APIVersion,
			Kind:       pod.Kind,
			Name:       pod.Name,
			UID:        pod.UID,
		},
	}

	return
}
