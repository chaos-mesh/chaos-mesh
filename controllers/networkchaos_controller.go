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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

const (
	networkPartitionActionMsg = "partition network duration %s"
	networkTcActionMsg        = "network traffic control action duration %s"
)

// NetworkChaosReconciler reconciles a NetworkChaos object
type NetworkChaosReconciler struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

// +kubebuilder:rbac:groups=chaos-mesh.org,resources=networkchaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chaos-mesh.org,resources=networkchaos/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;watch;list

// renewChaosSelectItems renews selected items(including srcPods, targetPods, externalTargets etc) for a chaos object
func (r *NetworkChaosReconciler) renewChaosSelectItems(chaos *v1alpha1.NetworkChaos) error {
	var (
		targetPods []v1.Pod

		sourcePodStatuses []v1alpha1.PodStatus
		targetPodStatuses []v1alpha1.PodStatus
		externalCidrs     []string
	)

	ctx := context.Background()
	sourcePods, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &chaos.Spec)
	if chaos.Spec.Target != nil {
		targetPods, err = utils.SelectAndFilterPods(ctx, r.Client, r.Reader, chaos.Spec.Target)
		if err != nil {
			r.Log.Error(err, "failed to select and filter target pods")
			return err
		}
	}

	for _, pod := range sourcePods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(chaos.Spec.Action),
		}
		if chaos.Spec.Duration != nil {
			ps.Message = func(action v1alpha1.NetworkChaosAction, duration string) (msg string) {
				switch action {
				case v1alpha1.BandwidthAction, v1alpha1.NetemAction, v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
					msg = fmt.Sprintf(networkTcActionMsg, duration)
					return
				case v1alpha1.PartitionAction:
					msg = fmt.Sprintf(networkPartitionActionMsg, duration)
					return
				default:
					msg = ""
					return
				}
			}(chaos.Spec.Action, *chaos.Spec.Duration)
		}
		sourcePodStatuses = append(sourcePodStatuses, ps)
	}

	for _, pod := range targetPods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(chaos.Spec.Action),
		}
		if chaos.Spec.Duration != nil {
			ps.Message = func(action v1alpha1.NetworkChaosAction, duration string) (msg string) {
				switch action {
				case v1alpha1.BandwidthAction, v1alpha1.NetemAction, v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
					msg = fmt.Sprintf(networkTcActionMsg, duration)
					return
				case v1alpha1.PartitionAction:
					msg = fmt.Sprintf(networkPartitionActionMsg, duration)
					return
				default:
					msg = ""
					return
				}
			}(chaos.Spec.Action, *chaos.Spec.Duration)
		}
		targetPodStatuses = append(targetPodStatuses, ps)
	}

	externalCidrs, err = netutils.ResolveCidrs(chaos.Spec.ExternalTargets)
	if err != nil {
		r.Log.Error(err, "failed to resolve external targets")
		return err
	}

	// NOTE: all these three things are staging values
	chaos.Status.Experiment.StagingSourcePodRecords = sourcePodStatuses
	chaos.Status.Experiment.StagingTargetPodRecords = targetPodStatuses
	chaos.Status.Experiment.StagingExternalCIDRs = externalCidrs

	return nil
}

// Reconcile reconciles a NetworkChaos resource
func (r *NetworkChaosReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	logger := r.Log.WithValues("reconciler", "networkchaos")

	reconciler := networkchaos.Reconciler{
		Client:        r.Client,
		Reader:        r.Reader,
		EventRecorder: r.EventRecorder,
		Log:           logger,
	}

	chaos := &v1alpha1.NetworkChaos{}
	if err := r.Client.Get(context.Background(), req.NamespacedName, chaos); err != nil {
		r.Log.Error(err, "unable to get network chaos")
		return ctrl.Result{}, nil
	}

	// TODO: from now, we ONLY execute renew chaos select items for networkchaos controller
	// and for the other controllers, it wouldn't enter the renew logic since it didn't set its select items statuses.
	// it should be a TODO, and we need to clean the other controllers code.
	err = r.renewChaosSelectItems(chaos)
	if err != nil {
		return ctrl.Result{}, nil
	}

	result, err = reconciler.Reconcile(req, chaos)
	if err != nil {
		// FIXME: the error may not happens at renew stage,
		// but at the case which after renewed, e.g. re-apply action.
		if chaos.IsRenewed() {
			r.Event(chaos, v1.EventTypeWarning, utils.EventChaosRenewFailed, err.Error())
		} else if chaos.IsDeleted() || chaos.IsPaused() {
			r.Event(chaos, v1.EventTypeWarning, utils.EventChaosRecoverFailed, err.Error())
		} else {
			r.Event(chaos, v1.EventTypeWarning, utils.EventChaosInjectFailed, err.Error())
		}
	}

	return result, nil

}

// SetupWithManager setup networkchaos reconciler which called by controller-manager
func (r *NetworkChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	podToChaosMapFn := handler.ToRequestsFunc(
		func(a handler.MapObject) []reconcile.Request {
			reqs := []reconcile.Request{}

			pod, ok := a.Object.(*v1.Pod)
			if !ok {
				return reqs
			}

			associateNetworkChaos, err := utils.SelectAndFilterNetworkChaosByPod(context.Background(), r.Client, pod)
			if err != nil {
				r.Log.Error(err, "error filter networkchaos by pod")
				return reqs
			}

			for _, chaos := range associateNetworkChaos {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      chaos.GetObjectMeta().GetName(),
						Namespace: chaos.GetObjectMeta().GetNamespace(),
					},
				})
			}

			return reqs
		})

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NetworkChaos{}).
		Watches(&source.Kind{Type: &v1.Pod{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: podToChaosMapFn,
		}). // NOTE: we need to subscribe pod events to sync networkchaos
		Complete(r)
}
