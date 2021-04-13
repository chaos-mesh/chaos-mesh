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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// WorkflowEntryReconciler watches on Workflow, creates new Entry Node for created Workflow.
type WorkflowEntryReconciler struct {
	kubeClient    client.Client
	eventRecorder record.EventRecorder
	logger        logr.Logger
}

func NewWorkflowEntryReconciler(kubeClient client.Client, eventRecorder record.EventRecorder, logger logr.Logger) *WorkflowEntryReconciler {
	return &WorkflowEntryReconciler{kubeClient: kubeClient, eventRecorder: eventRecorder, logger: logger}
}

func (it *WorkflowEntryReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()

	workflow := v1alpha1.Workflow{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &workflow)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if workflow.Status.EntryNode == nil {
		// This workflow is just created, create entry node
		nodes, err := renderNodesByTemplates(&workflow, nil, workflow.Spec.Entry)
		if err != nil {
			it.logger.Error(err, "failed create entry node", "workflow", workflow.Name, "entry", workflow.Spec.Entry)
			it.eventRecorder.Event(&workflow, corev1.EventTypeWarning, v1alpha1.InvalidEntry, "can not find workflow's entry template")
			return reconcile.Result{}, nil
		}

		if len(nodes) > 1 {
			it.logger.V(1).Info("the results of entry nodes are more than 1", "workflow", request.NamespacedName, "nodes", nodes)
		}

		entryNode := nodes[0]
		err = it.kubeClient.Create(ctx, entryNode)
		if err != nil {
			it.logger.V(1).Info("failed to create workflow nodes")
			return reconcile.Result{}, err
		}

		it.eventRecorder.Event(&workflow, corev1.EventTypeNormal, v1alpha1.EntryCreated, "Entry node created")

		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			workflowNeedUpdate := v1alpha1.Workflow{}
			err := it.kubeClient.Get(ctx, request.NamespacedName, &workflowNeedUpdate)
			if err != nil {
				return err
			}
			workflowNeedUpdate.Status.EntryNode = &entryNode.Name

			// TODO: add metav1.FinalizerDeleteDependents for workflowNeedUpdate's finalizer in webhook
			err = it.kubeClient.Status().Update(ctx, &workflowNeedUpdate)
			if err != nil {
				it.logger.Error(err, "failed to update workflowNeedUpdate status")
				return err
			}
			return nil
		})

		if updateError != nil {
			return reconcile.Result{}, updateError
		}

	}
	return reconcile.Result{}, nil
}
