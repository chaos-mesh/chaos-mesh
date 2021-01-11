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

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/iptable"
	tcpkg "github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/tc"
	pbutils "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/netem"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/netem"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	invalidNetemSpecMsg = "invalid spec for netem action, at least one is required from delay, loss, duplicate, corrupt"
)

// Handler applys podnetworkchaos
type Handler struct {
	client.Client
	client.Reader
	Log                     logr.Logger
	AllowHostNetworkTesting bool
}

// Apply flushes network configuration on pod
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

	if !h.AllowHostNetworkTesting {
		if pod.Spec.HostNetwork {
			err := errors.Errorf("it's dangerous to inject network chaos on a pod(%s/%s) with `hostNetwork`", pod.Namespace, pod.Name)
			return err
		}
	}

	err = h.SetIPSets(ctx, pod, chaos)
	if err != nil {
		return err
	}

	err = h.SetIptables(ctx, pod, chaos)
	if err != nil {
		return err
	}

	err = h.SetTcs(ctx, pod, chaos)
	if err != nil {
		return err
	}

	return nil
}

// SetIPSets sets ipset on pod
func (h *Handler) SetIPSets(ctx context.Context, pod *corev1.Pod, chaos *v1alpha1.PodNetworkChaos) error {
	ipsets := []*pb.IPSet{}
	for _, ipset := range chaos.Spec.IPSets {
		ipsets = append(ipsets, &pb.IPSet{
			Name:  ipset.Name,
			Cidrs: ipset.Cidrs,
		})
	}
	return ipset.FlushIPSets(ctx, h.Client, pod, ipsets)
}

// SetIptables sets iptables on pod
func (h *Handler) SetIptables(ctx context.Context, pod *corev1.Pod, chaos *v1alpha1.PodNetworkChaos) error {
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
			Target:    "DROP",
		})
	}
	return iptable.SetIptablesChains(ctx, h.Client, pod, chains)
}

// SetTcs sets traffic control related chaos on pod
func (h *Handler) SetTcs(ctx context.Context, pod *corev1.Pod, chaos *v1alpha1.PodNetworkChaos) error {
	tcs := []*pb.Tc{}
	for _, tc := range chaos.Spec.TrafficControls {
		if tc.Type == v1alpha1.Bandwidth {
			tbf, err := netem.FromBandwidth(tc.Bandwidth)
			if err != nil {
				return err
			}
			tcs = append(tcs, &pb.Tc{
				Type:  pb.Tc_BANDWIDTH,
				Tbf:   tbf,
				Ipset: tc.IPSet,
			})
		} else if tc.Type == v1alpha1.Netem {
			netem, err := mergeNetem(tc.TcParameter)
			if err != nil {
				return err
			}
			tcs = append(tcs, &pb.Tc{
				Type:  pb.Tc_NETEM,
				Netem: netem,
				Ipset: tc.IPSet,
			})
		} else {
			return fmt.Errorf("unknown tc type")
		}
	}

	h.Log.Info("setting tcs", "tcs", tcs)
	return tcpkg.SetTcs(ctx, h.Client, pod, tcs)
}

// NetemSpec defines the interface to convert to a Netem protobuf
type NetemSpec interface {
	ToNetem() (*pb.Netem, error)
}

// mergeNetem calls ToNetem on all non nil network emulation specs and merges them into one request.
func mergeNetem(spec v1alpha1.TcParameter) (*pb.Netem, error) {
	// NOTE: a cleaner way like
	// emSpecs = []NetemSpec{spec.Delay, spec.Loss} won't work.
	// Because in the for _, spec := range emSpecs loop,
	// spec != nil would always be true.
	// See https://stackoverflow.com/questions/13476349/check-for-nil-and-nil-interface-in-go
	// And https://groups.google.com/forum/#!topic/golang-nuts/wnH302gBa4I/discussion
	// > In short: If you never store (*T)(nil) in an interface, then you can reliably use comparison against nil
	var emSpecs []*pb.Netem
	if spec.Delay != nil {
		em, err := netem.FromDelay(spec.Delay)
		if err != nil {
			return nil, err
		}
		emSpecs = append(emSpecs, em)
	}
	if spec.Loss != nil {
		em, err := netem.FromLoss(spec.Loss)
		if err != nil {
			return nil, err
		}
		emSpecs = append(emSpecs, em)
	}
	if spec.Duplicate != nil {
		em, err := netem.FromDuplicate(spec.Duplicate)
		if err != nil {
			return nil, err
		}
		emSpecs = append(emSpecs, em)
	}
	if spec.Corrupt != nil {
		em, err := netem.FromCorrupt(spec.Corrupt)
		if err != nil {
			return nil, err
		}
		emSpecs = append(emSpecs, em)
	}
	if len(emSpecs) == 0 {
		return nil, errors.New(invalidNetemSpecMsg)
	}

	merged := &pb.Netem{}
	for _, em := range emSpecs {
		merged = pbutils.MergeNetem(merged, em)
	}
	return merged, nil
}
