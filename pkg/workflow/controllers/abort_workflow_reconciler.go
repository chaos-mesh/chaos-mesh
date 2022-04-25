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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type AbortWorkflowReconciler struct {
	*ChildNodesFetcher
	kubeClient    client.Client
	eventRecorder recorder.ChaosRecorder
	logger        logr.Logger
}

func NewAbortWorkflowReconciler(kubeClient client.Client, eventRecorder recorder.ChaosRecorder, logger logr.Logger) *AbortWorkflowReconciler {
	return &AbortWorkflowReconciler{
		ChildNodesFetcher: NewChildNodesFetcher(kubeClient, logger),
		kubeClient:        kubeClient,
		eventRecorder:     eventRecorder,
		logger:            logger,
	}
}

// Reconcile watches `Workflows`, if the workflow has the abort annotation,
// it will set the abort condition of the `entry node` to `True`.
func (it *AbortWorkflowReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	workflow := v1alpha1.Workflow{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &workflow)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		workflowNeedUpdate := v1alpha1.Workflow{}
		err := it.kubeClient.Get(ctx, request.NamespacedName, &workflowNeedUpdate)
		if err != nil {
			return errors.Wrapf(err, "get workflow")
		}

		entryNodes, err := fetchEntryNode(ctx, it.kubeClient, workflowNeedUpdate)
		if err != nil {
			return errors.Wrapf(err, "fetch entry nodes of workflow")
		}

		if len(entryNodes) == 0 {
			it.logger.Info("omit set abort condition, workflow has no entry node", "key", request.NamespacedName)
			return nil
		}
		if len(entryNodes) > 1 {
			it.logger.Info("there are more than 1 entry nodes of workflow", "key", request.NamespacedName)
		}

		entryNode := entryNodes[0]
		if WorkflowAborted(workflowNeedUpdate) {
			if !ConditionEqualsTo(entryNode.Status, v1alpha1.ConditionAborted, corev1.ConditionTrue) {
				it.eventRecorder.Event(&entryNode, recorder.WorkflowAborted{WorkflowName: workflow.Name})
			}
			SetCondition(&entryNode.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAborted,
				Status: corev1.ConditionTrue,
				Reason: v1alpha1.WorkflowAborted,
			})
		} else {
			SetCondition(&entryNode.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAborted,
				Status: corev1.ConditionFalse,
				Reason: "",
			})
		}

		return client.IgnoreNotFound(it.kubeClient.Status().Update(ctx, &entryNode))
	})

	return reconcile.Result{}, client.IgnoreNotFound(updateError)
}
