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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

// WorkflowEntryReconciler watches on Workflow, creates new Entry Node for created Workflow.
type WorkflowEntryReconciler struct {
	kubeClient    client.Client
	eventRecorder recorder.ChaosRecorder
	logger        logr.Logger
}

func NewWorkflowEntryReconciler(kubeClient client.Client, eventRecorder recorder.ChaosRecorder, logger logr.Logger) *WorkflowEntryReconciler {
	return &WorkflowEntryReconciler{kubeClient: kubeClient, eventRecorder: eventRecorder, logger: logger}
}

func (it *WorkflowEntryReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	startTime := time.Now()
	defer func() {
		it.logger.V(4).Info("Finished syncing for workflow",
			"node", request.NamespacedName,
			"duration", time.Since(startTime),
		)
	}()

	ctx := context.TODO()

	workflow := v1alpha1.Workflow{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &workflow)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	entryNodes, err := it.fetchEntryNode(ctx, workflow)
	if err != nil {
		it.logger.Error(err, "failed to list entry nodes of workflow",
			"workflow", request.NamespacedName)
		return reconcile.Result{}, err
	}

	if len(entryNodes) == 0 {
		func() {
			// Not scheduled yet, spawn the entry workflow node
			spawnedEntryNode, err := it.spawnEntryNode(ctx, workflow)
			if err != nil {
				it.eventRecorder.Event(&workflow, recorder.InvalidEntry{
					EntryTemplate: workflow.Spec.Entry,
				})
				it.logger.Error(err, "failed to spawn new entry node of workflow",
					"workflow", request.NamespacedName,
					"entry", workflow.Spec.Entry)
				// failed to spawn new entry, but will not break the reconcile, continue to sync status
				return
			}
			it.logger.Info(
				"entry node for workflow created",
				"workflow", request.NamespacedName,
				"entry node", fmt.Sprintf("%s/%s", spawnedEntryNode.Namespace, spawnedEntryNode.Name),
			)
			it.eventRecorder.Event(&workflow, recorder.EntryCreated{Entry: spawnedEntryNode.Name})
		}()
	}

	if len(entryNodes) > 1 {
		var nodeNames []string
		for _, node := range entryNodes {
			nodeNames = append(nodeNames, node.GetName())
		}
		it.logger.Info("there are more than 1 entry nodes of workflow, cleaning up except first one",
			"workflow", request.NamespacedName,
			"entry nodes", nodeNames,
		)
		for _, redundantEntryNode := range entryNodes[1:] {
			redundantEntryNode := redundantEntryNode
			// best effort deletion
			err := it.kubeClient.Delete(ctx, &redundantEntryNode)
			if err != nil {
				it.logger.Error(err,
					"failed to delete redundant entry node",
					"workflow", request.NamespacedName,
					"redundant entry node", fmt.Sprintf("%s/%s", redundantEntryNode.Namespace, redundantEntryNode.Name),
				)
			}
		}
	}

	// sync the status
	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		workflowNeedUpdate := v1alpha1.Workflow{}
		err := it.kubeClient.Get(ctx, request.NamespacedName, &workflowNeedUpdate)
		if err != nil {
			it.logger.Error(err,
				"failed to fetch the latest state of workflow",
				"workflow", request.NamespacedName,
			)
			return err
		}

		entryNodes, err := it.fetchEntryNode(ctx, workflowNeedUpdate)
		if err != nil {
			it.logger.Error(err,
				"failed to list entry nodes of workflow",
				"workflow", request.NamespacedName,
			)
			return err
		}

		if len(entryNodes) > 0 {
			if len(entryNodes) > 1 {
				var nodeNames []string
				for _, node := range entryNodes {
					nodeNames = append(nodeNames, node.GetName())
				}
				it.logger.Info("there are more than 1 entry nodes of workflow",
					"workflow", request.NamespacedName,
					"entry nodes", nodeNames,
				)
			}
			SetWorkflowCondition(&workflowNeedUpdate.Status, v1alpha1.WorkflowCondition{
				Type:   v1alpha1.WorkflowConditionScheduled,
				Status: corev1.ConditionTrue,
				Reason: "",
			})

			if WorkflowNodeFinished(entryNodes[0].Status) {
				SetWorkflowCondition(&workflowNeedUpdate.Status, v1alpha1.WorkflowCondition{
					Type:   v1alpha1.WorkflowConditionAccomplished,
					Status: corev1.ConditionTrue,
					Reason: "",
				})
				if workflowNeedUpdate.Status.EndTime == nil {
					now := metav1.NewTime(time.Now())
					workflowNeedUpdate.Status.EndTime = &now
				}
				it.eventRecorder.Event(&workflow, recorder.WorkflowAccomplished{})
			} else {
				SetWorkflowCondition(&workflowNeedUpdate.Status, v1alpha1.WorkflowCondition{
					Type:   v1alpha1.WorkflowConditionAccomplished,
					Status: corev1.ConditionFalse,
					Reason: "",
				})
				workflowNeedUpdate.Status.EndTime = nil
			}
		} else {
			SetWorkflowCondition(&workflowNeedUpdate.Status, v1alpha1.WorkflowCondition{
				Type:   v1alpha1.WorkflowConditionScheduled,
				Status: corev1.ConditionFalse,
				Reason: "",
			})
			SetWorkflowCondition(&workflowNeedUpdate.Status, v1alpha1.WorkflowCondition{
				Type:   v1alpha1.WorkflowConditionAccomplished,
				Status: corev1.ConditionFalse,
				Reason: "",
			})
			workflowNeedUpdate.Status.EndTime = nil
		}

		if workflowNeedUpdate.Status.StartTime == nil {
			tmp := metav1.NewTime(startTime)
			workflowNeedUpdate.Status.StartTime = &tmp
		}

		err = it.kubeClient.Status().Update(ctx, &workflowNeedUpdate)
		if err != nil {
			it.logger.Error(err, "failed to update workflowNeedUpdate status")
			return err
		}
		return nil
	})

	return reconcile.Result{}, client.IgnoreNotFound(updateError)
}

// fetchEntryNode will return the entry workflow node(s) of that workflow, return nil if not exists.
//
// The expected length of result is 1, but due to the reconcile and the inconsistent cache, there might be more than one
// entry nodes created, if should be reported to the upper logic.
func (it *WorkflowEntryReconciler) fetchEntryNode(ctx context.Context, workflow v1alpha1.Workflow) ([]v1alpha1.WorkflowNode, error) {
	entryNodesList := v1alpha1.WorkflowNodeList{}
	controlledByWorkflow, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			v1alpha1.LabelControlledBy: workflow.Name,
		},
	})
	if err != nil {
		it.logger.Error(err, "failed to build label selector with filtering entry workflow node controlled by current workflow",
			"workflow", fmt.Sprintf("%s/%s", workflow.Namespace, workflow.Name))
		return nil, err
	}

	err = it.kubeClient.List(ctx, &entryNodesList, &client.ListOptions{
		Namespace:     workflow.Namespace,
		LabelSelector: controlledByWorkflow,
	})
	if err != nil {
		it.logger.Error(err, "failed to list entry workflow node controlled by workflow",
			"workflow", fmt.Sprintf("%s/%s", workflow.Namespace, workflow.Name))
		return nil, err
	}

	sortedEntryNodes := SortByCreationTimestamp(entryNodesList.Items)
	sort.Sort(sortedEntryNodes)

	return sortedEntryNodes, nil
}

// spawnEntryNode will create **one** entry workflow node for current workflow
func (it *WorkflowEntryReconciler) spawnEntryNode(ctx context.Context, workflow v1alpha1.Workflow) (*v1alpha1.WorkflowNode, error) {
	// This workflow is just created, create entry node
	nodes, err := renderNodesByTemplates(&workflow, nil, workflow.Spec.Entry)
	if err != nil {
		it.logger.Error(err, "failed create entry node", "workflow", workflow.Name, "entry", workflow.Spec.Entry)
		return nil, err
	}

	if len(nodes) > 1 {
		it.logger.Info("the results of entry nodes are more than 1, will only pick the first one",
			"workflow", fmt.Sprintf("%s/%s", workflow.Namespace, workflow.Name),
			"nodes", nodes,
		)
	}

	entryNode := nodes[0]
	err = it.kubeClient.Create(ctx, entryNode)
	if err != nil {
		it.logger.Info("failed to create workflow nodes")
		return nil, err
	}
	it.logger.Info("entry workflow node created",
		"workflow", fmt.Sprintf("%s/%s", workflow.Namespace, workflow.Name),
		"entry node", entryNode.Name,
	)

	return entryNode, nil
}
