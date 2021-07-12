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
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/task"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/task/collector"
)

type TaskReconciler struct {
	*ChildNodesFetcher
	kubeClient    client.Client
	restConfig    *rest.Config
	eventRecorder recorder.ChaosRecorder
	logger        logr.Logger
}

func NewTaskReconciler(kubeClient client.Client, restConfig *rest.Config, eventRecorder recorder.ChaosRecorder, logger logr.Logger) *TaskReconciler {
	return &TaskReconciler{
		ChildNodesFetcher: NewChildNodesFetcher(kubeClient, logger),
		kubeClient:        kubeClient,
		restConfig:        restConfig,
		eventRecorder:     eventRecorder,
		logger:            logger,
	}
}

func (it *TaskReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	startTime := time.Now()
	defer func() {
		it.logger.V(4).Info("Finished syncing for task node",
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

	// only resolve task nodes
	if node.Spec.Type != v1alpha1.TypeTask {
		return reconcile.Result{}, nil
	}

	it.logger.V(4).Info("resolve task node", "node", request)

	pods, err := it.FetchPodControlledByThisWorkflowNode(ctx, node)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(pods) == 0 {
		if workflowName, ok := node.Labels[v1alpha1.LabelWorkflow]; ok {
			parentWorkflow := v1alpha1.Workflow{}
			err := it.kubeClient.Get(ctx, types.NamespacedName{
				Namespace: node.Namespace,
				Name:      workflowName,
			}, &parentWorkflow)
			if err != nil {
				return reconcile.Result{}, err
			}
			spawnedPod, err := it.SpawnTaskPod(ctx, &node, &parentWorkflow)
			if err != nil {
				it.logger.Error(err, "failed to spawn pod for Task Node", "node", request)
				it.eventRecorder.Event(&node, recorder.TaskPodSpawnFailed{})
				return reconcile.Result{}, err
			}
			it.eventRecorder.Event(&node, recorder.TaskPodSpawned{PodName: spawnedPod.Name})
		} else {
			return reconcile.Result{}, errors.Errorf("node %s/%s does not contains label %s", node.Namespace, node.Name, v1alpha1.LabelWorkflow)
		}

	}

	if len(pods) > 1 {
		var podNames []string
		for _, pod := range pods {
			podNames = append(podNames, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
		}
		it.logger.Info("unexpected more than 1 pod created by task node, it will pick random one",
			"node", request,
			"pods", podNames,
			"picked", fmt.Sprintf("%s/%s", pods[0].Namespace, pods[0].Name),
		)
	}

	// update the status about conditional tasks
	if len(pods) > 0 && (pods[0].Status.Phase == corev1.PodFailed || pods[0].Status.Phase == corev1.PodSucceeded) {
		if !conditionalBranchesEvaluated(node) {
			it.eventRecorder.Event(&node, recorder.TaskPodPodCompleted{PodName: pods[0].Name})
			// task pod is terminated
			updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				nodeNeedUpdate := v1alpha1.WorkflowNode{}
				err := it.kubeClient.Get(ctx, request.NamespacedName, &nodeNeedUpdate)
				if err != nil {
					return err
				}

				if nodeNeedUpdate.Status.ConditionalBranchesStatus == nil {
					nodeNeedUpdate.Status.ConditionalBranchesStatus = &v1alpha1.ConditionalBranchesStatus{}
				}

				// TODO: update related condition
				defaultCollector := collector.DefaultCollector(it.kubeClient, it.restConfig, pods[0].Namespace, pods[0].Name, nodeNeedUpdate.Spec.Task.Container.Name)
				env, err := defaultCollector.CollectContext(ctx)
				if err != nil {
					it.logger.Error(err, "failed to fetch env from task",
						"task", fmt.Sprintf("%s/%s", nodeNeedUpdate.Namespace, nodeNeedUpdate.Name),
					)
					return err
				}
				if env != nil {
					jsonString, err := json.Marshal(env)
					if err != nil {
						it.logger.Error(err, "failed to convert env to json",
							"task", fmt.Sprintf("%s/%s", nodeNeedUpdate.Namespace, nodeNeedUpdate.Name),
							"env", env)
					} else {
						nodeNeedUpdate.Status.ConditionalBranchesStatus.Context = []string{string(jsonString)}
					}
				}

				evaluator := task.NewEvaluator(it.logger, it.kubeClient)
				evaluateConditionBranches, err := evaluator.EvaluateConditionBranches(nodeNeedUpdate.Spec.ConditionalBranches, env)
				if err != nil {
					it.logger.Error(err, "failed to evaluate expression",
						"task", fmt.Sprintf("%s/%s", nodeNeedUpdate.Namespace, nodeNeedUpdate.Name),
					)
					return err
				}

				nodeNeedUpdate.Status.ConditionalBranchesStatus.Branches = evaluateConditionBranches

				var selectedBranches []string
				for _, item := range evaluateConditionBranches {
					if item.EvaluationResult == corev1.ConditionTrue {
						selectedBranches = append(selectedBranches, item.Target)
					}
				}
				it.eventRecorder.Event(&nodeNeedUpdate, recorder.ConditionalBranchesSelected{SelectedBranches: selectedBranches})

				err = it.kubeClient.Status().Update(ctx, &nodeNeedUpdate)
				return err
			})
			if client.IgnoreNotFound(updateError) != nil {
				it.logger.Error(updateError, "failed to update the condition status of task",
					"task", request)
			}
		}
	} else {
		// task pod is still running or not exists
		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			nodeNeedUpdate := v1alpha1.WorkflowNode{}
			err := it.kubeClient.Get(ctx, request.NamespacedName, &nodeNeedUpdate)
			if err != nil {
				return err
			}
			// TODO: update related condition
			var branches []v1alpha1.ConditionalBranchStatus

			if nodeNeedUpdate.Status.ConditionalBranchesStatus == nil {
				nodeNeedUpdate.Status.ConditionalBranchesStatus = &v1alpha1.ConditionalBranchesStatus{}
			}

			for _, conditionalTask := range nodeNeedUpdate.Spec.ConditionalBranches {
				branch := v1alpha1.ConditionalBranchStatus{
					Target:           conditionalTask.Target,
					EvaluationResult: corev1.ConditionUnknown,
				}
				branches = append(branches, branch)
			}

			nodeNeedUpdate.Status.ConditionalBranchesStatus.Branches = branches

			err = it.kubeClient.Status().Update(ctx, &nodeNeedUpdate)
			return err
		})

		if client.IgnoreNotFound(updateError) != nil {
			it.logger.Error(updateError, "k failed to update the condition status of task",
				"task", request)
		}

	}

	// update the status about children nodes
	var evaluatedNode v1alpha1.WorkflowNode

	err = it.kubeClient.Get(ctx, request.NamespacedName, &evaluatedNode)
	if err != nil {
		return reconcile.Result{}, err
	}
	if conditionalBranchesEvaluated(evaluatedNode) {
		err = it.syncChildNodes(ctx, evaluatedNode)
		if err != nil {
			return reconcile.Result{}, err
		}

		// update the status of children workflow nodes
		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			nodeNeedUpdate := v1alpha1.WorkflowNode{}
			err := it.kubeClient.Get(ctx, request.NamespacedName, &nodeNeedUpdate)
			if err != nil {
				return err
			}
			var tasks []string
			for _, branch := range evaluatedNode.Status.ConditionalBranchesStatus.Branches {
				if branch.EvaluationResult == corev1.ConditionTrue {
					tasks = append(tasks, branch.Target)
				}
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

			if conditionalBranchesEvaluated(nodeNeedUpdate) && len(finishedChildren) == len(tasks) {
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

		return reconcile.Result{}, client.IgnoreNotFound(updateError)
	}

	return reconcile.Result{}, nil

}

func (it *TaskReconciler) syncChildNodes(ctx context.Context, evaluatedNode v1alpha1.WorkflowNode) error {

	var tasks []string
	for _, branch := range evaluatedNode.Status.ConditionalBranchesStatus.Branches {
		if branch.EvaluationResult == corev1.ConditionTrue {
			tasks = append(tasks, branch.Target)
		}
	}

	if len(tasks) == 0 {
		it.logger.V(4).Info("0 condition of branch in task node is True, Noop",
			"node", fmt.Sprintf("%s/%s", evaluatedNode.Namespace, evaluatedNode.Name),
		)
		return nil
	}

	activeChildNodes, finishedChildNodes, err := it.fetchChildNodes(ctx, evaluatedNode)
	if err != nil {
		return err
	}
	existsChildNodes := append(activeChildNodes, finishedChildNodes...)

	var taskNamesOfNodes []string
	for _, childNode := range existsChildNodes {
		taskNamesOfNodes = append(taskNamesOfNodes, getTaskNameFromGeneratedName(childNode.GetName()))
	}

	// TODO: check the specific of task and workflow nodes
	// the definition of tasks changed, remove all the existed nodes
	if len(setDifference(taskNamesOfNodes, tasks)) > 0 ||
		len(setDifference(tasks, taskNamesOfNodes)) > 0 {

		var nodesToCleanup []string
		for _, item := range existsChildNodes {
			nodesToCleanup = append(nodesToCleanup, item.Name)
		}
		it.eventRecorder.Event(&evaluatedNode, recorder.RerunBySpecChanged{CleanedChildrenNode: nodesToCleanup})

		for _, childNode := range existsChildNodes {
			// best effort deletion
			err := it.kubeClient.Delete(ctx, &childNode)
			if err != nil {
				it.logger.Error(err, "failed to delete outdated child node",
					"node", fmt.Sprintf("%s/%s", evaluatedNode.Namespace, evaluatedNode.Name),
					"child node", fmt.Sprintf("%s/%s", childNode.Namespace, childNode.Name),
				)
			}
		}
	} else {
		// exactly same, NOOP
		return nil
	}

	parentWorkflow := v1alpha1.Workflow{}
	err = it.kubeClient.Get(ctx, types.NamespacedName{
		Namespace: evaluatedNode.Namespace,
		Name:      evaluatedNode.Spec.WorkflowName,
	}, &parentWorkflow)
	if err != nil {
		it.logger.Error(err, "failed to fetch parent workflow",
			"node", fmt.Sprintf("%s/%s", evaluatedNode.Namespace, evaluatedNode.Name),
			"workflow name", evaluatedNode.Spec.WorkflowName)
		return err
	}

	childNodes, err := renderNodesByTemplates(&parentWorkflow, &evaluatedNode, tasks...)
	if err != nil {
		it.logger.Error(err, "failed to render children childNodes",
			"node", fmt.Sprintf("%s/%s", evaluatedNode.Namespace, evaluatedNode.Name))
		return err
	}

	// TODO: emit event
	var childrenNames []string
	for _, childNode := range childNodes {
		err := it.kubeClient.Create(ctx, childNode)
		if err != nil {
			it.logger.Error(err, "failed to create child node",
				"node", fmt.Sprintf("%s/%s", evaluatedNode.Namespace, evaluatedNode.Name),
				"child node", childNode)
			return err
		}
		childrenNames = append(childrenNames, childNode.Name)
	}
	it.eventRecorder.Event(&evaluatedNode, recorder.NodesCreated{ChildNodes: childrenNames})
	it.logger.Info("task node spawn new child node",
		"node", fmt.Sprintf("%s/%s", evaluatedNode.Namespace, evaluatedNode.Name),
		"child node", childrenNames)

	return nil
}

func (it *TaskReconciler) FetchPodControlledByThisWorkflowNode(ctx context.Context, node v1alpha1.WorkflowNode) ([]corev1.Pod, error) {
	controlledByThisNode, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			v1alpha1.LabelControlledBy: node.Name,
		},
	})

	if err != nil {
		it.logger.Error(err, "failed to build label selector with filtering children workflow node controlled by current node",
			"current node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return nil, err
	}

	var childPods corev1.PodList

	err = it.kubeClient.List(ctx, &childPods, &client.ListOptions{
		LabelSelector: controlledByThisNode,
	})
	if err != nil {
		return nil, err
	}
	return childPods.Items, nil
}

func (it *TaskReconciler) SpawnTaskPod(ctx context.Context, node *v1alpha1.WorkflowNode, workflow *v1alpha1.Workflow) (*corev1.Pod, error) {
	if node.Spec.Task == nil {
		return nil, errors.Errorf("node %s/%s does not contains spec of Target", node.Namespace, node.Name)
	}
	podSpec, err := task.SpawnPodForTask(*node.Spec.Task)
	if err != nil {
		return nil, err
	}
	taskPod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", node.Name),
			Namespace:    node.Namespace,
			Labels: map[string]string{
				v1alpha1.LabelControlledBy: node.Name,
				v1alpha1.LabelWorkflow:     workflow.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         ApiVersion,
					Kind:               KindWorkflowNode,
					Name:               node.Name,
					UID:                node.UID,
					Controller:         &isController,
					BlockOwnerDeletion: &blockOwnerDeletion,
				},
			},
			Finalizers: []string{metav1.FinalizerDeleteDependents},
		},
		Spec: podSpec,
	}
	err = it.kubeClient.Create(ctx, &taskPod)
	if err != nil {
		return nil, err
	}
	return &taskPod, nil
}

func conditionalBranchesEvaluated(node v1alpha1.WorkflowNode) bool {
	if node.Status.ConditionalBranchesStatus == nil {
		return false
	}
	if len(node.Spec.ConditionalBranches) != len(node.Status.ConditionalBranchesStatus.Branches) {
		return false
	}
	for _, branch := range node.Status.ConditionalBranchesStatus.Branches {
		if branch.EvaluationResult == corev1.ConditionUnknown {
			return false
		}
	}
	return true
}
