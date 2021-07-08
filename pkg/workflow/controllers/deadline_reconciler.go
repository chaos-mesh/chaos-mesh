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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type DeadlineReconciler struct {
	*ChildNodesFetcher
	kubeClient    client.Client
	eventRecorder recorder.ChaosRecorder
	logger        logr.Logger
}

func NewDeadlineReconciler(kubeClient client.Client, eventRecorder recorder.ChaosRecorder, logger logr.Logger) *DeadlineReconciler {
	return &DeadlineReconciler{
		ChildNodesFetcher: NewChildNodesFetcher(kubeClient, logger),
		kubeClient:        kubeClient,
		eventRecorder:     eventRecorder,
		logger:            logger}
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
			nodeNeedUpdate := v1alpha1.WorkflowNode{}
			err := it.kubeClient.Get(ctx, request.NamespacedName, &nodeNeedUpdate)
			if err != nil {
				return err
			}

			if ConditionEqualsTo(nodeNeedUpdate.Status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionTrue) {
				// no need to update
				return nil
			}

			var reason string
			if ConditionEqualsTo(nodeNeedUpdate.Status, v1alpha1.ConditionAccomplished, corev1.ConditionTrue) {
				reason = v1alpha1.NodeDeadlineOmitted
			} else {
				reason = v1alpha1.NodeDeadlineExceed
			}

			if !ConditionEqualsTo(nodeNeedUpdate.Status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionTrue) && reason == v1alpha1.NodeDeadlineExceed {
				it.eventRecorder.Event(&node, recorder.DeadlineExceed{})
			}

			SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionDeadlineExceed,
				Status: corev1.ConditionTrue,
				Reason: reason,
			})

			return it.kubeClient.Status().Update(ctx, &nodeNeedUpdate)
		})

		if updateError != nil {
			return reconcile.Result{}, updateError
		}
		it.logger.Info("deadline exceed", "key", request.NamespacedName, "deadline", node.Spec.Deadline.Time)
	} else {
		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			nodeNeedUpdate := v1alpha1.WorkflowNode{}
			err := it.kubeClient.Get(ctx, request.NamespacedName, &nodeNeedUpdate)
			if err != nil {
				return err
			}

			if ConditionEqualsTo(nodeNeedUpdate.Status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionFalse) {
				// no need to update
				return nil
			}

			SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionDeadlineExceed,
				Status: corev1.ConditionFalse,
				Reason: v1alpha1.NodeDeadlineNotExceed,
			})
			return it.kubeClient.Status().Update(ctx, &nodeNeedUpdate)
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

	if ConditionEqualsTo(node.Status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionTrue) {
		// if this node deadline is exceed, try propagating to children node
		return reconcile.Result{}, it.propagateDeadlineToChildren(ctx, &node)
	}

	return reconcile.Result{}, nil
}

func (it *DeadlineReconciler) propagateDeadlineToChildren(ctx context.Context, parent *v1alpha1.WorkflowNode) error {
	switch parent.Spec.Type {
	case v1alpha1.TypeSerial, v1alpha1.TypeParallel, v1alpha1.TypeTask:
		activeChildNodes, _, err := it.ChildNodesFetcher.fetchChildNodes(ctx, *parent)
		if err != nil {
			return err
		}
		for _, childNode := range activeChildNodes {
			childNode := childNode

			if WorkflowNodeFinished(childNode.Status) {
				it.logger.V(4).Info("child node already finished, skip for propagate deadline", "node", fmt.Sprintf("%s/%s", childNode.Namespace, childNode.Name))
				continue
			}

			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				nodeNeedUpdate := v1alpha1.WorkflowNode{}
				err := it.kubeClient.Get(ctx, types.NamespacedName{
					Namespace: childNode.Namespace,
					Name:      childNode.Name,
				}, &nodeNeedUpdate)
				if err != nil {
					return err
				}
				SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
					Type:   v1alpha1.ConditionDeadlineExceed,
					Status: corev1.ConditionTrue,
					Reason: v1alpha1.ParentNodeDeadlineExceed,
				})
				it.eventRecorder.Event(&nodeNeedUpdate, recorder.ParentNodeDeadlineExceed{ParentNodeName: parent.Name})
				return it.kubeClient.Status().Update(ctx, &nodeNeedUpdate)
			})
			if err != nil {
				return err
			}
			it.logger.Info("propagate deadline for child node",
				"child node", fmt.Sprintf("%s/%s", childNode.Namespace, childNode.Name),
				"parent node", fmt.Sprintf("%s/%s", parent.Namespace, parent.Name),
			)
		}
		return nil
	default:
		it.logger.V(4).Info("no need to propagate with this type of workflow node", "type", parent.Spec.Type)
		return nil
	}
}
