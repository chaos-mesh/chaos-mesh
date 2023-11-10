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
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

// ChaosCollector represents a collector for Chaos Object.
type ChaosCollector struct {
	client.Client
	Log     logr.Logger
	apiType runtime.Object
	archive core.ExperimentStore
	event   core.EventStore
}

// Reconcile reconciles a chaos collector.
func (r *ChaosCollector) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.apiType == nil {
		r.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}

	obj, ok := r.apiType.DeepCopyObject().(v1alpha1.InnerObject)
	if !ok {
		r.Log.Error(nil, "it's not a stateful object")
		return ctrl.Result{}, nil
	}

	err := r.Get(ctx, req.NamespacedName, obj)
	if apierrors.IsNotFound(err) {
		if err = r.archiveExperiment(req.Namespace, req.Name); err != nil {
			r.Log.Error(err, "failed to archive experiment")
		}

		// If the experiment was created by schedule or workflow,
		// it and its events will be deleted from database.
		if err = r.deleteManagedExperiments(req.Namespace, req.Name); err != nil {
			r.Log.Error(err, "delete managed experiments", "namespace", req.Namespace, "name", req.Name)
		}

		return ctrl.Result{}, nil
	}

	if err != nil {
		r.Log.Error(err, "failed to get chaos object", "request", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if err := r.setUnarchivedExperiment(req, obj); err != nil {
		r.Log.Error(err, "failed to archive experiment")
		// ignore error here
	}

	return ctrl.Result{}, nil
}

// Setup setups collectors by Manager.
func (r *ChaosCollector) Setup(mgr ctrl.Manager, apiType client.Object) error {
	r.apiType = apiType

	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		Complete(r)
}

func (r *ChaosCollector) setUnarchivedExperiment(req ctrl.Request, obj v1alpha1.InnerObject) error {
	archive, err := convertInnerObjectToExperiment(obj)
	if err != nil {
		r.Log.Error(err, "failed to covert InnerObject")
		return err
	}

	find, err := r.archive.FindByUID(context.Background(), archive.UID)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		r.Log.Error(err, "failed to find experiment", "UID", archive.UID)
		return err
	}

	if find != nil {
		archive.ID = find.ID
		archive.CreatedAt = find.CreatedAt
		archive.UpdatedAt = find.UpdatedAt
	}

	if err := r.archive.Set(context.Background(), archive); err != nil {
		r.Log.Error(err, "failed to update experiment", "archive", archive)
		return err
	}

	return nil
}

func (r *ChaosCollector) archiveExperiment(ns, name string) error {
	if err := r.archive.Archive(context.Background(), ns, name); err != nil {
		r.Log.Error(err, "failed to archive experiment", "namespace", ns, "name", name)
		return err
	}

	return nil
}

func (r *ChaosCollector) deleteManagedExperiments(ns, name string) error {
	archives, err := r.archive.FindManagedByNamespaceName(context.Background(), ns, name)
	if gorm.IsRecordNotFoundError(err) {
		return nil
	}

	if err != nil {
		return err
	}

	for _, expr := range archives {
		if err = r.event.DeleteByUID(context.Background(), expr.UID); err != nil {
			r.Log.Error(err, "failed to delete experiment related events")
		}

		if err = r.archive.Delete(context.Background(), expr); err != nil {
			r.Log.Error(err, "failed to delete managed experiment")
		}
	}

	return nil
}

func convertInnerObjectToExperiment(obj v1alpha1.InnerObject) (*core.Experiment, error) {
	chaosMeta, ok := obj.(metav1.Object)
	if !ok {
		return nil, errors.New("chaos meta information not found")
	}
	UID := string(chaosMeta.GetUID())

	archive := &core.Experiment{
		ExperimentMeta: core.ExperimentMeta{
			Namespace: chaosMeta.GetNamespace(),
			Name:      chaosMeta.GetName(),
			Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
			UID:       UID,
			Archived:  false,
		},
	}

	switch chaos := obj.(type) {
	case *v1alpha1.PodChaos:
		archive.Action = string(chaos.Spec.Action)
	case *v1alpha1.NetworkChaos:
		archive.Action = string(chaos.Spec.Action)
	case *v1alpha1.IOChaos:
		archive.Action = string(chaos.Spec.Action)
	case *v1alpha1.TimeChaos, *v1alpha1.KernelChaos, *v1alpha1.StressChaos, *v1alpha1.HTTPChaos:
		archive.Action = ""
	case *v1alpha1.DNSChaos:
		archive.Action = string(chaos.Spec.Action)
	case *v1alpha1.PhysicalMachineChaos:
		archive.Action = string(chaos.Spec.Action)
	case *v1alpha1.AWSChaos:
		archive.Action = string(chaos.Spec.Action)
	case *v1alpha1.GCPChaos:
		archive.Action = string(chaos.Spec.Action)
	case *v1alpha1.JVMChaos:
		archive.Action = string(chaos.Spec.Action)
	case *v1alpha1.BlockChaos:
		archive.Action = string(chaos.Spec.Action)
	case *v1alpha1.K8SChaos:
		archive.Action = ""
	case *v1alpha1.CloudStackVMChaos:
		archive.Action = ""
	default:
		return nil, errors.New("unsupported chaos type " + archive.Kind)
	}

	archive.StartTime = chaosMeta.GetCreationTimestamp().Time
	if chaosMeta.GetDeletionTimestamp() != nil {
		archive.FinishTime = &chaosMeta.GetDeletionTimestamp().Time
	}

	data, err := json.Marshal(chaosMeta)
	if err != nil {
		return nil, err
	}

	archive.Experiment = string(data)

	return archive, nil
}
