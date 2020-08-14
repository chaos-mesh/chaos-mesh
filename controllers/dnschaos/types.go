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

package dnschaos

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type Reconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

// Reconcile reconciles a DNSChaos resource
func (r *Reconciler) Reconcile(req ctrl.Request, chaos *v1alpha1.DNSChaos) (ctrl.Result, error) {
	r.Log.Info("Reconciling dnschaos")

	/*
		scheduler := chaos.GetScheduler()
		duration, err := chaos.GetDuration()
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("unable to get dnschaos[%s/%s]'s duration", chaos.Namespace, chaos.Name))
			return ctrl.Result{}, err
		}
		if scheduler == nil && duration == nil {
			return r.commonDNSChaos(chaos, req)
		} else if scheduler != nil && duration != nil {
			return r.scheduleDNSChaos(chaos, req)
		}

		err = fmt.Errorf("dnschaos[%s/%s] spec invalid", chaos.Namespace, chaos.Name)
		// This should be ensured by admission webhook in the future
		r.Log.Error(err, "scheduler and duration should be omitted or defined at the same time")
	*/
	return ctrl.Result{}, nil
}

/*
func (r *Reconciler) commonDNSChaos(dnschaos *v1alpha1.DNSChaos, req ctrl.Request) (ctrl.Result, error) {
	var cr *common.Reconciler
	switch dnschaos.Spec.Action {
	case v1alpha1.NetemAction, v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		cr = netem.NewCommonReconciler(r.Client, r.Log.WithValues("action", "netem"),
			req, r.EventRecorder)
	case v1alpha1.PartitionAction:
		cr = partition.NewCommonReconciler(r.Client, r.Log.WithValues("action", "partition"),
			req, r.EventRecorder)
	case v1alpha1.BandwidthAction:
		cr = tbf.NewCommonReconciler(r.Client, r.Log.WithValues("action", "bandwidth"), req, r.EventRecorder)
	default:
		return r.invalidActionResponse(dnschaos)
	}
	return cr.Reconcile(req)
}

func (r *Reconciler) scheduleDNSChaos(dnschaos *v1alpha1.DNSChaos, req ctrl.Request) (ctrl.Result, error) {
	var sr *twophase.Reconciler
	switch dnschaos.Spec.Action {
	case v1alpha1.NetemAction, v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		sr = netem.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "netem"),
			req, r.EventRecorder)
	case v1alpha1.PartitionAction:
		sr = partition.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "partition"),
			req, r.EventRecorder)
	case v1alpha1.BandwidthAction:
		sr = tbf.NewTwoPhaseReconciler(r.Client, r.Log.WithValues("action", "bandwidth"), req, r.EventRecorder)
	default:
		return r.invalidActionResponse(dnschaos)
	}
	return sr.Reconcile(req)
}

func (r *Reconciler) invalidActionResponse(dnschaos *v1alpha1.DNSChaos) (ctrl.Result, error) {
	r.Log.Error(nil, "dnschaos action is invalid", "action", dnschaos.Spec.Action)
	return ctrl.Result{}, fmt.Errorf("invalid dnschaos action")
}
*/
