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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	if !v1alpha1.IsChaosTemplateType(node.Spec.Type) {
		return reconcile.Result{}, nil
	}

	if ConditionEqualsTo(node.Status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionTrue) {
		err := it.recoverChaos(ctx, node)
		return reconcile.Result{}, err
	}

	if !ConditionEqualsTo(node.Status, v1alpha1.ConditionChaosInjected, corev1.ConditionTrue) {
		err = it.injectChaos(ctx, node)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (it *ChaosNodeReconciler) injectChaos(ctx context.Context, node v1alpha1.WorkflowNode) error {

	chaosObject, meta, err := node.Spec.EmbedChaos.SpawnNewObject(node.Spec.Type)
	if err != nil {
		return err
	}

	meta.SetGenerateName(fmt.Sprintf("%s-", node.Name))
	meta.SetNamespace(node.Namespace)
	meta.SetOwnerReferences(append(meta.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion:         node.APIVersion,
		Kind:               node.Kind,
		Name:               node.Name,
		UID:                node.UID,
		Controller:         &isController,
		BlockOwnerDeletion: &blockOwnerDeletion,
	}))

	err = it.kubeClient.Create(ctx, chaosObject)
	if err != nil {
		it.eventRecorder.Event(&node, corev1.EventTypeWarning, v1alpha1.ChaosCRCreateFailed, "Failed to create chaos CR")
		it.logger.Error(err, "failed to create chaos")
		return nil
	}
	it.logger.Info("chaos object created", "namespace", meta.GetNamespace(), "name", meta.GetName())

	it.eventRecorder.Event(&node, corev1.EventTypeNormal, v1alpha1.ChaosCRCreated, fmt.Sprintf("Chaos CR %s/%s created", meta.GetNamespace(), meta.GetName()))

	group := chaosObject.GetObjectKind().GroupVersionKind().Group
	chaosRef := corev1.TypedLocalObjectReference{
		APIGroup: &group,
		Kind:     chaosObject.GetObjectKind().GroupVersionKind().Kind,
		Name:     meta.GetName(),
	}

	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		err := it.kubeClient.Get(ctx, types.NamespacedName{
			Namespace: node.Namespace,
			Name:      node.Name,
		}, &nodeNeedUpdate)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		nodeNeedUpdate.Status.ChaosResource = &chaosRef

		// TODO: this condition should be set by observation
		SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
			Type:   v1alpha1.ConditionChaosInjected,
			Status: corev1.ConditionTrue,
			Reason: v1alpha1.ChaosCRCreated,
		})

		return it.kubeClient.Update(ctx, &nodeNeedUpdate)

	})
	return updateError
}

func (it *ChaosNodeReconciler) recoverChaos(ctx context.Context, node v1alpha1.WorkflowNode) error {
	if node.Status.ChaosResource == nil {
		return nil
	}

	var err error
	chaosObject, err := v1alpha1.FetchChaosByTemplateType(node.Spec.Type)
	if err != nil {
		return err
	}

	err = it.kubeClient.Get(ctx,
		types.NamespacedName{Namespace: node.Namespace, Name: node.Status.ChaosResource.Name},
		chaosObject)

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
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		err := it.kubeClient.Get(ctx, types.NamespacedName{
			Namespace: node.Namespace,
			Name:      node.Name,
		}, &nodeNeedUpdate)
		if err != nil {
			return client.IgnoreNotFound(err)
		}

		nodeNeedUpdate.Status.ChaosResource = nil
		err = it.kubeClient.Update(ctx, &nodeNeedUpdate)
		return client.IgnoreNotFound(err)
	})

	return err
}
