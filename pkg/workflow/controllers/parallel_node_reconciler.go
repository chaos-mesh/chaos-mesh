// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

// ParallelNodeReconciler watches on nodes which type is Parallel
type ParallelNodeReconciler struct {
	*ChildNodesFetcher
	kubeClient    client.Client
	eventRecorder recorder.ChaosRecorder
	logger        logr.Logger
}

func NewParallelNodeReconciler(kubeClient client.Client, eventRecorder recorder.ChaosRecorder, logger logr.Logger) *ParallelNodeReconciler {
	return &ParallelNodeReconciler{
		ChildNodesFetcher: NewChildNodesFetcher(kubeClient, logger),
		kubeClient:        kubeClient,
		eventRecorder:     eventRecorder,
		logger:            logger,
	}
}

// Reconcile schedules children in parallel and retries missing children after a partial creation.
func (it *ParallelNodeReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	startTime := time.Now()
	defer func() {
		it.logger.V(4).Info("Finished syncing for parallel node",
			"node", request.NamespacedName,
			"duration", time.Since(startTime),
		)
	}()

	node := v1alpha1.WorkflowNode{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// only resolve parallel nodes
	if node.Spec.Type != v1alpha1.TypeParallel || !node.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, nil
	}

	it.logger.V(4).Info("resolve parallel node", "node", request)

	// make effects, create/remove children nodes
	waiting, err := it.syncChildNodes(ctx, node)
	if err != nil {
		return reconcile.Result{}, err
	}
	if waiting {
		return reconcile.Result{RequeueAfter: time.Second}, nil
	}

	// update status
	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		err := it.kubeClient.Get(ctx, request.NamespacedName, &nodeNeedUpdate)
		if err != nil {
			return err
		}

		activeChildren, finishedChildren, err := it.fetchChildNodes(ctx, nodeNeedUpdate)
		if err != nil {
			return err
		}

		nodeNeedUpdate.Status.FinishedChildren = nil
		for _, finishedChild := range finishedChildren {
			nodeNeedUpdate.Status.FinishedChildren = append(nodeNeedUpdate.Status.FinishedChildren,
				corev1.LocalObjectReference{
					Name: finishedChild.Name,
				})
		}

		nodeNeedUpdate.Status.ActiveChildren = nil
		for _, activeChild := range activeChildren {
			nodeNeedUpdate.Status.ActiveChildren = append(nodeNeedUpdate.Status.ActiveChildren,
				corev1.LocalObjectReference{
					Name: activeChild.Name,
				})
		}

		missing, stale := missingChildTemplates(nodeNeedUpdate.Spec.Children, finishedChildren)
		if len(missing) == 0 && len(stale) == 0 && len(activeChildren) == 0 {
			if !WorkflowNodeFinished(nodeNeedUpdate.Status) {
				it.eventRecorder.Event(&nodeNeedUpdate, recorder.NodeAccomplished{})
			}
			SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAccomplished,
				Status: corev1.ConditionTrue,
				Reason: "",
			})
		} else {
			SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAccomplished,
				Status: corev1.ConditionFalse,
				Reason: "",
			})
		}

		return it.kubeClient.Status().Update(ctx, &nodeNeedUpdate)
	})

	if updateError != nil {
		it.logger.Error(updateError, "failed to update the status of node", "node", request)
		return reconcile.Result{}, updateError
	}

	return reconcile.Result{}, nil
}

func (it *ParallelNodeReconciler) syncChildNodes(ctx context.Context, node v1alpha1.WorkflowNode) (bool, error) {
	if WorkflowNodeFinished(node.Status) {
		return false, nil
	}
	active, finished, err := it.fetchChildNodes(ctx, node)
	if err != nil {
		return false, err
	}

	tasksToStartup, stale := missingChildTemplates(node.Spec.Children, append(finished, active...))
	if len(stale) > 0 {
		return true, deleteStaleChildren(ctx, it.kubeClient, it.eventRecorder, &node, stale)
	}

	if len(tasksToStartup) == 0 {
		it.logger.Info("no need to spawn new child node", "node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return false, nil
	}

	parentWorkflow := v1alpha1.Workflow{}
	err = it.kubeClient.Get(ctx, types.NamespacedName{
		Namespace: node.Namespace,
		Name:      node.Spec.WorkflowName,
	}, &parentWorkflow)
	if err != nil {
		it.logger.Error(err, "failed to fetch parent workflow",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
			"workflow name", node.Spec.WorkflowName)
		return false, err
	}

	childNodes, err := renderNodesByTemplates(&parentWorkflow, &node, tasksToStartup...)
	if err != nil {
		it.logger.Error(err, "failed to render children childNodes",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return false, err
	}

	var childrenNames []string
	for _, childNode := range childNodes {
		err := it.kubeClient.Create(ctx, childNode)
		if err != nil {
			it.logger.Error(err, "failed to create child node",
				"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
				"child node", childNode)
			return false, err
		}
		childrenNames = append(childrenNames, childNode.Name)
	}
	it.eventRecorder.Event(&node, recorder.NodesCreated{ChildNodes: childrenNames})
	it.logger.Info("parallel node spawn new child node",
		"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		"child node", childrenNames)

	return false, nil
}
