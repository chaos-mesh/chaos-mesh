// Copyright 2020 PingCAP, Inc.
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

package stresschaos

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/common"
	"github.com/pingcap/chaos-mesh/controllers/reconciler"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	"github.com/pingcap/chaos-mesh/pkg/utils"
)

const stressChaosMsg = "stressors: %v"

// Reconciler is time-chaos reconciler
type Reconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

// Reconcile reconciles a StressChaos resource
func (r *Reconciler) Reconcile(req ctrl.Request, chaos *v1alpha1.StressChaos) (ctrl.Result, error) {
	r.Log.Info("Reconciling stresschaos")
	scheduler := chaos.GetScheduler()
	duration, err := chaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get stresschaos[%s/%s]'s duration", chaos.Namespace, chaos.Name))
		return ctrl.Result{}, err
	}
	if scheduler == nil && duration == nil {
		return r.commonStressChaos(chaos, req)
	} else if scheduler != nil && duration != nil {
		return r.scheduleStressChaos(chaos, req)
	}

	// This should be ensured by admission webhook in the future
	r.Log.Error(fmt.Errorf("stresschaos[%s/%s] spec invalid", chaos.Namespace, chaos.Name), "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, fmt.Errorf("invalid scheduler and duration")
}

func (r *Reconciler) commonStressChaos(stresschaos *v1alpha1.StressChaos, req ctrl.Request) (ctrl.Result, error) {
	cr := common.NewReconciler(r, r.Client, r.Log)
	return cr.Reconcile(req)
}

func (r *Reconciler) scheduleStressChaos(stresschaos *v1alpha1.StressChaos, req ctrl.Request) (ctrl.Result, error) {
	sr := twophase.NewReconciler(r, r.Client, r.Log)
	return sr.Reconcile(req)
}

// Apply applies time-chaos
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	stresschaos, ok := chaos.(*v1alpha1.StressChaos)
	if !ok {
		err := errors.New("chaos is not stresschaos")
		r.Log.Error(err, "chaos is not StressChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndGeneratePods(ctx, r.Client, &stresschaos.Spec)

	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	err = r.applyAllPods(ctx, pods, stresschaos)
	if err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}

	stresschaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}

	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Message:   fmt.Sprintf(stressChaosMsg, stresschaos.Spec.Stressors),
		}

		stresschaos.Status.Experiment.Pods = append(stresschaos.Status.Experiment.Pods, ps)
	}
	r.Event(stresschaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

// Recover means the reconciler recovers the chaos action
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	stresschaos, ok := chaos.(*v1alpha1.StressChaos)
	if !ok {
		err := errors.New("chaos is not StressChaos")
		r.Log.Error(err, "chaos is not StressChaos", "chaos", chaos)
		return err
	}

	err := r.cleanFinalizersAndRecover(ctx, stresschaos)
	if err != nil {
		return err
	}
	r.Event(stresschaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.StressChaos) error {
	if len(chaos.Finalizers) == 0 {
		return nil
	}

	for _, key := range chaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return err
		}

		var pod v1.Pod
		err = r.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &pod)

		if err != nil {
			if !k8serror.IsNotFound(err) {
				return err
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod, chaos)
		if err != nil {
			return err
		}

		chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
	}

	return nil
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.StressChaos) error {
	r.Log.Info("Try to recover pod", "namespace", pod.Namespace, "name", pod.Name)
	return nil
}

// Object would return the instance of chaos
func (r *Reconciler) Object() reconciler.InnerObject {
	return &v1alpha1.StressChaos{}
}

func (r *Reconciler) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.StressChaos) error {
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

func (r *Reconciler) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.StressChaos) error {
	r.Log.Info("Try to apply stress tests", "namespace", pod.Namespace, "name", pod.Name)
	return nil
}
