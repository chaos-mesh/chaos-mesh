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

package collector

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"gorm.io/gorm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type WorkflowCollector struct {
	kubeClient client.Client
	Log        logr.Logger
	apiType    runtime.Object
	store      core.WorkflowStore
}

func (it *WorkflowCollector) Setup(mgr ctrl.Manager, apiType client.Object) error {
	it.apiType = apiType

	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		Complete(it)
}

func (it *WorkflowCollector) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	if it.apiType == nil {
		it.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}
	workflow := v1alpha1.Workflow{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &workflow)
	if apierrors.IsNotFound(err) {
		// target
		if err = it.markAsArchived(ctx, request.Namespace, request.Name); err != nil {
			it.Log.Error(err, "failed to archive experiment")
		}
		return ctrl.Result{}, nil
	}
	if err != nil {
		it.Log.Error(err, "failed to get workflow object", "request", request.NamespacedName)
		return ctrl.Result{}, nil
	}

	if err := it.persistentWorkflow(&workflow); err != nil {
		it.Log.Error(err, "failed to archive workflow")
	}

	return ctrl.Result{}, nil
}

func (it *WorkflowCollector) markAsArchived(ctx context.Context, namespace, name string) error {
	return it.store.MarkAsArchived(ctx, namespace, name)
}

func (it *WorkflowCollector) persistentWorkflow(workflow *v1alpha1.Workflow) error {
	newEntity, err := core.WorkflowCR2WorkflowEntity(workflow)
	if err != nil {
		return err
	}

	existedEntity, err := it.store.FindByUID(context.Background(), string(workflow.UID))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		it.Log.Error(err, "failed to find workflow", "UID", workflow.UID)
		return err
	}

	if existedEntity != nil {
		newEntity.ID = existedEntity.ID
	}

	err = it.store.Save(context.Background(), newEntity)
	if err != nil {
		it.Log.Error(err, "failed to update workflow", "archive", newEntity)
	}
	return err
}
