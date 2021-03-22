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
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SerialNodeReconciler watches on nodes which type is Serial
type SerialNodeReconciler struct {
	kubeClient    client.Client
	eventRecorder record.EventRecorder
	logger        logr.Logger
}

func (it *SerialNodeReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()

	node := v1alpha1.WorkflowNode{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// only resolve serial nodes
	if node.Spec.Type != v1alpha1.TypeSerial {
		return reconcile.Result{}, nil
	}

	// empty serial node
	if len(node.Spec.Tasks) == 0 {
		it.logger.V(4).Info("empty serial node, NOOP", "key", request.NamespacedName)
		return reconcile.Result{}, nil
	}

	// un-synced expected children
	if node.Status.ExpectedChildren == nil {
		expected := len(node.Spec.Tasks)
		node.Status.ExpectedChildren = &expected
	}

	// this node should finished
	if len(node.Status.FinishedChildren) == *node.Status.ExpectedChildren {
		if !ConditionEqualsTo(node.Status, v1alpha1.ConditionAccomplished, corev1.ConditionTrue) {
			updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				node := v1alpha1.WorkflowNode{}
				err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
				if err != nil {
					return client.IgnoreNotFound(err)
				}
				SetCondition(&node.Status, v1alpha1.WorkflowNodeCondition{
					Type:   v1alpha1.ConditionAccomplished,
					Status: corev1.ConditionTrue,
					Reason: v1alpha1.NodeAccomplished,
				})
				return it.kubeClient.Update(ctx, &node)
			})

			if updateError != nil {
				return reconcile.Result{}, updateError
			}

			it.eventRecorder.Event(&node, corev1.EventTypeNormal, v1alpha1.NodeAccomplished, "Serial node accomplished")
		}
		return reconcile.Result{}, nil
	}

	if len(node.Status.ActiveChildren) == 0 {
		it.logger.Info("schedule next", "key", request.NamespacedName, "status", node.Status)
		taskToStartup := node.Spec.Tasks[len(node.Status.FinishedChildren)]
		parentWorkflow := v1alpha1.Workflow{}

		err := it.kubeClient.Get(ctx, types.NamespacedName{
			Namespace: node.Namespace,
			Name:      node.Spec.WorkflowName,
		}, &parentWorkflow)
		if err != nil {
			it.logger.Error(err, "failed to fetch parent workflow", "key", request.NamespacedName, "workflow name", node.Spec.WorkflowName)
			return reconcile.Result{}, err
		}
		childrenNodes, err := renderNodesByTemplates(&parentWorkflow, &node, taskToStartup)
		if err != nil {
			it.logger.Error(err, "failed to render children childrenNodes", "node", request.NamespacedName)
			return reconcile.Result{}, err
		}

		for _, childNode := range childrenNodes {
			err := it.kubeClient.Create(ctx, childNode)
			if err != nil {
				it.logger.Error(err, "failed to create child node", "node", request.NamespacedName, "child node", childNode)
				return reconcile.Result{}, err
			}
		}

		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			node := v1alpha1.WorkflowNode{}
			err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
			if err != nil {
				return err
			}
			for _, item := range childrenNodes {
				node.Status.ActiveChildren = append(node.Status.ActiveChildren, corev1.LocalObjectReference{Name: item.Name})
			}
			return it.kubeClient.Update(ctx, &node)
		})

		if updateError != nil {
			return reconcile.Result{}, updateError
		}

		var childrenNames []string
		for _, item := range childrenNodes {
			childrenNames = append(childrenNames, item.Name)
		}
		it.logger.Info("serial node's child created", "key", request.NamespacedName, "children", childrenNames)
	}

	return reconcile.Result{}, nil
}

func NewSerialNodeReconciler(kubeClient client.Client, eventRecorder record.EventRecorder, logger logr.Logger) *SerialNodeReconciler {
	return &SerialNodeReconciler{kubeClient: kubeClient, eventRecorder: eventRecorder, logger: logger}
}
