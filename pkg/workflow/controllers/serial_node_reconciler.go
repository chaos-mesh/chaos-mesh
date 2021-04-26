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
	"sort"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// SerialNodeReconciler watches on nodes which type is Serial
type SerialNodeReconciler struct {
	kubeClient    client.Client
	eventRecorder record.EventRecorder
	logger        logr.Logger
}

func NewSerialNodeReconciler(kubeClient client.Client, eventRecorder record.EventRecorder, logger logr.Logger) *SerialNodeReconciler {
	return &SerialNodeReconciler{kubeClient: kubeClient, eventRecorder: eventRecorder, logger: logger}
}

func (it *SerialNodeReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	startTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing for serial node %q (%v)", request.NamespacedName, time.Since(startTime))
	}()

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

	it.logger.V(4).Info("resolve serial node", "node", request)

	err = it.syncChildrenNodes(ctx, node)
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

		// un-synced expected children
		if nodeNeedUpdate.Status.ExpectedChildrenNum == nil || *nodeNeedUpdate.Status.ExpectedChildrenNum != len(nodeNeedUpdate.Spec.Tasks) {
			expected := len(nodeNeedUpdate.Spec.Tasks)
			nodeNeedUpdate.Status.ExpectedChildrenNum = &expected
		}

		activeChildren, finishedChildren, err := it.fetchChildrenNodes(ctx, nodeNeedUpdate)
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

		if len(activeChildren) > 1 {
			it.logger.Info("warning: serial node has more than 1 active children", "namespace", nodeNeedUpdate.Namespace, "name", nodeNeedUpdate.Name, "children", nodeNeedUpdate.Status.ActiveChildren)
		}

		if nodeNeedUpdate.Status.ExpectedChildrenNum != nil && len(finishedChildren) == *nodeNeedUpdate.Status.ExpectedChildrenNum {
			SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAccomplished,
				Status: corev1.ConditionTrue,
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

func (it *SerialNodeReconciler) syncChildrenNodes(ctx context.Context, node v1alpha1.WorkflowNode) error {

	// empty serial node
	if len(node.Spec.Tasks) == 0 {
		it.logger.V(4).Info("empty serial node, NOOP",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		)
		return nil
	}

	activeChildrenNodes, finishedChildrenNodes, err := it.fetchChildrenNodes(ctx, node)
	if err != nil {
		return err
	}
	var taskToStartup string
	if len(activeChildrenNodes) == 0 {
		for index, task := range node.Spec.Tasks {
			if index < len(finishedChildrenNodes) {
				// TODO: if the definition/spec of task changed, we should also respawn the node
				// child node start with task name
				if strings.Index(finishedChildrenNodes[index].Name, task) != 0 {
					// TODO: emit event
					taskToStartup = task

					// TODO: nodes to delete should be all other unrecognized children nodes, include not contained in finishedChildrenNodes
					// delete that related nodes with best-effort pattern
					nodesToDelete := finishedChildrenNodes[index:]
					for _, refToDelete := range nodesToDelete {
						nodeToDelete := v1alpha1.WorkflowNode{}
						err := it.kubeClient.Get(ctx, types.NamespacedName{
							Namespace: node.Namespace,
							Name:      refToDelete.Name,
						}, &nodeToDelete)
						if client.IgnoreNotFound(err) != nil {
							it.logger.Error(err, "failed to fetch outdated child node",
								"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
								"child node", fmt.Sprintf("%s/%s", node.Namespace, nodeToDelete.Name))
						}
						err = it.kubeClient.Delete(ctx, &nodeToDelete)
						if client.IgnoreNotFound(err) != nil {
							it.logger.Error(err, "failed to fetch outdated child node",
								"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
								"child node", fmt.Sprintf("%s/%s", node.Namespace, nodeToDelete.Name))
						}
					}
					break
				}
			} else {
				// spawn child node
				taskToStartup = task
				break
			}
		}
	} else {
		it.logger.V(4).Info("serial node has active child/children, skip scheduling",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
			"active children", activeChildrenNodes)
	}

	if len(taskToStartup) == 0 {
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
	childrenNodes, err := renderNodesByTemplates(&parentWorkflow, &node, taskToStartup)
	if err != nil {
		it.logger.Error(err, "failed to render children childrenNodes",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return err
	}

	// TODO: emit event
	var childrenNames []string
	for _, childNode := range childrenNodes {
		err := it.kubeClient.Create(ctx, childNode)
		if err != nil {
			it.logger.Error(err, "failed to create child node",
				"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
				"child node", childNode)
			return err
		}
		childrenNames = append(childrenNames, childNode.Name)
	}
	it.logger.Info("serial node spawn new child node",
		"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		"child node", childrenNames)

	return nil
}

func (it *SerialNodeReconciler) fetchChildrenNodes(ctx context.Context, node v1alpha1.WorkflowNode) (activeChildrenNodes []v1alpha1.WorkflowNode, finishedChildrenNodes []v1alpha1.WorkflowNode, err error) {
	childrenNodes := v1alpha1.WorkflowNodeList{}
	controlledByThisNode, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			LabelControlledBy: node.Name,
		},
	})

	if err != nil {
		it.logger.Error(err, "failed to build label selector with filtering children workflow node controlled by current node",
			"current node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return nil, nil, err
	}

	// TODO: sort with CreationTimestamp
	err = it.kubeClient.List(ctx, &childrenNodes, &client.ListOptions{
		LabelSelector: controlledByThisNode,
	})

	if err != nil {
		it.logger.Error(err, "failed to list children workflow node controlled by current node",
			"current node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return nil, nil, err
	}

	sortedChildrenNodes := SortByCreationTimestamp(childrenNodes.Items)
	sort.Sort(sortedChildrenNodes)

	it.logger.V(4).Info("list children node", "current node",
		"current node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		len(sortedChildrenNodes), "children", sortedChildrenNodes)

	var activeChildren []v1alpha1.WorkflowNode
	var finishedChildren []v1alpha1.WorkflowNode

	for _, item := range sortedChildrenNodes {
		childNode := item
		if WorkflowNodeFinished(childNode.Status) {
			finishedChildren = append(finishedChildren, childNode)
		} else {
			activeChildren = append(activeChildren, childNode)
		}
	}
	return activeChildren, finishedChildren, nil
}

type SortByCreationTimestamp []v1alpha1.WorkflowNode

func (it SortByCreationTimestamp) Len() int {
	return len(it)
}

func (it SortByCreationTimestamp) Less(i, j int) bool {
	return it[j].GetCreationTimestamp().After(it[i].GetCreationTimestamp().Time)
}

func (it SortByCreationTimestamp) Swap(i, j int) {
	it[i], it[j] = it[j], it[i]
}
