// Copyright 2019 Chaos Mesh Authors.
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

package collector

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-logr/logr"
	"github.com/jinzhu/gorm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
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
func (r *ChaosCollector) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var (
		chaosMeta  metav1.Object
		ok         bool
		manageFlag bool
	)

	if r.apiType == nil {
		r.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}
	ctx := context.Background()
	manageFlag = false

	obj, ok := r.apiType.DeepCopyObject().(v1alpha1.InnerObject)
	if !ok {
		r.Log.Error(nil, "it's not a stateful object")
		return ctrl.Result{}, nil
	}

	err := r.Get(ctx, req.NamespacedName, obj)
	if apierrors.IsNotFound(err) {
		if chaosMeta, ok = obj.(metav1.Object); !ok {
			r.Log.Error(nil, "failed to get chaos meta information")
		}
		if chaosMeta.GetLabels()["managed-by"] != "" {
			manageFlag = true
		}
		if !manageFlag {
			if err = r.archiveExperiment(req.Namespace, req.Name); err != nil {
				r.Log.Error(err, "failed to archive experiment")
			}
		} else {
			if err = r.event.DeleteByUID(ctx, string(chaosMeta.GetUID())); err != nil {
				r.Log.Error(err, "failed to delete experiment related events")
			}
		}
		return ctrl.Result{}, nil
	}

	if err != nil {
		r.Log.Error(err, "failed to get chaos object", "request", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if chaosMeta, ok = obj.(metav1.Object); !ok {
		r.Log.Error(nil, "failed to get chaos meta information")
	}

	if chaosMeta.GetLabels()["managed-by"] != "" {
		manageFlag = true
	}

	if obj.IsDeleted() {
		if !manageFlag {
			if err = r.archiveExperiment(req.Namespace, req.Name); err != nil {
				r.Log.Error(err, "failed to archive experiment")
			}
		} else {
			if err = r.event.DeleteByUID(ctx, string(chaosMeta.GetUID())); err != nil {
				r.Log.Error(err, "failed to delete experiment related events")
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.setUnarchivedExperiment(req, obj); err != nil {
		r.Log.Error(err, "failed to archive experiment")
		// ignore error here
	}

	return ctrl.Result{}, nil
}

// Setup setups collectors by Manager.
func (r *ChaosCollector) Setup(mgr ctrl.Manager, apiType runtime.Object) error {
	r.apiType = apiType

	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		Complete(r)
}

func (r *ChaosCollector) setUnarchivedExperiment(req ctrl.Request, obj v1alpha1.InnerObject) error {
	var (
		chaosMeta metav1.Object
		ok        bool
	)

	if chaosMeta, ok = obj.(metav1.Object); !ok {
		r.Log.Error(nil, "failed to get chaos meta information")
	}
	UID := string(chaosMeta.GetUID())

	archive := &core.Experiment{
		ExperimentMeta: core.ExperimentMeta{
			Namespace: req.Namespace,
			Name:      req.Name,
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
	case *v1alpha1.TimeChaos, *v1alpha1.KernelChaos, *v1alpha1.StressChaos:
		archive.Action = ""
	case *v1alpha1.DNSChaos:
		archive.Action = string(chaos.Spec.Action)
	default:
		return errors.New("unsupported chaos type " + archive.Kind)
	}

	archive.StartTime = chaosMeta.GetCreationTimestamp().Time
	if chaosMeta.GetDeletionTimestamp() != nil {
		archive.FinishTime = chaosMeta.GetDeletionTimestamp().Time
	}

	data, err := json.Marshal(chaosMeta)
	if err != nil {
		r.Log.Error(err, "failed to marshal chaos", "kind", archive.Kind,
			"namespace", archive.Namespace, "name", archive.Name)
		return err
	}

	archive.Experiment = string(data)

	find, err := r.archive.FindByUID(context.Background(), UID)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		r.Log.Error(err, "failed to find experiment", "UID", UID)
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
