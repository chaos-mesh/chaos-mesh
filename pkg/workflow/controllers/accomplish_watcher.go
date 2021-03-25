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
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type AccomplishWatcher struct {
	kubeClient    client.Client
	eventRecorder record.EventRecorder
	logger        logr.Logger
}

func NewAccomplishWatcher(kubeClient client.Client, eventRecorder record.EventRecorder, logger logr.Logger) *AccomplishWatcher {
	return &AccomplishWatcher{kubeClient: kubeClient, eventRecorder: eventRecorder, logger: logger}
}

func (it *AccomplishWatcher) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()

	node := v1alpha1.WorkflowNode{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if ConditionEqualsTo(node.Status, v1alpha1.ConditionAccomplished, corev1.ConditionTrue) ||
		ConditionEqualsTo(node.Status, v1alpha1.ConditionDeadlineExceed, corev1.ConditionTrue) {
		owners := node.OwnerReferences
		// NOOP
		if len(owners) == 0 {
			it.logger.V(1).Info("dangling node has no owner", "node", request.NamespacedName)
			return reconcile.Result{}, nil
		}

		// unexpected situation
		if len(owners) > 1 {
			it.logger.V(1).Info("node has more than one owner, it will not take any operates", "node", request.NamespacedName, "owners", owners)
			return reconcile.Result{}, nil
		}

		owner := owners[0]
		it.logger.V(4).Info("fetch node's owner", "node", request.NamespacedName, "owner", owner)
		if owner.Kind == GetKindOf(&v1alpha1.WorkflowNode{}) {
			parentNode := v1alpha1.WorkflowNode{}

			err := it.kubeClient.Get(ctx, types.NamespacedName{
				Namespace: request.Namespace,
				Name:      owner.Name,
			}, &parentNode)
			if err != nil {
				return reconcile.Result{}, err
			}
			if parentNode.Spec.Type == v1alpha1.TypeSerial {
				err = it.updateParentSerialNode(ctx, node, parentNode)
				return reconcile.Result{}, err
			}
			it.logger.Info("unsupported owner node type", "node type", parentNode.Spec.Type)
		} else if owner.Kind == GetKindOf(&v1alpha1.Workflow{}) {
			// TODO: update the status of workflow
			it.logger.Info("unsupported update for workflow", "kind", owner.Kind)
		} else {
			it.logger.Info("unsupported owner type", "kind", owner.Kind)
		}
	}
	return reconcile.Result{}, nil
}

func (it *AccomplishWatcher) updateParentSerialNode(ctx context.Context, childNode, parentNode v1alpha1.WorkflowNode) error {

	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		err := it.kubeClient.Get(ctx, types.NamespacedName{
			Namespace: parentNode.Namespace,
			Name:      parentNode.Name,
		}, &nodeNeedUpdate)

		if err != nil {
			return client.IgnoreNotFound(err)
		}

		// filter out accomplished node
		var newActiveChildren []corev1.LocalObjectReference

		for _, item := range nodeNeedUpdate.Status.ActiveChildren {
			item := item
			if item.Name == childNode.Name {
				continue
			}
			newActiveChildren = append(newActiveChildren, item)
		}

		nodeNeedUpdate.Status.ActiveChildren = newActiveChildren

		if !childrenContains(nodeNeedUpdate.Status.FinishedChildren, childNode.Name) {
			nodeNeedUpdate.Status.FinishedChildren = append(nodeNeedUpdate.Status.FinishedChildren, corev1.LocalObjectReference{Name: childNode.Name})
		}
		return it.kubeClient.Update(ctx, &nodeNeedUpdate)
	})

	return updateError
}

func childrenContains(list []corev1.LocalObjectReference, name string) bool {
	for _, item := range list {
		if item.Name == name {
			return true
		}
	}
	return false
}

func GetKindOf(obj runtime.Object) string {
	t := reflect.TypeOf(obj)
	if t.Kind() != reflect.Ptr {
		panic("All types must be pointers to structs.")
	}
	t = t.Elem()
	return t.Name()
}
