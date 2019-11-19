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

package delay

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/controllers/twophase"
	"github.com/pingcap/chaos-operator/pkg/utils"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
)

const (
	ioChaosDelayActionMsg = "delay file system io for %s"
)

func NewConciler(c client.Client, log logr.Logger, req ctrl.Request) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client: c,
			Log:    log,
		},
		Client: c,
		Log:    log,
	}
}

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func (r *Reconciler) Object() twophase.InnerObject {
	return &v1alpha1.IoChaos{}
}

func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	iochaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		err := errors.New("chaos is not IoChaos")
		r.Log.Error(err, "chaos is not IoChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndGeneratePods(ctx, r.Client, &iochaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	if err := r.delayAllPods(ctx, pods, iochaos); err != nil {
		return err
	}

	iochaos.Status.Experiment.StartTime = &metav1.Time{
		Time: time.Now(),
	}

	iochaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}
	iochaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning

	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(iochaos.Spec.Action),
			Message:   fmt.Sprintf(ioChaosDelayActionMsg, iochaos.Spec.Duration),
		}

		iochaos.Status.Experiment.Pods = append(iochaos.Status.Experiment.Pods, ps)
	}

	return nil
}

func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	iochaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		err := errors.New("chaos is not IoChaos")
		r.Log.Error(err, "chaos is not IoChaos", "chaos", chaos)
		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, iochaos); err != nil {
		return err
	}

	iochaos.Status.Experiment.EndTime = &metav1.Time{
		Time: time.Now(),
	}

	iochaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, iochaos *v1alpha1.IoChaos) error {
	if len(iochaos.Finalizers) == 0 {
		return nil
	}

	for _, key := range iochaos.Finalizers {
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

			r.Log.Info("Pod not found", "namespace", ns, "names", name)
			iochaos.Finalizers = utils.RemoveFromFinalizer(iochaos.Finalizers, key)
			continue
		}

		if err := r.recoverPod(ctx, &pod, iochaos); err != nil {
			return err
		}

		iochaos.Finalizers = utils.RemoveFromFinalizer(iochaos.Finalizers, key)
	}

	return nil
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, iochaos *v1alpha1.IoChaos) error {
	r.Log.Info("Recovering", "namespace", pod.Namespace, "name", pod.Name)

	if err := utils.UnsetIoInjection(ctx, r.Client, pod, iochaos); err != nil {
		r.Log.Error(err, "failed to unset I/O injection",
			"namespace", pod.Namespace, "name", pod.Name)
		return err
	}

	return r.Delete(ctx, pod, &client.DeleteOptions{
		GracePeriodSeconds: new(int64),
	})
}

func (r *Reconciler) delayAllPods(ctx context.Context, pods []v1.Pod, iochaos *v1alpha1.IoChaos) error {
	g := errgroup.Group{}

	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}

		iochaos.Finalizers = utils.InsertFinalizer(iochaos.Finalizers, key)

		g.Go(func() error {
			return r.delayPod(ctx, pod, iochaos)
		})

		return err
	}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return r.Update(ctx, iochaos)
	})
	if err != nil {
		r.Log.Error(err, "unable to update iochaos finalizers")
		return err
	}

	return g.Wait()
}

func (r *Reconciler) delayPod(ctx context.Context, pod *v1.Pod, iochaos *v1alpha1.IoChaos) error {
	r.Log.Info("Failing", "namespace", pod.Namespace, "name", pod.Name)

	if err := utils.SetIoInjection(ctx, r.Client, pod, iochaos); err != nil {
		r.Log.Error(err, "failed to set I/O injection",
			"namespace", pod.Namespace, "name", pod.Name)
		return err
	}

	// need to recreate pod when to inject sidecar
	time.Sleep(2 * time.Second)
	return r.Delete(ctx, pod, &client.DeleteOptions{
		GracePeriodSeconds: new(int64),
	})
}
