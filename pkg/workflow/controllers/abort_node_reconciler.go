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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type AbortNodeReconciler struct {
	*ChildNodesFetcher
	kubeClient    client.Client
	eventRecorder recorder.ChaosRecorder
	logger        logr.Logger
}

func NewAbortNodeReconciler(kubeClient client.Client, eventRecorder recorder.ChaosRecorder, logger logr.Logger) *AbortNodeReconciler {
	return &AbortNodeReconciler{
		ChildNodesFetcher: NewChildNodesFetcher(kubeClient, logger),
		kubeClient:        kubeClient,
		eventRecorder:     eventRecorder,
		logger:            logger,
	}
}

// Reconcile watches `WorkflowNodes`, if:
// 1. the abort condition is `False`, just return.
// 2. the abort condition is `True`, the node is not `TypeStatusCheck`, it will propagate abort condition to children nodes.
// 3. the abort condition is `True`, the node is `TypeStatusCheck`, it will add abort annotation to the parent workflow.
func (it *AbortNodeReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	node := v1alpha1.WorkflowNode{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if !ConditionEqualsTo(node.Status, v1alpha1.ConditionAborted, corev1.ConditionTrue) {
		return reconcile.Result{}, nil
	}

	if node.Spec.Type != v1alpha1.TypeStatusCheck {
		// if this node is aborted, try propagating to children node
		return reconcile.Result{}, it.propagateAbortToChildren(ctx, &node)
	}

	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := it.abortWorkflow(ctx, node); client.IgnoreNotFound(err) != nil {
			return errors.Wrapf(err, "abort parent workflow")
		}
		return nil
	})

	return reconcile.Result{}, client.IgnoreNotFound(updateError)
}

func (it *AbortNodeReconciler) propagateAbortToChildren(ctx context.Context, parent *v1alpha1.WorkflowNode) error {
	switch parent.Spec.Type {
	case v1alpha1.TypeSerial, v1alpha1.TypeParallel, v1alpha1.TypeTask:
		activeChildNodes, _, err := it.ChildNodesFetcher.fetchChildNodes(ctx, *parent)
		if err != nil {
			return errors.Wrap(err, "fetch children nodes")
		}
		for _, childNode := range activeChildNodes {
			childNode := childNode

			if WorkflowNodeFinished(childNode.Status) {
				it.logger.Info("child node already finished, skip for propagate abort", "node", fmt.Sprintf("%s/%s", childNode.Namespace, childNode.Name))
				continue
			}

			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				nodeNeedUpdate := v1alpha1.WorkflowNode{}
				err := it.kubeClient.Get(ctx, types.NamespacedName{
					Namespace: childNode.Namespace,
					Name:      childNode.Name,
				}, &nodeNeedUpdate)
				if err != nil {
					return errors.Wrap(err, "get child workflow node")
				}
				if ConditionEqualsTo(nodeNeedUpdate.Status, v1alpha1.ConditionAborted, corev1.ConditionTrue) {
					it.logger.Info("omit propagate abort to children, child already aborted",
						"node", fmt.Sprintf("%s/%s", nodeNeedUpdate.Namespace, nodeNeedUpdate.Name),
						"parent node", fmt.Sprintf("%s/%s", parent.Namespace, parent.Name),
					)
					return nil
				}
				SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
					Type:   v1alpha1.ConditionAborted,
					Status: corev1.ConditionTrue,
					Reason: v1alpha1.ParentNodeAborted,
				})
				it.eventRecorder.Event(&nodeNeedUpdate, recorder.ParentNodeAborted{ParentNodeName: parent.Name})
				return it.kubeClient.Status().Update(ctx, &nodeNeedUpdate)
			})
			if err != nil {
				return errors.Wrap(err, "update status of child workflow node")
			}
			it.logger.Info("propagate abort for child node",
				"child node", fmt.Sprintf("%s/%s", childNode.Namespace, childNode.Name),
				"parent node", fmt.Sprintf("%s/%s", parent.Namespace, parent.Name),
			)
		}
		return nil
	default:
		it.logger.V(4).Info("no need to propagate with this type of workflow node", "type", parent.Spec.Type)
		return nil
	}
}

func (it *AbortNodeReconciler) abortWorkflow(ctx context.Context, node v1alpha1.WorkflowNode) error {
	parentWorkflow, err := getParentWorkflow(ctx, it.kubeClient, node)
	if err != nil {
		return errors.WithStack(err)
	}
	if WorkflowAborted(*parentWorkflow) {
		return nil
	}

	it.logger.Info("add abort annotation to parent workflow",
		"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		"workflow", fmt.Sprintf("%s/%s", parentWorkflow.Namespace, parentWorkflow.Name))
	parentWorkflow.Annotations[v1alpha1.WorkflowAnnotationAbort] = "true"
	return it.kubeClient.Update(ctx, parentWorkflow)
}
