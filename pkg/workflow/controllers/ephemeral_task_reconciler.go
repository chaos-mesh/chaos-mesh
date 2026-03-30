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
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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

const (
	ephemeralTaskPodCreatedAnnotation      = "workflow.chaos-mesh.org/ephemeral-task-pod-created"
	ephemeralTaskResultCollectedAnnotation = "workflow.chaos-mesh.org/ephemeral-task-result-collected"
)

type EphemeralTaskReconciler struct {
	*TaskReconciler
}

func NewEphemeralTaskReconciler(kubeClient client.Client, restConfig *rest.Config, eventRecorder recorder.ChaosRecorder, logger logr.Logger) *EphemeralTaskReconciler {
	return &EphemeralTaskReconciler{
		TaskReconciler: NewTaskReconciler(kubeClient, restConfig, eventRecorder, logger),
	}
}

func (it *EphemeralTaskReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	startTime := time.Now()
	defer func() {
		it.logger.V(4).Info("Finished syncing for ephemeral task node",
			"node", request.NamespacedName,
			"duration", time.Since(startTime),
		)
	}()

	node := v1alpha1.WorkflowNode{}
	if err := it.kubeClient.Get(ctx, request.NamespacedName, &node); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if node.Spec.Type != v1alpha1.TypeEphemeralTask {
		return reconcile.Result{}, nil
	}

	if err := it.ensurePendingBranchesStatus(ctx, request.NamespacedName, node); err != nil {
		return reconcile.Result{}, err
	}

	evaluated, err := it.ephemeralConditionalBranchesEvaluated(node)
	if err != nil {
		return reconcile.Result{}, err
	}

	pods, err := it.FetchPodControlledByThisWorkflowNode(ctx, node)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(pods) == 0 {
		switch {
		case evaluated:
		case !ephemeralTaskPodCreated(node):
			if err := it.spawnEphemeralTaskPod(ctx, request.NamespacedName, &node); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		default:
			if err := it.persistSyntheticFailure(ctx, request.NamespacedName); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		if len(pods) > 1 {
			var podNames []string
			for _, pod := range pods {
				podNames = append(podNames, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
			}
			it.logger.Info("unexpected more than 1 pod created by ephemeral task node, it will pick random one",
				"node", request,
				"pods", podNames,
				"picked", fmt.Sprintf("%s/%s", pods[0].Namespace, pods[0].Name),
			)
		}

		pod := pods[0]
		if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded {
			if !evaluated {
				if err := it.persistPodResult(ctx, request.NamespacedName, pod); err != nil {
					return reconcile.Result{}, err
				}
			}
			if err := client.IgnoreNotFound(it.kubeClient.Delete(ctx, &pod)); err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	evaluatedNode := v1alpha1.WorkflowNode{}
	if err := it.kubeClient.Get(ctx, request.NamespacedName, &evaluatedNode); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	evaluated, err = it.ephemeralConditionalBranchesEvaluated(evaluatedNode)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !evaluated {
		return reconcile.Result{}, nil
	}

	if err := it.syncChildNodes(ctx, evaluatedNode); err != nil {
		return reconcile.Result{}, err
	}

	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		if err := it.kubeClient.Get(ctx, request.NamespacedName, &nodeNeedUpdate); err != nil {
			return err
		}

		var tasks []string
		for _, branch := range nodeNeedUpdate.Status.ConditionalBranchesStatus.Branches {
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
			nodeNeedUpdate.Status.FinishedChildren = append(nodeNeedUpdate.Status.FinishedChildren, corev1.LocalObjectReference{
				Name: finishedChild.Name,
			})
		}

		nodeNeedUpdate.Status.ActiveChildren = nil
		for _, activeChild := range activeChildren {
			nodeNeedUpdate.Status.ActiveChildren = append(nodeNeedUpdate.Status.ActiveChildren, corev1.LocalObjectReference{
				Name: activeChild.Name,
			})
		}

		evaluated, err := it.ephemeralConditionalBranchesEvaluated(nodeNeedUpdate)
		if err != nil {
			return err
		}
		if evaluated && len(finishedChildren) == len(tasks) {
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

	return reconcile.Result{}, client.IgnoreNotFound(updateError)
}

func (it *EphemeralTaskReconciler) ensurePendingBranchesStatus(ctx context.Context, namespacedName types.NamespacedName, node v1alpha1.WorkflowNode) error {
	if ephemeralTaskResultCollected(node) {
		return nil
	}

	needsUpdate := node.Status.ConditionalBranchesStatus == nil || len(node.Status.ConditionalBranchesStatus.Branches) != len(node.Spec.ConditionalBranches)
	if !needsUpdate {
		allUnknown := true
		for _, branch := range node.Status.ConditionalBranchesStatus.Branches {
			if branch.EvaluationResult != corev1.ConditionUnknown {
				allUnknown = false
				break
			}
		}
		if allUnknown {
			return nil
		}
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		if err := it.kubeClient.Get(ctx, namespacedName, &nodeNeedUpdate); err != nil {
			return err
		}

		if ephemeralTaskResultCollected(nodeNeedUpdate) {
			return nil
		}

		branches := make([]v1alpha1.ConditionalBranchStatus, 0, len(nodeNeedUpdate.Spec.ConditionalBranches))
		for _, conditionalTask := range nodeNeedUpdate.Spec.ConditionalBranches {
			branches = append(branches, v1alpha1.ConditionalBranchStatus{
				Target:           conditionalTask.Target,
				EvaluationResult: corev1.ConditionUnknown,
			})
		}

		if nodeNeedUpdate.Status.ConditionalBranchesStatus == nil {
			nodeNeedUpdate.Status.ConditionalBranchesStatus = &v1alpha1.ConditionalBranchesStatus{}
		}
		nodeNeedUpdate.Status.ConditionalBranchesStatus.Branches = branches
		if len(nodeNeedUpdate.Status.ConditionalBranchesStatus.Context) == 0 {
			nodeNeedUpdate.Status.ConditionalBranchesStatus.Context = nil
		}

		return it.kubeClient.Status().Update(ctx, &nodeNeedUpdate)
	})
}

func (it *EphemeralTaskReconciler) spawnEphemeralTaskPod(ctx context.Context, namespacedName types.NamespacedName, node *v1alpha1.WorkflowNode) error {
	workflowName, ok := node.Labels[v1alpha1.LabelWorkflow]
	if !ok {
		return errors.Errorf("node %s/%s does not contains label %s", node.Namespace, node.Name, v1alpha1.LabelWorkflow)
	}

	parentWorkflow := v1alpha1.Workflow{}
	if err := it.kubeClient.Get(ctx, types.NamespacedName{
		Namespace: node.Namespace,
		Name:      workflowName,
	}, &parentWorkflow); err != nil {
		return err
	}

	spawnedPod, err := it.SpawnTaskPod(ctx, node, &parentWorkflow)
	if err != nil {
		it.logger.Error(err, "failed to spawn pod for ephemeral task node", "node", namespacedName)
		it.eventRecorder.Event(node, recorder.TaskPodSpawnFailed{})
		return err
	}

	it.eventRecorder.Event(node, recorder.TaskPodSpawned{PodName: spawnedPod.Name})

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		if err := it.kubeClient.Get(ctx, namespacedName, &nodeNeedUpdate); err != nil {
			return err
		}
		ensureWorkflowNodeAnnotations(&nodeNeedUpdate)
		nodeNeedUpdate.Annotations[ephemeralTaskPodCreatedAnnotation] = "true"
		return it.kubeClient.Update(ctx, &nodeNeedUpdate)
	})
}

func (it *EphemeralTaskReconciler) persistPodResult(ctx context.Context, namespacedName types.NamespacedName, pod corev1.Pod) error {
	node := v1alpha1.WorkflowNode{}
	if err := it.kubeClient.Get(ctx, namespacedName, &node); err != nil {
		return err
	}

	defaultCollector := collector.DefaultCollector(it.kubeClient, it.restConfig, pod.Namespace, pod.Name, taskContainerNameForCollection(node, pod))
	env, err := defaultCollector.CollectContext(ctx)
	if err != nil {
		it.logger.Error(err, "failed to fetch env from ephemeral task",
			"task", namespacedName.String(),
			"pod", fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
		)
		return err
	}

	it.eventRecorder.Event(&node, recorder.TaskPodPodCompleted{PodName: pod.Name})
	return it.persistEvaluatedResult(ctx, namespacedName, env)
}

func (it *EphemeralTaskReconciler) persistSyntheticFailure(ctx context.Context, namespacedName types.NamespacedName) error {
	it.logger.Info("ephemeral task pod disappeared before the result was collected, mark as failed without respawn",
		"node", namespacedName,
	)
	return it.persistEvaluatedResult(ctx, namespacedName, map[string]interface{}{
		"exitCode": -1,
		"stdout":   "",
	})
}

func (it *EphemeralTaskReconciler) persistEvaluatedResult(ctx context.Context, namespacedName types.NamespacedName, env map[string]interface{}) error {
	node := v1alpha1.WorkflowNode{}
	if err := it.kubeClient.Get(ctx, namespacedName, &node); err != nil {
		return err
	}
	if ephemeralTaskResultCollected(node) {
		return nil
	}

	evaluator := task.NewEvaluator(it.logger, it.kubeClient)
	evaluatedBranches, err := evaluator.EvaluateConditionBranches(node.Spec.ConditionalBranches, env)
	if err != nil {
		it.logger.Error(err, "failed to evaluate expression",
			"task", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		)
		return err
	}

	contextValue := ""
	if env != nil {
		jsonString, err := json.Marshal(env)
		if err != nil {
			it.logger.Error(err, "failed to convert env to json",
				"task", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
				"env", env)
			return err
		}
		contextValue = string(jsonString)
	}

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		if err := it.kubeClient.Get(ctx, namespacedName, &nodeNeedUpdate); err != nil {
			return err
		}
		if nodeNeedUpdate.Status.ConditionalBranchesStatus == nil {
			nodeNeedUpdate.Status.ConditionalBranchesStatus = &v1alpha1.ConditionalBranchesStatus{}
		}
		if len(contextValue) > 0 {
			nodeNeedUpdate.Status.ConditionalBranchesStatus.Context = []string{contextValue}
		} else {
			nodeNeedUpdate.Status.ConditionalBranchesStatus.Context = nil
		}
		nodeNeedUpdate.Status.ConditionalBranchesStatus.Branches = evaluatedBranches

		return it.kubeClient.Status().Update(ctx, &nodeNeedUpdate)
	}); err != nil {
		return err
	}

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		if err := it.kubeClient.Get(ctx, namespacedName, &nodeNeedUpdate); err != nil {
			return err
		}
		ensureWorkflowNodeAnnotations(&nodeNeedUpdate)
		nodeNeedUpdate.Annotations[ephemeralTaskResultCollectedAnnotation] = "true"
		return it.kubeClient.Update(ctx, &nodeNeedUpdate)
	}); err != nil {
		return err
	}

	var selectedBranches []string
	for _, item := range evaluatedBranches {
		if item.EvaluationResult == corev1.ConditionTrue {
			selectedBranches = append(selectedBranches, item.Target)
		}
	}
	nodeNeedUpdate := v1alpha1.WorkflowNode{}
	if err := it.kubeClient.Get(ctx, namespacedName, &nodeNeedUpdate); err != nil {
		return err
	}
	it.eventRecorder.Event(&nodeNeedUpdate, recorder.ConditionalBranchesSelected{SelectedBranches: selectedBranches})
	return nil
}

func (it *EphemeralTaskReconciler) ephemeralConditionalBranchesEvaluated(node v1alpha1.WorkflowNode) (bool, error) {
	if node.Status.ConditionalBranchesStatus == nil {
		return false, nil
	}

	if len(node.Spec.ConditionalBranches) != len(node.Status.ConditionalBranchesStatus.Branches) {
		return false, nil
	}

	for _, branch := range node.Status.ConditionalBranchesStatus.Branches {
		if branch.EvaluationResult == corev1.ConditionUnknown {
			return false, nil
		}
	}

	return true, nil
}

func ensureWorkflowNodeAnnotations(node *v1alpha1.WorkflowNode) {
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
}

func ephemeralTaskPodCreated(node v1alpha1.WorkflowNode) bool {
	return node.Annotations[ephemeralTaskPodCreatedAnnotation] == "true"
}

func ephemeralTaskResultCollected(node v1alpha1.WorkflowNode) bool {
	return node.Annotations[ephemeralTaskResultCollectedAnnotation] == "true" || ephemeralTaskStatusReady(node)
}

func taskContainerNameForCollection(node v1alpha1.WorkflowNode, pod corev1.Pod) string {
	if node.Spec.Task != nil && node.Spec.Task.Container != nil && len(node.Spec.Task.Container.Name) > 0 {
		return node.Spec.Task.Container.Name
	}

	if len(pod.Spec.Containers) == 0 {
		return ""
	}
	return pod.Spec.Containers[0].Name
}

func ephemeralTaskStatusReady(node v1alpha1.WorkflowNode) bool {
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
