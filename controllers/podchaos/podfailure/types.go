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

package podfailure

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/recover"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

const (

	// Always fails a container
	pauseImage = "gcr.io/google-containers/pause:latest"

	podFailureActionMsg = "pod failure duration %s"
)

type endpoint struct {
	ctx.Context
}

type recoverer struct {
	client.Client
	Log logr.Logger
}

// Object implements the reconciler.InnerReconciler.Object
func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.PodChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {

	podchaos, ok := chaos.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &podchaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}
	err = r.failAllPods(ctx, pods, podchaos)
	if err != nil {
		return err
	}

	podchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(podchaos.Spec.Action),
		}
		if podchaos.Spec.Duration != nil {
			ps.Message = fmt.Sprintf(podFailureActionMsg, *podchaos.Spec.Duration)
		}
		podchaos.Status.Experiment.PodRecords = append(podchaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(podchaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

// Recover means the reconciler recovers the chaos action
func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	podchaos, ok := chaos.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", chaos)
		return err
	}

	rd := recover.Delegate{Client: r.Client, Log: r.Log, RecoverIntf: &recoverer{r.Client, r.Log}}

	finalizers, err := rd.CleanFinalizersAndRecover(ctx, chaos, podchaos.Finalizers, podchaos.Annotations)
	if err != nil {
		return err
	}
	podchaos.Finalizers = finalizers
	r.Event(podchaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")

	return nil
}

func (r *recoverer) RecoverPod(ctx context.Context, pod *v1.Pod, somechaos v1alpha1.InnerObject) error {
	// judged type in `Recover` already so no need to judge again
	chaos, _ := somechaos.(*v1alpha1.PodChaos)
	r.Log.Info("Recovering", "namespace", pod.Namespace, "name", pod.Name)

	for index := range pod.Spec.Containers {
		name := pod.Spec.Containers[index].Name
		_ = utils.GenAnnotationKeyForImage(chaos, name)

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

func (r *endpoint) failAllPods(ctx context.Context, pods []v1.Pod, podchaos *v1alpha1.PodChaos) error {
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

func (r *endpoint) failPod(ctx context.Context, pod *v1.Pod, podchaos *v1alpha1.PodChaos) error {
	r.Log.Info("Try to inject pod-failure", "namespace", pod.Namespace, "name", pod.Name)

	// TODO: check the annotations or others in case that this pod is used by other chaos
	for index := range pod.Spec.InitContainers {
		originImage := pod.Spec.InitContainers[index].Image
		name := pod.Spec.InitContainers[index].Name

		key := utils.GenAnnotationKeyForImage(podchaos, name)
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		// If the annotation is already existed, we could skip the reconcile for this container
		if _, ok := pod.Annotations[key]; ok {
			continue
		}
		pod.Annotations[key] = originImage
		pod.Spec.InitContainers[index].Image = pauseImage
	}

	for index := range pod.Spec.Containers {
		originImage := pod.Spec.Containers[index].Image
		name := pod.Spec.Containers[index].Name

		key := utils.GenAnnotationKeyForImage(podchaos, name)
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		// If the annotation is already existed, we could skip the reconcile for this container
		if _, ok := pod.Annotations[key]; ok {
			continue
		}
		pod.Annotations[key] = originImage
		pod.Spec.Containers[index].Image = pauseImage
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
	}
	if podchaos.Spec.Duration != nil {
		ps.Message = fmt.Sprintf(podFailureActionMsg, *podchaos.Spec.Duration)
	}

	podchaos.Status.Experiment.PodRecords = append(podchaos.Status.Experiment.PodRecords, ps)

	return nil
}

func init() {
	router.Register("podchaos", &v1alpha1.PodChaos{}, func(obj runtime.Object) bool {
		chaos, ok := obj.(*v1alpha1.PodChaos)
		if !ok {
			return false
		}

		return chaos.Spec.Action == v1alpha1.PodFailureAction
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
