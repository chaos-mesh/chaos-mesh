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

package podnetworkchaos

import (
	"context"
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/iptable"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	client.Client
	Log logr.Logger
}

func (h *Handler) Apply(ctx context.Context, chaos *v1alpha1.PodNetworkChaos) error {
	h.Log.Info("updating network chaos", "pod", chaos.Namespace+"/"+chaos.Name, "spec", chaos.Spec)

	pod := &corev1.Pod{}

	err := h.Client.Get(ctx, types.NamespacedName{
		Name:      chaos.Name,
		Namespace: chaos.Namespace,
	}, pod)
	if err != nil {
		h.Log.Error(err, "fail to find pod")
		return err
	}

	ipsets := []*pb.IPSet{}
	for _, ipset := range chaos.Spec.IPSets {
		ipsets = append(ipsets, &pb.IPSet{
			Name:  ipset.Name,
			Cidrs: ipset.Cidrs,
		})
	}
	ipset.FlushIPSets(ctx, h.Client, pod, ipsets)

	chains := []*pb.Chain{}
	for _, chain := range chaos.Spec.Iptables {
		var direction pb.Chain_Direction
		if chain.Direction == v1alpha1.Input {
			direction = pb.Chain_INPUT
		} else if chain.Direction == v1alpha1.Output {
			direction = pb.Chain_OUTPUT
		} else {
			err := fmt.Errorf("unknown direction %s", string(chain.Direction))
			h.Log.Error(err, "unknown direction")
			return err
		}
		chains = append(chains, &pb.Chain{
			Name:      chain.Name,
			Ipsets:    chain.IPSets,
			Direction: direction,
		})
	}
	iptable.SetIptablesChains(ctx, h.Client, pod, chains)
	return nil
}
