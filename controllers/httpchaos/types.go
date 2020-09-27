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

package httpchaos

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/controllers/httpchaos/tc"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/httpchaos/iptables"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

type Reconciler struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	httpFaultChaos, ok := chaos.(*v1alpha1.HTTPChaos)
	if !ok {
		err := errors.New("chaos is not HttpFaultChaos")
		r.Log.Error(err, "chaos is not HttpFaultChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &httpFaultChaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}
	if err = r.applyAllPods(ctx, pods, httpFaultChaos); err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}
	return nil
}

func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	httpFaultChaos, ok := chaos.(*v1alpha1.HTTPChaos)
	if !ok {
		err := errors.New("chaos is not HttpChaos")
		r.Log.Error(err, "chaos is not HttpChaos", "chaos", chaos)
		return err
	}
	r.Event(httpFaultChaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")
	return nil
}

func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.HTTPChaos{}
}

func (r *Reconciler) Reconcile(req ctrl.Request, chaos *v1alpha1.HTTPChaos) (ctrl.Result, error) {
	r.Log.Info("Reconciling HttpFaultChaos")
	duration, err := chaos.GetDuration()
	if err != nil {
		msg := fmt.Sprintf("unable to get iochaos[%s/%s]'s duration",
			req.Namespace, req.Name)
		r.Log.Error(err, msg)
		return ctrl.Result{}, err
	}

	if duration != nil {
		return r.commonHttpFaultChaos(chaos, req)
	}
	err = fmt.Errorf("HttpFaultChaos[%s/%s] spec invalid", req.Namespace, req.Name)
	r.Log.Error(err, "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, err
}

func (r *Reconciler) commonHttpFaultChaos(httpFaultChaos *v1alpha1.HTTPChaos, req ctrl.Request) (ctrl.Result, error) {
	cr := common.NewReconciler(r, r.Client, r.Reader, r.Log)
	return cr.Reconcile(req)
}

func (r *Reconciler) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.HTTPChaos) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		chaos.Finalizers = utils.InsertFinalizer(chaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, chaos)
		})
	}

	return g.Wait()
}

func (r *Reconciler) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.HTTPChaos) error {
	r.Log.Info("Try to inject Http chaos on pod", "namespace", pod.Namespace, "name", pod.Name)
	if chaos.Spec.Action == v1alpha1.HTTPAbortAction ||
		chaos.Spec.Action == v1alpha1.HTTPMixedAction {
		err := r.SetAbort(ctx, pod, chaos)
		if err != nil {
			return err
		}
	}
	// Warning: Traffic control is not support on wsl2
	// because of the linux kernel problem
	if chaos.Spec.Action == v1alpha1.HTTPDelayAction ||
		chaos.Spec.Action == v1alpha1.HTTPMixedAction {
		err := r.SetDelay(ctx, pod, chaos)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetIptables sets iptables on pod
// The iptables rules are:
//1: -N HTTP-CHAOS-INPUT, -N HTTP-CHAOS-OUTPUT
//2: -A INPUT -dport container_ports -j HTTP-CHAOS-INPUT
//3: -A OUTPUT -sport container_ports -j HTTP-CHAOS-OUTPUT
//4: if abort: -A HTTP-CHAOS-INPUT --probability percent -j DROP
//5: if abort: -A HTTP-CHAOS-OUTPUT --probability percent -j DROP
func (r *Reconciler) SetAbort(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.HTTPChaos) error {
	var chains []*pb.Chain
	//1: -N HTTP-CHAOS-INPUT, -N HTTP-CHAOS-OUTPUT
	inputFilterName := "HTTP-CHAOS-INPUT"
	chains = append(chains, &pb.Chain{
		Command:   pb.Chain_NEW,
		ChainName: inputFilterName,
	})
	outputFilterName := "HTTP-CHAOS-OUTPUT"
	chains = append(chains, &pb.Chain{
		Command:   pb.Chain_NEW,
		ChainName: outputFilterName,
	})
	var ports []string
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			ports = append(ports, strconv.Itoa(int(port.ContainerPort)))
		}
	}
	//2: -A INPUT -dport container_ports -j HTTP-CHAOS-INPUT
	chains = append(chains, &pb.Chain{
		Command:   pb.Chain_ADD,
		ChainName: "INPUT",
		Dport:     strings.Join(ports, ","),
		Action:    inputFilterName,
	})
	//3: -A OUTPUT -sport container_ports -j HTTP-CHAOS-OUTPUT
	chains = append(chains, &pb.Chain{
		Command:   pb.Chain_ADD,
		ChainName: "OUPUT",
		Sport:     strings.Join(ports, ","),
		Action:    outputFilterName,
	})
	//4: -A HTTP-CHAOS-INPUT --probability percent -j DROP
	chains = append(chains, &pb.Chain{
		Command:     pb.Chain_ADD,
		ChainName:   inputFilterName,
		Action:      "DROP",
		Probability: chaos.Spec.Percent,
	})
	//5: -A HTTP-CHAOS-OUTPUT --probability percent -j DROP
	chains = append(chains, &pb.Chain{
		Command:     pb.Chain_ADD,
		ChainName:   outputFilterName,
		Action:      "DROP",
		Probability: chaos.Spec.Percent,
	})
	return iptables.SetIptablesChains(ctx, r.Client, pod, chains)
}

func (r *Reconciler) SetDelay(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.HTTPChaos) error {
	var chains []*pb.Chain
	//1: -t mangle -N HTTP-CHAOS-OUTPUT
	outputFilterName := "HTTP-CHAOS-OUTPUT"
	chains = append(chains, &pb.Chain{
		Table:     "mangle",
		Command:   pb.Chain_NEW,
		ChainName: outputFilterName,
	})
	var ports []string
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			ports = append(ports, strconv.Itoa(int(port.ContainerPort)))
		}
	}

	//2: -t mangle -A OUTPUT -sport container_ports -j HTTP-CHAOS-OUTPUT
	chains = append(chains, &pb.Chain{
		Table:     "mangle",
		Command:   pb.Chain_ADD,
		ChainName: "OUPUT",
		Action:    outputFilterName,
	})

	markInit := 1

	//3: -A HTTP-CHAOS-OUTPUT --probability percent --set-mark MarkIndex -j MARK
	chains = append(chains, &pb.Chain{
		Command:     pb.Chain_ADD,
		ChainName:   outputFilterName,
		Action:      "MARK",
		MarkIndex:   strconv.Itoa(markInit),
		Probability: chaos.Spec.Percent,
	})

	//4
	err := iptables.SetIptablesChains(ctx, r.Client, pod, chains)
	if err != nil {
		return err
	}
	var tcs []*pb.Tc
	netem := &pb.Netem{}
	duration, err := chaos.GetDuration()
	if err != nil {
		return err
	}
	netem.Time = uint32(duration.Microseconds())
	tcs = append(tcs, &pb.Tc{
		Type:      pb.Tc_NETEM,
		Netem:     netem,
		MarkIndex: strconv.Itoa(markInit),
	})
	r.Log.Info("setting tcs", "tcs", tcs)
	return tc.SetTcs(ctx, r.Client, pod, tcs)
}
