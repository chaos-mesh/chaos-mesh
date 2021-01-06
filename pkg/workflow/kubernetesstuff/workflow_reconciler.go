// Copyright 2020 Chaos Mesh Authors.
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

package kubernetesstuff

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	workflowv1alpha1 "github.com/chaos-mesh/chaos-mesh/pkg/workflow/apis/workflow/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

type WorkflowReconciler struct {
	logger          logr.Logger
	kubeClient      client.Client
	operableTrigger trigger.OperableTrigger
}

func NewWorkflowReconciler(logger logr.Logger, kubeClient client.Client, operableTrigger trigger.OperableTrigger) *WorkflowReconciler {
	return &WorkflowReconciler{logger: logger, kubeClient: kubeClient, operableTrigger: operableTrigger}
}

func (it *WorkflowReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	target := workflowv1alpha1.Workflow{}
	err := it.kubeClient.Get(context.Background(), request.NamespacedName, &target)
	if err != nil {
		return reconcile.Result{}, err
	}
	if target.Status.Phase == workflow.Init {
		err := it.operableTrigger.Notify(trigger.NewEvent(request.Namespace, request.Name, "", trigger.WorkflowCreated))
		if err != nil {
			it.logger.Error(err, "failed to notify WorkflowCreated event")
		}
	}
	return reconcile.Result{}, nil
}
