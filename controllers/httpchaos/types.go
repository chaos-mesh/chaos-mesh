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

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
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

// Promotes means reconciler promotes staging select items to production
func (r *Reconciler) Promotes(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	httpFaultChaos, ok := chaos.(*v1alpha1.HTTPChaos)
	if !ok {
		err := errors.New("chaos is not HttpChaos")
		r.Log.Error(err, "chaos is not HttpChaos", "chaos", chaos)
		return err
	}
	return httpFaultChaos.PromoteSelectItems()
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
	return cr.Reconcile(httpFaultChaos, req)
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
	//TODO: The way to connect with sidecar need be discussed & It will work after the sidecar add to the repo.
	r.Log.Info("Try to inject Http chaos on pod", "namespace", pod.Namespace, "name", pod.Name)
	return nil
}
