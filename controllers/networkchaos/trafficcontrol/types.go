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

package trafficcontrol

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/iochaos/podiochaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/podnetworkchaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/finalizer"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

const (
	networkTcActionMsg    = "network traffic control action duration %s"
	networkChaosSourceMsg = "This is a source pod."
	networkChaosTargetMsg = "This is a target pod."
)

type endpoint struct {
	ctx.Context
}

// Object implements the reconciler.InnerReconciler.Object
func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	r.Log.Info("traffic control Apply", "req", req, "chaos", chaos)

	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}

	source := networkchaos.Namespace + "/" + networkchaos.Name
	m := podnetworkchaosmanager.New(source, r.Log, r.Client)

	sources, err := selector.SelectAndFilterPods(ctx, r.Client, r.Reader, &networkchaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
	if err != nil {
		r.Log.Error(err, "failed to select and filter source pods")
		return err
	}

	var targets []v1.Pod

	// We should only apply filter when we specify targets
	if networkchaos.Spec.Target != nil {
		targets, err = selector.SelectAndFilterPods(ctx, r.Client, r.Reader, networkchaos.Spec.Target, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
		if err != nil {
			r.Log.Error(err, "failed to select and filter target pods")
			return err
		}
	}

	pods := append(sources, targets...)
	type podPositionTuple struct {
		Pod      v1.Pod
		Position string
	}
	keyPodMap := make(map[types.NamespacedName]podPositionTuple)
	for index, pod := range pods {
		position := ""
		if index < len(sources) {
			position = "source"
		} else {
			position = "target"
		}
		keyPodMap[types.NamespacedName{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		}] = podPositionTuple{
			Pod:      pod,
			Position: position,
		}
	}

	externalCidrs, err := netutils.ResolveCidrs(networkchaos.Spec.ExternalTargets)
	if err != nil {
		r.Log.Error(err, "failed to resolve external targets")
		return err
	}

	switch networkchaos.Spec.Direction {
	case v1alpha1.To:
		err = r.applyTc(ctx, sources, targets, externalCidrs, m, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply traffic control", "sources", sources, "targets", targets)
			return err
		}
	case v1alpha1.From:
		err = r.applyTc(ctx, targets, sources, []string{}, m, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply traffic control", "sources", targets, "targets", sources)
			return err
		}
	case v1alpha1.Both:
		err = r.applyTc(ctx, pods, pods, externalCidrs, m, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply traffic control", "sources", pods, "targets", pods)
			return err
		}
	default:
		err = fmt.Errorf("unknown direction %s", networkchaos.Spec.Direction)
		r.Log.Error(err, "unknown direction", "direction", networkchaos.Spec.Direction)
		return err
	}

	responses := m.Commit(ctx)

	networkchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, keyErrorTuple := range responses {
		key := keyErrorTuple.Key
		err := keyErrorTuple.Err
		if err != nil {
			if err != podiochaosmanager.ErrPodNotFound && err != podiochaosmanager.ErrPodNotRunning {
				r.Log.Error(err, "fail to commit")
			} else {
				r.Log.Info("pod is not found or not running", "key", key)
			}
			return err
		}

		pod := keyPodMap[keyErrorTuple.Key]
		ps := v1alpha1.PodStatus{
			Namespace: pod.Pod.Namespace,
			Name:      pod.Pod.Name,
			HostIP:    pod.Pod.Status.HostIP,
			PodIP:     pod.Pod.Status.PodIP,
			Action:    string(networkchaos.Spec.Action),
		}
		if pod.Position == "source" {
			ps.Message = networkChaosSourceMsg
		} else {
			ps.Message = networkChaosTargetMsg
		}

		// TODO: add source, target and tc action message
		if networkchaos.Spec.Duration != nil {
			ps.Message += fmt.Sprintf(networkTcActionMsg, *networkchaos.Spec.Duration)
		}
		networkchaos.Status.Experiment.PodRecords = append(networkchaos.Status.Experiment.PodRecords, ps)
	}

	r.Event(networkchaos, v1.EventTypeNormal, events.ChaosInjected, "")
	return nil
}

// Recover means the reconciler recovers the chaos action
func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, networkchaos); err != nil {
		return err
	}
	r.Event(networkchaos, v1.EventTypeNormal, events.ChaosRecovered, "")
	return nil
}

func (r *endpoint) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.NetworkChaos) error {
	var result error

	source := chaos.Namespace + "/" + chaos.Name
	m := podnetworkchaosmanager.New(source, r.Log, r.Client)

	for _, key := range chaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		_ = m.WithInit(types.NamespacedName{
			Namespace: ns,
			Name:      name,
		})
	}
	responses := m.Commit(ctx)
	for _, response := range responses {
		key := response.Key
		err := response.Err
		// if pod not found or not running, directly return and giveup recover.
		if err != nil {
			if err != podnetworkchaosmanager.ErrPodNotFound && err != podnetworkchaosmanager.ErrPodNotRunning {
				r.Log.Error(err, "fail to commit", "key", key)

				result = multierror.Append(result, err)
				continue
			}

			r.Log.Info("pod is not found or not running", "key", key)
		}

		chaos.Finalizers = finalizer.RemoveFromFinalizer(chaos.Finalizers, response.Key.String())
	}
	r.Log.Info("After recovering", "finalizers", chaos.Finalizers)

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = make([]string, 0)
		return nil
	}

	return result
}

func (r *endpoint) applyTc(ctx context.Context, sources, targets []v1.Pod, externalTargets []string, m *podnetworkchaosmanager.PodNetworkManager, networkchaos *v1alpha1.NetworkChaos) error {
	for index := range sources {
		pod := &sources[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}

		networkchaos.Finalizers = finalizer.InsertFinalizer(networkchaos.Finalizers, key)
	}

	tcType := v1alpha1.Bandwidth
	switch networkchaos.Spec.Action {
	case v1alpha1.NetemAction, v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		tcType = v1alpha1.Netem
	case v1alpha1.BandwidthAction:
		tcType = v1alpha1.Bandwidth
	default:
		return fmt.Errorf("unknown action %s", networkchaos.Spec.Action)
	}

	// if we don't specify targets, then sources pods apply traffic control on all egress traffic
	if len(targets)+len(externalTargets) == 0 {
		r.Log.Info("apply traffic control", "sources", sources)
		for index := range sources {
			pod := &sources[index]

			t := m.WithInit(types.NamespacedName{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			})
			t.Append(v1alpha1.RawTrafficControl{
				Type:        tcType,
				TcParameter: networkchaos.Spec.TcParameter,
				Source:      m.Source,
			})
		}
		return nil
	}

	// create ipset contains all target ips
	dstIpset := ipset.BuildIPSet(targets, externalTargets, networkchaos, string(tcType)[0:5], m.Source)
	r.Log.Info("apply traffic control with filter", "sources", sources, "ipset", dstIpset)

	for index := range sources {
		pod := &sources[index]

		t := m.WithInit(types.NamespacedName{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})
		t.Append(dstIpset)
		t.Append(v1alpha1.RawTrafficControl{
			Type:        tcType,
			TcParameter: networkchaos.Spec.TcParameter,
			Source:      m.Source,
			IPSet:       dstIpset.Name,
		})
	}

	return nil
}

func init() {
	router.Register("networkchaos", &v1alpha1.NetworkChaos{}, func(obj runtime.Object) bool {
		chaos, ok := obj.(*v1alpha1.NetworkChaos)
		if !ok {
			return false
		}

		return chaos.Spec.Action == v1alpha1.BandwidthAction ||
			chaos.Spec.Action == v1alpha1.NetemAction ||
			chaos.Spec.Action == v1alpha1.DelayAction ||
			chaos.Spec.Action == v1alpha1.LossAction ||
			chaos.Spec.Action == v1alpha1.DuplicateAction ||
			chaos.Spec.Action == v1alpha1.CorruptAction

	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
