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
	"github.com/pkg/errors"
	"gorm.io/gorm"
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
	Log        logr.Logger
	apiType    runtime.Object
	store      core.ExperimentStore
	eventStore core.EventStore
}

// Setup setups collectors by Manager.
func (r *ChaosCollector) Setup(mgr ctrl.Manager, apiType client.Object) error {
	r.apiType = apiType

	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		Complete(r)
}

// Reconcile reconciles a chaos collector.
func (r *ChaosCollector) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if err := ctx.Err(); err != nil {
		return ctrl.Result{}, err
	}
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
	if err := ctx.Err(); err != nil {
		return ctrl.Result{}, err
	}
	if apierrors.IsNotFound(err) {
		if err = r.archiveExperiment(ctx, req.Namespace, req.Name); err != nil {
			r.Log.Error(err, "failed to archive experiment")
			return retryCollector(ctx)
		}

		// If the experiment was created by schedule or workflow,
		// it and its events should be deleted from database.
		if err = r.deleteManagedExperiments(ctx, req.Namespace, req.Name); err != nil {
			r.Log.Error(err, "delete managed experiments", "namespace", req.Namespace, "name", req.Name)
			return retryCollector(ctx)
		}

		return ctrl.Result{}, nil
	}

	if err != nil {
		r.Log.Error(err, "failed to get chaos object", "request", req.NamespacedName)
		return retryCollector(ctx)
	}

	if err := r.createOrUpdateExperiment(ctx, obj); err != nil {
		r.Log.Error(err, "failed to create or update experiment in store")
		return retryCollector(ctx)
	}

	return ctrl.Result{}, nil
}

func (r *ChaosCollector) createOrUpdateExperiment(ctx context.Context, obj v1alpha1.InnerObject) error {
	exp, err := convertInnerObjectToExperiment(obj)
	if err != nil {
		return err
	}

	found, err := r.store.FindByUID(ctx, exp.UID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.store.Save(ctx, exp)
	}

	if err != nil {
		return err
	}

	// Chaos spec is immutable after creation. Only persist finish time updates.
	if exp.FinishTime == nil {
		return nil
	}

	if found.FinishTime != nil && found.FinishTime.Equal(*exp.FinishTime) {
		return nil
	}

	found.FinishTime = exp.FinishTime
	return r.store.Save(ctx, found)

}

func (r *ChaosCollector) archiveExperiment(ctx context.Context, ns, name string) error {
	if err := r.store.Archive(ctx, ns, name); err != nil {
		r.Log.Error(err, "failed to archive experiment", "namespace", ns, "name", name)
		return err
	}

	return nil
}

func (r *ChaosCollector) deleteManagedExperiments(ctx context.Context, ns, name string) error {
	archives, err := r.store.ListArchivedManagedByNamespaceName(ctx, ns, name)
	if err != nil {
		return err
	}

	if len(archives) == 0 {
		return nil
	}

	for _, exp := range archives {
		if err = r.eventStore.DeleteByUID(ctx, exp.UID); err != nil {
			r.Log.Error(err, "failed to delete experiment related events")
			return err
		}

		if err = r.store.Delete(ctx, exp); err != nil {
			r.Log.Error(err, "failed to delete managed experiment")
			return err
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

	experiment := &core.Experiment{
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
		experiment.Action = string(chaos.Spec.Action)
	case *v1alpha1.NetworkChaos:
		experiment.Action = string(chaos.Spec.Action)
	case *v1alpha1.IOChaos:
		experiment.Action = string(chaos.Spec.Action)
	case *v1alpha1.TimeChaos, *v1alpha1.KernelChaos, *v1alpha1.StressChaos, *v1alpha1.HTTPChaos:
		experiment.Action = ""
	case *v1alpha1.DNSChaos:
		experiment.Action = string(chaos.Spec.Action)
	case *v1alpha1.PhysicalMachineChaos:
		experiment.Action = string(chaos.Spec.Action)
	case *v1alpha1.AWSChaos:
		experiment.Action = string(chaos.Spec.Action)
	case *v1alpha1.GCPChaos:
		experiment.Action = string(chaos.Spec.Action)
	case *v1alpha1.AzureChaos:
		experiment.Action = string(chaos.Spec.Action)
	case *v1alpha1.JVMChaos:
		experiment.Action = string(chaos.Spec.Action)
	case *v1alpha1.BlockChaos:
		experiment.Action = string(chaos.Spec.Action)
	default:
		return nil, errors.New("unsupported chaos type " + experiment.Kind)
	}

	experiment.StartTime = chaosMeta.GetCreationTimestamp().Time
	if chaosMeta.GetDeletionTimestamp() != nil {
		experiment.FinishTime = &chaosMeta.GetDeletionTimestamp().Time
	}

	data, err := json.Marshal(chaosMeta)
	if err != nil {
		return nil, err
	}

	experiment.Experiment = string(data)

	return experiment, nil
}

func retryCollector(ctx context.Context) (ctrl.Result, error) {
	if err := ctx.Err(); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}
