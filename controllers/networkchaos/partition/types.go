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

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/podnetworkmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/iptable"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

const (
	networkPartitionActionMsg = "partition network duration %s"

	sourceIPSetPostFix = "src"
	targetIPSetPostFix = "tgt"
)

func newReconciler(c client.Client, r client.Reader, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client:        c,
			Reader:        r,
			EventRecorder: recorder,
			Log:           log,
		},
		Client: c,
		Log:    log,
	}
}

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, reader client.Reader, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *twophase.Reconciler {
	r := newReconciler(c, reader, log, req, recorder)
	return twophase.NewReconciler(r, r.Client, r.Reader, r.Log)
}

// NewCommonReconciler would create Reconciler for common package
func NewCommonReconciler(c client.Client, reader client.Reader, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *common.Reconciler {
	r := newReconciler(c, reader, log, req, recorder)
	return common.NewReconciler(r, r.Client, r.Reader, r.Log)
}

// Reconciler is network partition chaos reconciler
type Reconciler struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

// Object implements the reconciler.InnerReconciler.Object
func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	r.Log.Info("Applying network partition")

	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)

		return err
	}

	source := networkchaos.Namespace + "/" + networkchaos.Name
	m := podnetworkmanager.New(source, r.Log, r.Client, r.Reader)

	sources, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &networkchaos.Spec)

	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}

	var targets []v1.Pod

	if networkchaos.Spec.Target != nil {
		targets, err = utils.SelectAndFilterPods(ctx, r.Client, r.Reader, networkchaos.Spec.Target)
		if err != nil {
			r.Log.Error(err, "failed to select and filter pods")
			return err
		}
	}

	sourceSet := ipset.BuildIPSet(sources, []string{}, networkchaos, sourceIPSetPostFix, source)
	externalCidrs, err := netutils.ResolveCidrs(networkchaos.Spec.ExternalTargets)
	if err != nil {
		r.Log.Error(err, "failed to resolve external targets")
		return err
	}
	targetSet := ipset.BuildIPSet(targets, externalCidrs, networkchaos, targetIPSetPostFix, source)

	allPods := append(sources, targets...)

	// Set up ipset in every related pods
	for index := range allPods {
		pod := allPods[index]
		r.Log.Info("PODS", "name", pod.Name, "namespace", pod.Namespace)

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
	r.Log.Info("chains prepared", "sourcesChains", sourcesChains, "targetsChains", targetsChains)

	err = r.SetChains(ctx, sources, sourcesChains, m, networkchaos)
	if err != nil {
		return err
	}

	err = r.SetChains(ctx, targets, targetsChains, m, networkchaos)
	if err != nil {
		return err
	}

	err = m.Commit(ctx)
	if err != nil {
		r.Log.Error(err, "fail to commit")
		return err
	}

	networkchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(allPods))
	for _, pod := range allPods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(networkchaos.Spec.Action),
		}

		if networkchaos.Spec.Duration != nil {
			ps.Message = fmt.Sprintf(networkPartitionActionMsg, *networkchaos.Spec.Duration)
		}

		networkchaos.Status.Experiment.PodRecords = append(networkchaos.Status.Experiment.PodRecords, ps)
	}

	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

// SetChains sets iptables chains for pods
func (r *Reconciler) SetChains(ctx context.Context, pods []v1.Pod, chains []v1alpha1.RawIptables, m *podnetworkmanager.PodNetworkManager, networkchaos *v1alpha1.NetworkChaos) error {
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

		networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, key)

	}
	return nil
}

// Recover implements the reconciler.InnerReconciler.Recover
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)

		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, networkchaos); err != nil {
		r.Log.Error(err, "cleanFinalizersAndRecover failed")
		return err
	}
	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")

	return nil
}

// Promotes means reconciler promotes staging select items to production
func (r *Reconciler) Promotes(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}
	return networkchaos.PromoteSelectItems()
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, networkchaos *v1alpha1.NetworkChaos) error {
	var result error

	source := networkchaos.Namespace + "/" + networkchaos.Name
	m := podnetworkmanager.New(source, r.Log, r.Client, r.Reader)

	for _, key := range networkchaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		_ = m.WithInit(types.NamespacedName{
			Namespace: ns,
			Name:      name,
		})

		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		err = m.Commit(ctx)
		if err != nil {
			r.Log.Error(err, "fail to commit")
		}

		networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
	}
	r.Log.Info("After recovering", "finalizers", networkchaos.Finalizers)

	if networkchaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", networkchaos)
		networkchaos.Finalizers = networkchaos.Finalizers[:0]
		return nil
	}

	return result
}
