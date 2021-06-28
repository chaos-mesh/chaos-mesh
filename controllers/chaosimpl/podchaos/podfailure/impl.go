// Copyright 2021 Chaos Mesh Authors.
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

	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/pkg/annotation"
)

type Impl struct {
	client.Client
}

const (
	// Always fails a container
	pauseImage = "gcr.io/google-containers/pause:latest"
)

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	podchaos := obj.(*v1alpha1.PodChaos)

	var origin v1.Pod
	err := impl.Get(ctx, controller.ParseNamespacedName(records[index].Id), &origin)
	if err != nil {
		// TODO: handle this error
		return v1alpha1.NotInjected, err
	}
	pod := origin.DeepCopy()
	for index := range pod.Spec.Containers {
		originImage := pod.Spec.Containers[index].Image
		name := pod.Spec.Containers[index].Name

		key := annotation.GenKeyForImage(podchaos, name, false)
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		// If the annotation is already existed, we could skip the reconcile for this container
		if _, ok := pod.Annotations[key]; ok {
			continue
		}
		pod.Annotations[key] = originImage
		pod.Spec.Containers[index].Image = config.ControllerCfg.PodFailurePauseImage
	}

	for index := range pod.Spec.InitContainers {
		originImage := pod.Spec.InitContainers[index].Image
		name := pod.Spec.InitContainers[index].Name

		key := annotation.GenKeyForImage(podchaos, name, true)
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}

		// If the annotation is already existed, we could skip the reconcile for this container
		if _, ok := pod.Annotations[key]; ok {
			continue
		}
		pod.Annotations[key] = originImage
		pod.Spec.InitContainers[index].Image = config.ControllerCfg.PodFailurePauseImage
	}

	err = impl.Patch(ctx, pod, client.MergeFrom(&origin))
	if err != nil {
		// TODO: handle this error
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	podchaos := obj.(*v1alpha1.PodChaos)

	var origin v1.Pod
	err := impl.Get(ctx, controller.ParseNamespacedName(records[index].Id), &origin)
	if err != nil {
		// TODO: handle this error
		if k8sError.IsNotFound(err) {
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.NotInjected, err
	}
	pod := origin.DeepCopy()
	for index := range pod.Spec.Containers {
		name := pod.Spec.Containers[index].Name
		key := annotation.GenKeyForImage(podchaos, name, false)

		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		// check annotation
		if image, ok := pod.Annotations[key]; ok {
			pod.Spec.Containers[index].Image = image
			delete(pod.Annotations, key)
		}
	}

	for index := range pod.Spec.InitContainers {
		name := pod.Spec.InitContainers[index].Name
		key := annotation.GenKeyForImage(podchaos, name, true)

		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		// check annotation
		if image, ok := pod.Annotations[key]; ok {
			pod.Spec.InitContainers[index].Image = image
			delete(pod.Annotations, key)
		}
	}

	err = impl.Patch(ctx, pod, client.MergeFrom(&origin))
	if err != nil {
		// TODO: handle this error
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client) *Impl {
	return &Impl{
		Client: c,
	}
}
