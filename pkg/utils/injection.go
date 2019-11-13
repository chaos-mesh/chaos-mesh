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

package utils

import (
	"context"

	"github.com/pingcap/chaos-operator/api/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func SetIoInjection(ctx context.Context, c client.Client, pod *v1.Pod, ioChaos *v1alpha1.IoChaos) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var ns v1.Namespace
		if err := c.Get(ctx, types.NamespacedName{Name: pod.Namespace}, &ns); err != nil {
			return err
		}

		labels := ns.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}

		val, ok := labels[v1alpha1.WebhookNamespaceLabelKey]
		if !ok || val != v1alpha1.WebhookNamespaceLabelValue {
			labels[v1alpha1.WebhookNamespaceLabelKey] = v1alpha1.WebhookNamespaceLabelValue
			ns.SetLabels(labels)
		}

		annotations := ns.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}

		annotations[GenAnnotationKeyForWebhook(v1alpha1.WebhookPodAnnotationKey, pod.Name)] = ioChaos.Spec.ConfigName
		ns.SetAnnotations(annotations)

		if err := c.Update(ctx, &ns); err != nil {
			return err
		}

		return nil
	})
}

func UnsetIoInjection(ctx context.Context, c client.Client, pod *v1.Pod, iochaos *v1alpha1.IoChaos) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var ns v1.Namespace
		if err := c.Get(ctx, types.NamespacedName{Name: pod.Namespace}, &ns); err != nil {
			return err
		}

		annotations := ns.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}

		delete(annotations, GenAnnotationKeyForWebhook(v1alpha1.WebhookPodAnnotationKey, pod.Name))
		ns.SetAnnotations(annotations)

		labels := ns.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}

		if len(annotations) == 0 {
			delete(annotations, v1alpha1.WebhookPodAnnotationKey)
			ns.SetLabels(labels)
		}

		if err := c.Update(ctx, &ns); err != nil {
			return err
		}

		return nil
	})
}
