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
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func BootstrapWorkflowControllers(mgr manager.Manager, logger logr.Logger) error {

	noCacheClient, err := client.New(mgr.GetConfig(), client.Options{
		Scheme: mgr.GetScheme(),
		Mapper: mgr.GetRESTMapper(),
	})
	if err != nil {
		return err
	}
	err = ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Workflow{}).
		Named("workflow-entry-reconciler").
		Complete(
			NewWorkflowEntryReconciler(
				mgr.GetClient(),
				mgr.GetEventRecorderFor("workflow-entry-reconciler"),
				logger.WithName("workflow-entry-reconciler"),
			),
		)
	if err != nil {
		return err
	}

	// TODO: serial node reconciler restore some state in the workflow node status(the active children), it requires keep syncing in time, so we could not use the default controller-runtime client with cache
	// TODO: maybe we could use select with labelSelector as instead
	err = ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.WorkflowNode{}).
		Named("workflow-serial-node-reconciler").
		Complete(
			NewSerialNodeReconciler(
				noCacheClient,
				mgr.GetEventRecorderFor("workflow-serial-node-reconciler"),
				logger.WithName("workflow-serial-node-reconciler"),
			),
		)
	if err != nil {
		return err
	}
	err = ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.WorkflowNode{}).
		Named("workflow-accomplish-watcher").
		Complete(
			NewAccomplishWatcher(
				mgr.GetClient(),
				mgr.GetEventRecorderFor("workflow-accomplish-watcher"),
				logger.WithName("workflow-accomplish-watcher"),
			),
		)
	if err != nil {
		return err
	}
	err = ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.WorkflowNode{}).
		Named("workflow-deadline-reconciler").
		Complete(
			NewDeadlineReconciler(
				mgr.GetClient(),
				mgr.GetEventRecorderFor("workflow-deadline-reconciler"),
				logger.WithName("workflow-deadline-reconciler"),
			),
		)
	if err != nil {
		return err
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.WorkflowNode{}).
		Named("workflow-chaos-node-reconciler").
		Complete(
			NewChaosNodeReconciler(
				mgr.GetClient(),
				mgr.GetEventRecorderFor("workflow-chaos-node-reconciler"),
				logger.WithName("workflow-chaos-node-reconciler"),
			),
		)
	if err != nil {
		return err
	}
	return nil
}
