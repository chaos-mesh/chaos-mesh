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

package podfailure

import (
	"context"
	"fmt"
	"errors"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/client-go/tools/cache"
	"time"
)

const (
	// fakeImage is a not-existing image.
	fakeImage = "pingcap.com/fake-chaos-operator:latest"

	podFailureActionMsg = "pause pod duration %s"
)

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error
	now := time.Now()

	r.Log.Info("reconciling pod failure	")
	ctx := context.Background()

	var podchaos v1alpha1.PodChaos
	if err = r.Get(ctx, req.NamespacedName, &podchaos); err != nil {
		r.Log.Error(err, "unable to get podchaos")
		return ctrl.Result{}, err
	}

	duration, err := time.ParseDuration(podchaos.Spec.Duration)

	if !podchaos.DeletionTimestamp.IsZero() {
		// This chaos was deleted
		err = r.cleanFinalizersAndRecover(ctx, &podchaos)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	} else if podchaos.Spec.NextAction.Time.Before(now) {
		// Start failure action
		pods, err := utils.SelectPods(ctx, r.Client, podchaos.Spec.Selector)
		if err != nil {
			r.Log.Error(err, "fail to get selected pods")
			return ctrl.Result{}, err
		}

		if pods == nil || len(pods) == 0 {
			err = errors.New("no pod is selected")
			r.Log.Error(err, "no pod is selected")
			return ctrl.Result{}, err
		}

		filteredPod, err := utils.GeneratePods(pods, podchaos.Spec.Mode, podchaos.Spec.Value)
		if err != nil {
			r.Log.Error(err, "fail to generate pods")
			return ctrl.Result{}, err
		}

		err = r.failAllPods(ctx, filteredPod, &podchaos)
		if  err != nil  {
			return ctrl.Result{}, err
		}

		next, err := utils.NextTime(podchaos.Spec.Scheduler, now)
		if err != nil {
			return ctrl.Result{}, err
		}

		podchaos.Spec.NextAction.Time = *next
		podchaos.Status.Experiment.StartTime.Time = now
		podchaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}
		for _, pod := range pods {
			ps := v1alpha1.PodStatus{
				Namespace: pod.Namespace,
				Name:      pod.Name,
				HostIP:    pod.Status.HostIP,
				PodIP:     pod.Status.PodIP,
				Action:    string(podchaos.Spec.Action),
				Message:   podFailureActionMsg,
			}

			podchaos.Status.Experiment.Pods = append(podchaos.Status.Experiment.Pods, ps)
		}
	} else if podchaos.Status.Experiment.StartTime.Add(duration).Before(now) {
		// Start recover
		err = r.cleanFinalizersAndRecover(ctx, &podchaos)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	podchaos.Status.Experiment.EndTime.Time = now
	if err := r.Update(ctx, &podchaos); err != nil {
		r.Log.Error(err, "unable to update chaosctl status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, podchaos *v1alpha1.PodChaos) error {
	if len(podchaos.Finalizers) == 0 {
		return nil
	}

	for index, key := range podchaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return err
		}

		var pod v1.Pod
		err = r.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name: name,
		}, &pod)

		if err != nil {
			if !k8sError.IsNotFound(err) {
				return err
			}

			r.Log.Info("Pod not found", "namespace", pod.Namespace, "name", pod.Name)
			continue
		}

		err = r.recoverPod(ctx, &pod, podchaos)
		if err != nil {
			return err
		}

		podchaos.Finalizers = utils.RemoveFromFinalizer(podchaos.Finalizers, index)
	}

	return nil
}

func (r *Reconciler) failAllPods(ctx context.Context, pods []v1.Pod, podchaos *v1alpha1.PodChaos) error {
	g := errgroup.Group{}
	for _, pod := range pods {
		g.Go(func() error {
			r.Log.Info("Failing", "namespace", pod.Namespace, "name", pod.Name)

			key, err := cache.MetaNamespaceKeyFunc(&pod)
			if err != nil {
				return err
			}
			podchaos.Finalizers = append(podchaos.Finalizers, key)

			if err := r.Update(ctx, podchaos); err != nil {
				r.Log.Error(err, "unable to update podchaos finalizers")
				return err
			}

			return r.failPod(ctx, &pod, podchaos)
		})
	}

	return g.Wait()
}

func (r *Reconciler) failPod(ctx context.Context, pod *v1.Pod, podchaos *v1alpha1.PodChaos) error {
	// TODO: check the annotations or others in case that this pod is used by other chaos
	for index := range pod.Spec.Containers {
		originImage := pod.Spec.Containers[index].Image
		name := pod.Spec.Containers[index].Name

		key := utils.GenAnnotationKeyForImage(podchaos, name)
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		if _, ok := pod.Annotations[key]; ok {
			return fmt.Errorf("annotation %s exist", key)
		}
		pod.Annotations[key] =originImage
		pod.Spec.Containers[index].Image = fakeImage
	}

	if err := r.Update(ctx, pod); err != nil {
		r.Log.Error(err, "unable to use fake image on pod")
		return err
	}

	ps := v1alpha1.PodStatus{
		Namespace: pod.Namespace,
		Name:      pod.Name,
		HostIP:    pod.Status.HostIP,
		PodIP:     pod.Status.PodIP,
		Action:    string(podchaos.Spec.Action),
		Message:   podFailureActionMsg,
	}

	podchaos.Status.Experiment.Pods = append(podchaos.Status.Experiment.Pods, ps)

	return nil
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, podchaos *v1alpha1.PodChaos) error {
	for index := range pod.Spec.Containers {
		name := pod.Spec.Containers[index].Name
		annotationKey := utils.GenAnnotationKeyForImage(podchaos, name)

		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		_, ok := pod.Annotations[annotationKey]
		if !ok {
			continue // FIXME: this `continue` is absolutely useless. Check annotations and return error.
		}
	}

	// chaos-operator don't support
	return r.Delete(ctx, pod)
}