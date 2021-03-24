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

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type ChaosNodeReconciler struct {
	kubeClient    client.Client
	eventRecorder record.EventRecorder
	logger        logr.Logger
}

func NewChaosNodeReconciler(kubeClient client.Client, eventRecorder record.EventRecorder, logger logr.Logger) *ChaosNodeReconciler {
	return &ChaosNodeReconciler{kubeClient: kubeClient, eventRecorder: eventRecorder, logger: logger}
}

func (it *ChaosNodeReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	node := v1alpha1.WorkflowNode{}

	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if ConditionEqualsTo(node.Status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionTrue) {
		err := it.recoverChaos(ctx, node)
		return reconcile.Result{}, err
	}

	if availableChaos(string(node.Spec.Type)) {
		if !ConditionEqualsTo(node.Status, v1alpha1.ConditionChaosInjected, corev1.ConditionTrue) {
			err = it.injectChaos(ctx, node)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (it *ChaosNodeReconciler) injectChaos(ctx context.Context, node v1alpha1.WorkflowNode) error {
	var chaosObject runtime.Object
	var meta metav1.Object

	if node.Spec.Type == v1alpha1.TypeNetworkChaos {
		networkChaos := v1alpha1.NetworkChaos{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", node.Name),
				Namespace:    node.Namespace,
			},
			Spec: *node.Spec.NetworkChaos,
		}
		meta = networkChaos.GetObjectMeta()
		chaosObject = &networkChaos
	} else if node.Spec.Type == v1alpha1.TypePodChaos {
		podChaos := v1alpha1.PodChaos{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: fmt.Sprintf("%s-", node.Name),
				Namespace:    node.Namespace,
			},
			Spec: *node.Spec.PodChaos,
		}
		meta = podChaos.GetObjectMeta()
		chaosObject = &podChaos
	} else {
		// TODO: support other type of chaos by code generation
		it.logger.Info("unsupported chaos nodes", "key", fmt.Sprintf("%s/%s", node.Namespace, node.Name), "type", node.Spec.Type)
		return nil
	}

	meta.SetOwnerReferences(append(meta.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion:         node.APIVersion,
		Kind:               node.Kind,
		Name:               node.Name,
		UID:                node.UID,
		Controller:         &isController,
		BlockOwnerDeletion: &blockOwnerDeletion,
	}))

	err := it.kubeClient.Create(ctx, chaosObject)
	if err != nil {
		it.eventRecorder.Event(&node, corev1.EventTypeWarning, v1alpha1.ChaosCRCreateFailed, "Failed to create chaos CR")
		it.logger.Error(err, "failed to create chaos")
		return nil
	}

	it.eventRecorder.Event(&node, corev1.EventTypeNormal, v1alpha1.ChaosCRCreated, fmt.Sprintf("Chaos CR %s/%s created", meta.GetNamespace(), meta.GetName()))

	group := chaosObject.GetObjectKind().GroupVersionKind().Group
	chaosRef := corev1.TypedLocalObjectReference{
		APIGroup: &group,
		Kind:     chaosObject.GetObjectKind().GroupVersionKind().Kind,
		Name:     meta.GetName(),
	}

	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		namespace := node.Namespace
		name := node.Name
		node := v1alpha1.WorkflowNode{}
		err := it.kubeClient.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}, &node)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		node.Status.ChaosResource = &chaosRef

		// TODO: this condition should be set by observation
		SetCondition(&node.Status, v1alpha1.WorkflowNodeCondition{
			Type:   v1alpha1.ConditionChaosInjected,
			Status: corev1.ConditionTrue,
			Reason: v1alpha1.ChaosCRCreated,
		})

		return it.kubeClient.Update(ctx, &node)

	})
	return updateError
}

func (it *ChaosNodeReconciler) recoverChaos(ctx context.Context, node v1alpha1.WorkflowNode) error {
	if node.Status.ChaosResource == nil {
		return nil
	}

	var chaosObject runtime.Object

	var err error
	if node.Spec.Type == v1alpha1.TypeNetworkChaos {
		target := v1alpha1.NetworkChaos{}
		err = it.kubeClient.Get(ctx,
			types.NamespacedName{Namespace: node.Namespace, Name: node.Status.ChaosResource.Name},
			&target)
		chaosObject = &target
	} else if node.Spec.Type == v1alpha1.TypePodChaos {
		target := v1alpha1.PodChaos{}
		err = it.kubeClient.Get(ctx,
			types.NamespacedName{Namespace: node.Namespace, Name: node.Status.ChaosResource.Name},
			&target)
		chaosObject = &target
	} else {
		it.logger.Info("unsupported chaos nodes", "key", fmt.Sprintf("%s/%s", node.Namespace, node.Name), "type", node.Spec.Type)
		return nil
	}

	if apierrors.IsNotFound(err) {
		it.logger.V(4).Info("target chaos not exist", "namespace", node.Namespace, "name", node.Status.ChaosResource.Name, "chaos kind", node.Status.ChaosResource.Kind)
		return nil
	}
	if err != nil {
		return err
	}

	err = it.kubeClient.Delete(ctx, chaosObject)

	if client.IgnoreNotFound(err) != nil {
		return err
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		node := v1alpha1.WorkflowNode{}
		err := it.kubeClient.Get(ctx, types.NamespacedName{
			Namespace: node.Namespace,
			Name:      node.Name,
		}, &node)
		if err != nil {
			return client.IgnoreNotFound(err)
		}

		node.Status.ChaosResource = nil
		err = it.kubeClient.Update(ctx, &node)
		return client.IgnoreNotFound(err)
	})

	return err
}

func availableChaos(kind string) bool {
	// TODO: use generated KindMap
	return strings.Contains(strings.ToLower(kind), "chaos")
}
