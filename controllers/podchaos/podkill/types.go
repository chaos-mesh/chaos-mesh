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
	"time"

	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/utils"
	"golang.org/x/sync/errgroup"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	podKillActionMsg = "delete pod"
)

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error
	now := time.Now()

	log := r.Log.WithValues("podkill", req.NamespacedName)
	log.Info("reconciling chaos pod")
	ctx := context.Background()

	var podchaos v1alpha1.PodChaos
	if err = r.Get(ctx, req.NamespacedName, &podchaos); err != nil {
		log.Error(err, "unable to get podchaos")
		return ctrl.Result{}, err
	}

	shouldAct := podchaos.Spec.NextAction.Time.Before(now)
	if !shouldAct {
		return ctrl.Result{RequeueAfter: podchaos.Spec.NextAction.Sub(now)}, nil
	} else {
		pods, err := utils.SelectPods(ctx, r.Client, podchaos.Spec.Selector)
		if err != nil {
			log.Error(err, "fail to get selected pods")
			return ctrl.Result{}, err
		}

		if pods == nil || len(pods) == 0 {
			err = errors.New("no pod is selected")
			log.Error(err, "no pod is selected")
			return ctrl.Result{}, err
		}

		filteredPod, err := utils.GeneratePods(pods, podchaos.Spec.Mode, podchaos.Spec.Value)
		if err != nil {
			log.Error(err, "fail to generate pods")
			return ctrl.Result{}, err
		}

		g := errgroup.Group{}
		for _, pod := range filteredPod {
			g.Go(func() error {
				log.Info("Deleting", "namespace", pod.Namespace, "name", pod.Name)

				periodSeconds := int64(0)
				if err := r.Delete(ctx, &pod, &client.DeleteOptions{
					GracePeriodSeconds: &periodSeconds, // PeriodSeconds has to be set specifically
				}); err != nil {
					log.Error(err, "unable to delete pod")
					return err
				} else {
					return nil
				}
			})
		}

		if err := g.Wait(); err != nil {
			return ctrl.Result{}, err
		} else {
			next, err := utils.NextTime(podchaos.Spec.Scheduler, now)
			if err != nil {
				return ctrl.Result{}, err
			}

			podchaos.Spec.NextAction.Time = *next

			podchaos.Status.Experiment.StartTime.Time = now
			podchaos.Status.Experiment.EndTime.Time = now

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
			if err := r.Update(ctx, &podchaos); err != nil {
				log.Error(err, "unable to update chaosctl status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	}
}
