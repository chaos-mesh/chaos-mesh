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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type DeadlineReconciler struct {
	kubeClient    client.Client
	eventRecorder record.EventRecorder
	logger        logr.Logger
}

func NewDeadlineReconciler(kubeClient client.Client, eventRecorder record.EventRecorder, logger logr.Logger) *DeadlineReconciler {
	return &DeadlineReconciler{kubeClient: kubeClient, eventRecorder: eventRecorder, logger: logger}
}

func (it *DeadlineReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()

	node := v1alpha1.WorkflowNode{}

	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if node.Spec.Deadline == nil {
		return reconcile.Result{}, nil
	}

	now := metav1.NewTime(time.Now())
	if node.Spec.Deadline.Before(&now) {

		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			node := v1alpha1.WorkflowNode{}
			err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
			if err != nil {
				return err
			}
			SetCondition(&node.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.DeadlineExceed,
				Status: corev1.ConditionTrue,
				Reason: v1alpha1.NodeDeadlineExceed,
			})
			return it.kubeClient.Update(ctx, &node)
		})

		if updateError != nil {
			return reconcile.Result{}, updateError
		}
		it.logger.Info("deadline exceed", "key", request.NamespacedName, "deadline", node.Spec.Deadline.Time)
	} else {
		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			node := v1alpha1.WorkflowNode{}
			err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
			if err != nil {
				return err
			}

			var reason string
			accomplishedCondition := GetCondition(node.Status, v1alpha1.Accomplished)
			if accomplishedCondition != nil && accomplishedCondition.Status == corev1.ConditionTrue {
				reason = v1alpha1.NodeDeadlineOmitted
			} else {
				reason = v1alpha1.NodeDeadlineExceed
			}

			SetCondition(&node.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.DeadlineExceed,
				Status: corev1.ConditionFalse,
				Reason: reason,
			})
			return it.kubeClient.Update(ctx, &node)
		})

		if updateError != nil {
			return reconcile.Result{}, updateError
		}
		duration := node.Spec.Deadline.Time.Sub(now.Time)
		it.logger.Info("deadline not exceed, requeue after a while", "key", request.NamespacedName, "deadline", node.Spec.Deadline.Time,
			"duration", duration)
		return reconcile.Result{
			RequeueAfter: duration,
		}, nil
	}

	return reconcile.Result{}, nil
}
