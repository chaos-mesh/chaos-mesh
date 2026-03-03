// Copyright Chaos Mesh Authors.
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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type StatusCheckReconciler struct {
	kubeClient    client.Client
	eventRecorder recorder.ChaosRecorder
	logger        logr.Logger
}

func NewStatusCheckReconciler(kubeClient client.Client, eventRecorder recorder.ChaosRecorder, logger logr.Logger) *StatusCheckReconciler {
	return &StatusCheckReconciler{kubeClient: kubeClient, eventRecorder: eventRecorder, logger: logger}
}

func (it *StatusCheckReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	startTime := time.Now()
	defer func() {
		it.logger.V(4).Info("finished syncing for status check node",
			"node", request.NamespacedName,
			"duration", time.Since(startTime),
		)
	}()

	node := v1alpha1.WorkflowNode{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	if node.Spec.Type != v1alpha1.TypeStatusCheck {
		return reconcile.Result{}, nil
	}

	it.logger.V(4).Info("resolve status check node", "node", request)
	if err := it.syncStatusCheck(ctx, request, node); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "sync status check")
	}

	updateError := retry.RetryOnConflict(retry.DefaultRetry, it.updateNodeStatus(ctx, request))

	return reconcile.Result{}, updateError
}

func (it *StatusCheckReconciler) syncStatusCheck(ctx context.Context, request reconcile.Request, node v1alpha1.WorkflowNode) error {
	statusChecks, err := it.fetchChildrenStatusCheck(ctx, node)
	if err != nil {
		return errors.Wrap(err, "fetch children status check")
	}

	if WorkflowNodeFinished(node.Status) {
		for _, item := range statusChecks {
			// best efforts deletion
			item := item
			err := it.kubeClient.Delete(ctx, &item)
			if client.IgnoreNotFound(err) != nil {
				it.logger.Error(err, "failed to delete StatusCheck for workflow status check node",
					"namespace", node.Namespace,
					"nodeName", node.Name,
					"statusCheckName", item.GetName(),
				)
				it.eventRecorder.Event(&node, recorder.StatusCheckDeletedFailed{Name: item.GetName()})
			} else {
				it.eventRecorder.Event(&node, recorder.StatusCheckDeleted{Name: item.GetName()})
			}
		}
		return nil
	}

	if len(statusChecks) == 0 {
		parentWorkflow, err := getParentWorkflow(ctx, it.kubeClient, node)
		if err != nil {
			return errors.WithStack(err)
		}
		spawnedStatusCheck, err := it.spawnStatusCheck(ctx, &node, parentWorkflow)
		if err != nil {
			it.eventRecorder.Event(&node, recorder.StatusCheckCreatedFailed{Name: spawnedStatusCheck.GetName()})
			return errors.Wrap(err, "spawn status check")
		}
		it.eventRecorder.Event(&node, recorder.StatusCheckCreated{Name: spawnedStatusCheck.GetName()})
	} else if len(statusChecks) > 1 {
		var statusCheckToRemove []string
		for _, item := range statusChecks[1:] {
			statusCheckToRemove = append(statusCheckToRemove, item.GetName())
		}
		it.logger.Info("removing duplicated StatusCheck",
			"node", request,
			"statusCheckToRemove", statusCheckToRemove)

		for _, item := range statusChecks[1:] {
			// best efforts deletion
			item := item
			err := it.kubeClient.Delete(ctx, &item)
			if client.IgnoreNotFound(err) != nil {
				it.logger.Error(err, "failed to delete StatusCheck for workflow status check node",
					"namespace", node.Namespace,
					"node", node.Name,
					"statusCheck", item.GetName(),
				)
			}
		}
	} else {
		it.logger.V(4).Info("do not need spawn or remove StatusCheck")
	}

	return nil
}

func (it *StatusCheckReconciler) updateNodeStatus(ctx context.Context, request reconcile.Request) func() error {
	return func() error {
		node := v1alpha1.WorkflowNode{}
		if err := it.kubeClient.Get(ctx, request.NamespacedName, &node); err != nil {
			return client.IgnoreNotFound(err)
		}

		statusChecks, err := it.fetchChildrenStatusCheck(ctx, node)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		if len(statusChecks) > 1 {
			it.logger.Info("the number of StatusCheck affected by status check node is more than 1",
				"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
				"statusCheck", statusChecks,
			)
		} else if len(statusChecks) == 0 {
			it.logger.Info("the number of StatusCheck affected by status check node is 0",
				"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
			)
			return nil
		}

		statusCheck := statusChecks[0]
		if statusCheck.IsCompleted() {
			SetCondition(&node.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAccomplished,
				Status: corev1.ConditionTrue,
				Reason: v1alpha1.StatusCheckCompleted,
			})
		} else {
			SetCondition(&node.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAccomplished,
				Status: corev1.ConditionFalse,
				Reason: "",
			})
		}

		if node.Spec.AbortWithStatusCheck && needToAbort(statusCheck) {
			SetCondition(&node.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAborted,
				Status: corev1.ConditionTrue,
				Reason: v1alpha1.StatusCheckNotExceedSuccessThreshold,
			})
		} else {
			SetCondition(&node.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAborted,
				Status: corev1.ConditionFalse,
				Reason: "",
			})
		}

		return client.IgnoreNotFound(it.kubeClient.Status().Update(ctx, &node))
	}
}

func (it *StatusCheckReconciler) fetchChildrenStatusCheck(ctx context.Context, node v1alpha1.WorkflowNode) ([]v1alpha1.StatusCheck, error) {
	controlledByThisNode, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			v1alpha1.LabelControlledBy: node.Name,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "build label selector")
	}

	var childStatusChecks v1alpha1.StatusCheckList
	if err = it.kubeClient.List(ctx, &childStatusChecks, &client.ListOptions{LabelSelector: controlledByThisNode}); err != nil {
		return nil, errors.Wrap(err, "list child status checks")
	}
	return childStatusChecks.Items, nil
}

func (it *StatusCheckReconciler) spawnStatusCheck(ctx context.Context, node *v1alpha1.WorkflowNode, workflow *v1alpha1.Workflow) (*v1alpha1.StatusCheck, error) {
	if node.Spec.StatusCheck == nil {
		return nil, errors.Errorf("node %s/%s does not contains spec of Target", node.Namespace, node.Name)
	}
	statusCheckSpec := node.Spec.StatusCheck.DeepCopy()
	statusCheck := v1alpha1.StatusCheck{
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
		Spec: *statusCheckSpec,
	}
	if err := it.kubeClient.Create(ctx, &statusCheck); err != nil {
		return nil, errors.Wrap(err, "create status check")
	}
	return &statusCheck, nil
}

func getParentWorkflow(ctx context.Context, kubeClient client.Client, node v1alpha1.WorkflowNode) (*v1alpha1.Workflow, error) {
	workflowName, ok := node.Labels[v1alpha1.LabelWorkflow]
	if !ok {
		return nil, errors.Errorf("node %s/%s does not contains label %s", node.Namespace, node.Name, v1alpha1.LabelWorkflow)
	}
	parentWorkflow := v1alpha1.Workflow{}
	if err := kubeClient.Get(ctx, types.NamespacedName{
		Namespace: node.Namespace,
		Name:      workflowName,
	}, &parentWorkflow); err != nil {
		return nil, errors.Wrap(err, "get parent workflow")
	}
	return &parentWorkflow, nil
}

func needToAbort(statusCheck v1alpha1.StatusCheck) bool {
	if !statusCheck.IsCompleted() {
		return false
	}
	for _, condition := range statusCheck.Status.Conditions {
		if condition.Type == v1alpha1.StatusCheckConditionSuccessThresholdExceed &&
			condition.Status != corev1.ConditionTrue {
			return true
		}
	}
	return false
}
