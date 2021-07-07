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

// Reconcile is extremely like the one in SerialNodeReconciler, only allows the parallel schedule, and respawn **all** the children tasks during retry
func (it *ParallelNodeReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	startTime := time.Now()
	defer func() {
		it.logger.V(4).Info("Finished syncing for parallel node",
			"node", request.NamespacedName,
			"duration", time.Since(startTime),
		)
	}()

	ctx := context.TODO()

	node := v1alpha1.WorkflowNode{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// only resolve parallel nodes
	if node.Spec.Type != v1alpha1.TypeParallel {
		return reconcile.Result{}, nil
	}

	it.logger.V(4).Info("resolve parallel node", "node", request)

	// make effects, create/remove children nodes
	err = it.syncChildNodes(ctx, node)
	if err != nil {
		return reconcile.Result{}, err
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

		// TODO: also check the consistent between spec in task and the spec in child node
		if len(finishedChildren) == len(nodeNeedUpdate.Spec.Children) {
			SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAccomplished,
				Status: corev1.ConditionTrue,
				Reason: "",
			})
			it.eventRecorder.Event(&nodeNeedUpdate, recorder.NodeAccomplished{})
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
		it.logger.Error(err, "failed to update the status of node", "node", request)
		return reconcile.Result{}, updateError
	}

	return reconcile.Result{}, nil
}

func (it *ParallelNodeReconciler) syncChildNodes(ctx context.Context, node v1alpha1.WorkflowNode) error {

	// empty parallel node
	if len(node.Spec.Children) == 0 {
		it.logger.V(4).Info("empty parallel node, NOOP",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		)
		return nil
	}

	if WorkflowNodeFinished(node.Status) {
		return nil
	}

	activeChildNodes, finishedChildNodes, err := it.fetchChildNodes(ctx, node)
	if err != nil {
		return err
	}
	existsChildNodes := append(activeChildNodes, finishedChildNodes...)

	var taskNamesOfNodes []string
	for _, childNode := range existsChildNodes {
		taskNamesOfNodes = append(taskNamesOfNodes, getTaskNameFromGeneratedName(childNode.GetName()))
	}

	var tasksToStartup []string

	// TODO: check the specific of task and workflow nodes
	// the definition of Spec.Children changed, remove all the existed nodes
	if len(setDifference(taskNamesOfNodes, node.Spec.Children)) > 0 ||
		len(setDifference(node.Spec.Children, taskNamesOfNodes)) > 0 {
		tasksToStartup = node.Spec.Children

		var nodesToCleanup []string
		for _, item := range existsChildNodes {
			nodesToCleanup = append(nodesToCleanup, item.Name)
		}
		it.eventRecorder.Event(&node, recorder.RerunBySpecChanged{CleanedChildrenNode: nodesToCleanup})

		for _, childNode := range existsChildNodes {
			// best effort deletion
			err := it.kubeClient.Delete(ctx, &childNode)
			if err != nil {
				it.logger.Error(err, "failed to delete outdated child node",
					"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
					"child node", fmt.Sprintf("%s/%s", childNode.Namespace, childNode.Name),
				)
			}
		}

	}

	if len(tasksToStartup) == 0 {
		it.logger.Info("no need to spawn new child node", "node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return nil
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
		return err
	}

	childNodes, err := renderNodesByTemplates(&parentWorkflow, &node, tasksToStartup...)
	if err != nil {
		it.logger.Error(err, "failed to render children childNodes",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return err
	}

	var childrenNames []string
	for _, childNode := range childNodes {
		err := it.kubeClient.Create(ctx, childNode)
		if err != nil {
			it.logger.Error(err, "failed to create child node",
				"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
				"child node", childNode)
			return err
		}
		childrenNames = append(childrenNames, childNode.Name)
	}
	it.eventRecorder.Event(&node, recorder.NodesCreated{ChildNodes: childrenNames})
	it.logger.Info("parallel node spawn new child node",
		"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		"child node", childrenNames)

	return nil
}
