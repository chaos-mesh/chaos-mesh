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
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ChaosNodeReconciler struct {
	kubeClient    client.Client
	eventRecorder record.EventRecorder
	logger        logr.Logger
}

func NewChaosNodeReconciler(kubeClient client.Client, eventRecorder record.EventRecorder, logger logr.Logger) *ChaosNodeReconciler {
	return &ChaosNodeReconciler{kubeClient: kubeClient, eventRecorder: eventRecorder, logger: logger}
}

func (it *ChaosNodeReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	node := v1alpha1.WorkflowNode{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1alpha1.WorkflowNodeSpec{},
		Status:     v1alpha1.WorkflowNodeStatus{},
	}

	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	return reconcile.Result{}, nil
}
