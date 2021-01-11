// Copyright 2019 Chaos Mesh Authors.
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

package partition

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
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/podnetworkchaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/iptable"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/finalizer"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

const (
	networkPartitionActionMsg = " partition network duration %s"
	networkChaosSourceMsg     = "This is a source pod."
	networkChaosTargetMsg     = "This is a target pod."

	sourceIPSetPostFix = "src"
	targetIPSetPostFix = "tgt"
)

type endpoint struct {
	ctx.Context
}

// Object implements the reconciler.InnerReconciler.Object
func (e *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

// Apply applies the chaos operation
func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	e.Log.Info("Applying network partition")

	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		e.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)

		return err
	}

	source := networkchaos.Namespace + "/" + networkchaos.Name
	m := podnetworkchaosmanager.New(source, e.Log, e.Client)

	sources, err := selector.SelectAndFilterPods(ctx, e.Client, e.Reader, &networkchaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)

	if err != nil {
		e.Log.Error(err, "failed to select and filter source pods")
		return err
	}

	var targets []v1.Pod

	if networkchaos.Spec.Target != nil {
		targets, err = selector.SelectAndFilterPods(ctx, e.Client, e.Reader, networkchaos.Spec.Target, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
		if err != nil {
			e.Log.Error(err, "failed to select and filter target pods")
			return err
		}
	}

	sourceSet := ipset.BuildIPSet(sources, []string{}, networkchaos, sourceIPSetPostFix, source)
	externalCidrs, err := netutils.ResolveCidrs(networkchaos.Spec.ExternalTargets)
	if err != nil {
		e.Log.Error(err, "failed to resolve external targets")
		return err
	}
	targetSet := ipset.BuildIPSet(targets, externalCidrs, networkchaos, targetIPSetPostFix, source)

	allPods := append(sources, targets...)

	type podPositionTuple struct {
		Pod      v1.Pod
		Position string
	}
	keyPodMap := make(map[types.NamespacedName]podPositionTuple)
	for index, pod := range allPods {
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

	// Set up ipset in every related pods
	for index := range allPods {
		pod := allPods[index]
		e.Log.Info("PODS", "name", pod.Name, "namespace", pod.Namespace)

		t := m.WithInit(types.NamespacedName{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})

		t.Append(sourceSet)
		t.Append(targetSet)
	}

	sourcesChains := []v1alpha1.RawIptables{}
	targetsChains := []v1alpha1.RawIptables{}
	if networkchaos.Spec.Direction == v1alpha1.To || networkchaos.Spec.Direction == v1alpha1.Both {
		sourcesChains = append(sourcesChains, v1alpha1.RawIptables{
			Name:      iptable.GenerateName(pb.Chain_OUTPUT, networkchaos),
			Direction: v1alpha1.Output,
			IPSets:    []string{targetSet.Name},
			RawRuleSource: v1alpha1.RawRuleSource{
				Source: source,
			},
		})

		targetsChains = append(targetsChains, v1alpha1.RawIptables{
			Name:      iptable.GenerateName(pb.Chain_INPUT, networkchaos),
			Direction: v1alpha1.Input,
			IPSets:    []string{sourceSet.Name},
			RawRuleSource: v1alpha1.RawRuleSource{
				Source: source,
			},
		})
	}

	if networkchaos.Spec.Direction == v1alpha1.From || networkchaos.Spec.Direction == v1alpha1.Both {
		sourcesChains = append(sourcesChains, v1alpha1.RawIptables{
			Name:      iptable.GenerateName(pb.Chain_INPUT, networkchaos),
			Direction: v1alpha1.Input,
			IPSets:    []string{targetSet.Name},
			RawRuleSource: v1alpha1.RawRuleSource{
				Source: source,
			},
		})

		targetsChains = append(targetsChains, v1alpha1.RawIptables{
			Name:      iptable.GenerateName(pb.Chain_OUTPUT, networkchaos),
			Direction: v1alpha1.Output,
			IPSets:    []string{sourceSet.Name},
			RawRuleSource: v1alpha1.RawRuleSource{
				Source: source,
			},
		})
	}
	e.Log.Info("chains prepared", "sourcesChains", sourcesChains, "targetsChains", targetsChains)

	err = e.SetChains(ctx, sources, sourcesChains, m, networkchaos)
	if err != nil {
		return err
	}

	err = e.SetChains(ctx, targets, targetsChains, m, networkchaos)
	if err != nil {
		return err
	}

	responses := m.Commit(ctx)

	networkchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(allPods))
	for _, keyErrorTuple := range responses {
		key := keyErrorTuple.Key
		err := keyErrorTuple.Err
		if err != nil {
			if err != podnetworkchaosmanager.ErrPodNotFound && err != podnetworkchaosmanager.ErrPodNotRunning {
				e.Log.Error(err, "fail to commit")
			} else {
				e.Log.Info("pod is not found or not running", "key", key)
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
			ps.Message += fmt.Sprintf(networkPartitionActionMsg, *networkchaos.Spec.Duration)
		}
		networkchaos.Status.Experiment.PodRecords = append(networkchaos.Status.Experiment.PodRecords, ps)
	}

	e.Event(networkchaos, v1.EventTypeNormal, events.ChaosInjected, "")
	return nil
}

// SetChains sets iptables chains for pods
func (e *endpoint) SetChains(ctx context.Context, pods []v1.Pod, chains []v1alpha1.RawIptables, m *podnetworkchaosmanager.PodNetworkManager, networkchaos *v1alpha1.NetworkChaos) error {
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}

		t := m.WithInit(types.NamespacedName{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})
		for _, chain := range chains {
			t.Append(chain)
		}

		networkchaos.Finalizers = finalizer.InsertFinalizer(networkchaos.Finalizers, key)

	}
	return nil
}

// Recover means the reconciler recovers the chaos action
func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		e.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}

	if err := e.cleanFinalizersAndRecover(ctx, networkchaos); err != nil {
		return err
	}
	e.Event(networkchaos, v1.EventTypeNormal, events.ChaosRecovered, "")

	return nil
}

func (e *endpoint) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.NetworkChaos) error {
	var result error

	source := chaos.Namespace + "/" + chaos.Name
	m := podnetworkchaosmanager.New(source, e.Log, e.Client)

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
				e.Log.Error(err, "fail to commit", "key", key)

				result = multierror.Append(result, err)
				continue
			}

			e.Log.Info("pod is not found or not running", "key", key)
		}

		chaos.Finalizers = finalizer.RemoveFromFinalizer(chaos.Finalizers, response.Key.String())
	}
	e.Log.Info("After recovering", "finalizers", chaos.Finalizers)

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		e.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = make([]string, 0)
		return nil
	}

	return result
}

func init() {
	router.Register("networkchaos", &v1alpha1.NetworkChaos{}, func(obj runtime.Object) bool {
		chaos, ok := obj.(*v1alpha1.NetworkChaos)
		if !ok {
			return false
		}

		return chaos.Spec.Action == v1alpha1.PartitionAction
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
