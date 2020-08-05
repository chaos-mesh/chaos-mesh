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

package tbf

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
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/podnetworkmap"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

const (
	networkTbfActionMsg = "network tbf action duration %s"
	ipsetPostFix        = "tbf"
)

type Reconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

// Object implements the reconciler.InnerReconciler.Object
func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

func newReconciler(c client.Client, log logr.Logger, req ctrl.Request, recorder record.EventRecorder) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client:        c,
			EventRecorder: recorder,
			Log:           log,
		},
		Client: c,
		Log:    log,
	}
}

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, log logr.Logger, req ctrl.Request, recorder record.EventRecorder) *twophase.Reconciler {
	r := newReconciler(c, log, req, recorder)
	return twophase.NewReconciler(r, r.Client, r.Log)
}

// NewCommonReconciler would create Reconciler for common package
func NewCommonReconciler(c client.Client, log logr.Logger, req ctrl.Request, recorder record.EventRecorder) *common.Reconciler {
	r := newReconciler(c, log, req, recorder)
	return common.NewReconciler(r, r.Client, r.Log)
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	r.Log.Info("tbf Apply", "req", req, "chaos", chaos)

	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}

	source := networkchaos.Namespace + "/" + networkchaos.Name
	m := podnetworkmap.New(source, r.Log, r.Client)

	sources, err := utils.SelectAndFilterPods(ctx, r.Client, &networkchaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}

	var targets []v1.Pod

	// We should only apply filter when we specify targets
	if networkchaos.Spec.Target != nil {
		targets, err = utils.SelectAndFilterPods(ctx, r.Client, networkchaos.Spec.Target)
		if err != nil {
			r.Log.Error(err, "failed to select and filter pods")
			return err
		}
	}

	pods := append(sources, targets...)

	externalCidrs, err := netutils.ResolveCidrs(networkchaos.Spec.ExternalTargets)
	if err != nil {
		r.Log.Error(err, "failed to resolve external targets")
		return err
	}

	switch networkchaos.Spec.Direction {
	case v1alpha1.To:
		err = r.applyTbf(ctx, sources, targets, externalCidrs, m, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply tbf", "sources", sources, "targets", targets)
			return err
		}
	case v1alpha1.From:
		err = r.applyTbf(ctx, targets, sources, []string{}, m, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply tbf", "sources", targets, "targets", sources)
			return err
		}
	case v1alpha1.Both:
		err = r.applyTbf(ctx, pods, pods, externalCidrs, m, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply tbf", "sources", pods, "targets", pods)
			return err
		}
	}

	err = m.Commit(ctx)
	if err != nil {
		r.Log.Error(err, "fail to commit")
		return err
	}

	networkchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(networkchaos.Spec.Action),
		}

		if networkchaos.Spec.Duration != nil {
			ps.Message = fmt.Sprintf(networkTbfActionMsg, *networkchaos.Spec.Duration)
		}

		networkchaos.Status.Experiment.PodRecords = append(networkchaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
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
		return err
	}
	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")
	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, networkchaos *v1alpha1.NetworkChaos) error {
	var result error

	source := networkchaos.Namespace + "/" + networkchaos.Name
	m := podnetworkmap.New(source, r.Log, r.Client)

	for _, key := range networkchaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		_, err = m.GetAndClear(ctx, types.NamespacedName{
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

	if networkchaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", networkchaos)
		networkchaos.Finalizers = networkchaos.Finalizers[:0]
		return nil
	}

	return result
}

func (r *Reconciler) applyTbf(ctx context.Context, sources, targets []v1.Pod, externalTargets []string, m *podnetworkmap.PodNetworkMap, networkchaos *v1alpha1.NetworkChaos) error {
	for index := range sources {
		pod := &sources[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}

		networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, key)
	}

	// if we don't specify targets, then sources pods apply tbf on all egress traffic
	if len(targets)+len(externalTargets) == 0 {
		r.Log.Info("apply tbf", "sources", sources)
		for index := range sources {
			pod := &sources[index]

			chaos, err := m.GetAndClear(ctx, types.NamespacedName{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			})
			if err != nil {
				r.Log.Error(err, "failed to get related podnetworkchaos")
				return err
			}
			chaos.Spec.TrafficControls = append(chaos.Spec.TrafficControls, v1alpha1.RawTrafficControl{
				Type:        v1alpha1.Bandwidth,
				TcParameter: networkchaos.Spec.TcParameter,
				Source:      m.Source,
			})
		}
		return nil
	}

	// create ipset contains all target ips
	dstIpset := ipset.BuildIPSet(targets, externalTargets, networkchaos, ipsetPostFix, m.Source)
	r.Log.Info("apply tbf with filter", "sources", sources, "ipset", dstIpset)

	for index := range sources {
		pod := &sources[index]

		chaos, err := m.GetAndClear(ctx, types.NamespacedName{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})
		if err != nil {
			r.Log.Error(err, "failed to get related podnetworkchaos")
			return err
		}
		chaos.Spec.IPSets = append(chaos.Spec.IPSets, dstIpset)
		chaos.Spec.TrafficControls = append(chaos.Spec.TrafficControls, v1alpha1.RawTrafficControl{
			Type:        v1alpha1.Bandwidth,
			TcParameter: networkchaos.Spec.TcParameter,
			Source:      m.Source,
			IPSet:       dstIpset.Name,
		})
	}

	return nil
}
