// Copyright 2019 PingCAP, Inc.
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

package podkill

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/reconciler"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	"github.com/pingcap/chaos-mesh/pkg/utils"
)

const (
	podKillActionMsg = "delete pod"
)

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func newReconciler(c client.Client, log logr.Logger, req ctrl.Request) *Reconciler {
	return &Reconciler{
		Client: c,
		Log:    log,
	}
}

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, log logr.Logger, req ctrl.Request) *twophase.Reconciler {
	r := newReconciler(c, log, req)
	return twophase.NewReconciler(r, r.Client, r.Log)
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, obj reconciler.InnerObject) error {
	podchaos, ok := obj.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", obj)
		return err
	}
	pods, err := utils.SelectPods(ctx, r.Client, podchaos.Spec.Selector)
	if err != nil {
		r.Log.Error(err, "fail to get selected pods")
		return err
	}
	if len(pods) == 0 {
		r.Log.Error(nil, "no pod is selected", "name", req.Name, "namespace", req.Namespace)
		return err
	}
	filteredPod, err := utils.GeneratePods(pods, podchaos.Spec.Mode, podchaos.Spec.Value)
	if err != nil {
		r.Log.Error(err, "fail to generate pods")
		return err
	}

	g := errgroup.Group{}
	for index := range filteredPod {
		pod := &filteredPod[index]
		g.Go(func() error {
			r.Log.Info("Deleting", "namespace", pod.Namespace, "name", pod.Name)

			if err := r.Delete(ctx, pod, &client.DeleteOptions{
				GracePeriodSeconds: new(int64), // PeriodSeconds has to be set specifically
			}); err != nil {
				r.Log.Error(err, "unable to delete pod")
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}
	podchaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(podchaos.Spec.Action),
			Message:   podKillActionMsg,
		}

		podchaos.Status.Experiment.Pods = append(podchaos.Status.Experiment.Pods, ps)
	}
	return nil
}

// Recover implements the reconciler.InnerReconciler.Recover
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, obj reconciler.InnerObject) error {
	return nil
}

// Object implements the reconciler.InnerReconciler.Object
func (r *Reconciler) Object() reconciler.InnerObject {
	return &v1alpha1.PodChaos{}
}
