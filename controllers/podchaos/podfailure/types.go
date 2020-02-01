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
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

const (
	// fakeImage is a not-existing image.
	fakeImage = "pingcap.com/fake-chaos-mesh:latest"

	podFailureActionMsg = "pause pod duration %s"
)

func NewReconciler(c client.Client, log logr.Logger, req ctrl.Request) twophase.Reconciler {
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
	return &v1alpha1.PodChaos{}
}

func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	podchaos, ok := chaos.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndGeneratePods(ctx, r.Client, &podchaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	err = r.failAllPods(ctx, pods, podchaos)
	if err != nil {
		return err
	}

	podchaos.Status.Experiment.StartTime = &metav1.Time{
		Time: time.Now(),
	}
	podchaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}
	podchaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning

	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(podchaos.Spec.Action),
			Message:   fmt.Sprintf(podFailureActionMsg, *podchaos.Spec.Duration),
		}

		podchaos.Status.Experiment.Pods = append(podchaos.Status.Experiment.Pods, ps)
	}

	return nil
}

func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	podchaos, ok := chaos.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", chaos)
		return err
	}

	err := r.cleanFinalizersAndRecover(ctx, podchaos)
	if err != nil {
		return err
	}
	podchaos.Status.Experiment.EndTime = &metav1.Time{
		Time: time.Now(),
	}
	podchaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, podchaos *v1alpha1.PodChaos) error {
	if len(podchaos.Finalizers) == 0 {
		return nil
	}

	for _, key := range podchaos.Finalizers {
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
			podchaos.Finalizers = utils.RemoveFromFinalizer(podchaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod, podchaos)
		if err != nil {
			return err
		}

		podchaos.Finalizers = utils.RemoveFromFinalizer(podchaos.Finalizers, key)
	}

	return nil
}

func (r *Reconciler) failAllPods(ctx context.Context, pods []v1.Pod, podchaos *v1alpha1.PodChaos) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		podchaos.Finalizers = utils.InsertFinalizer(podchaos.Finalizers, key)

		g.Go(func() error {
			return r.failPod(ctx, pod, podchaos)
		})
	}

	return g.Wait()
}

func (r *Reconciler) failPod(ctx context.Context, pod *v1.Pod, podchaos *v1alpha1.PodChaos) error {
	r.Log.Info("Try to inject pod-failure", "namespace", pod.Namespace, "name", pod.Name)

	// TODO: check the annotations or others in case that this pod is used by other chaos
	for index := range pod.Spec.InitContainers {
		originImage := pod.Spec.InitContainers[index].Image
		name := pod.Spec.InitContainers[index].Name

		key := utils.GenAnnotationKeyForImage(podchaos, name)
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		if _, ok := pod.Annotations[key]; ok {
			return fmt.Errorf("annotation %s exist", key)
		}
		pod.Annotations[key] = originImage
		pod.Spec.InitContainers[index].Image = fakeImage
	}

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
		pod.Annotations[key] = originImage
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
		Message:   fmt.Sprintf(podFailureActionMsg, *podchaos.Spec.Duration),
	}

	podchaos.Status.Experiment.Pods = append(podchaos.Status.Experiment.Pods, ps)

	return nil
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, podchaos *v1alpha1.PodChaos) error {
	r.Log.Info("Recovering", "namespace", pod.Namespace, "name", pod.Name)

	for index := range pod.Spec.Containers {
		name := pod.Spec.Containers[index].Name
		_ = utils.GenAnnotationKeyForImage(podchaos, name)

		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		// FIXME: Check annotations and return error.
	}

	// chaos-mesh don't support
	return r.Delete(ctx, pod, &client.DeleteOptions{
		GracePeriodSeconds: new(int64), // PeriodSeconds has to be set specifically
	})
}
