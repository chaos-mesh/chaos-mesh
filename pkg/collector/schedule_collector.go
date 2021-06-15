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

package collector

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-logr/logr"
	"github.com/jinzhu/gorm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
)

// ScheduleCollector represents a collector for Schedule Object.
type ScheduleCollector struct {
	client.Client
	Log     logr.Logger
	apiType runtime.Object
	archive core.ScheduleStore
}

// Reconcile reconciles a Schedule collector.
func (r *ScheduleCollector) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	if r.apiType == nil {
		r.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}
	ctx := context.Background()

	schedule := &v1alpha1.Schedule{}
	err := r.Get(ctx, req.NamespacedName, schedule)
	if apierrors.IsNotFound(err) {
		if err = r.archiveSchedule(req.Namespace, req.Name); err != nil {
			r.Log.Error(err, "failed to archive schedule")
		}
		return ctrl.Result{}, nil
	}
	if err != nil {
		r.Log.Error(err, "failed to get schedule object", "request", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if !schedule.DeletionTimestamp.IsZero() {
		if err = r.archiveSchedule(req.Namespace, req.Name); err != nil {
			r.Log.Error(err, "failed to archive schedule")
		}
		return ctrl.Result{}, nil
	}

	if err := r.setUnarchivedSchedule(req, *schedule); err != nil {
		r.Log.Error(err, "failed to archive schedule")
		// ignore error here
	}

	return ctrl.Result{}, nil
}

// Setup setups collectors by Manager.
func (r *ScheduleCollector) Setup(mgr ctrl.Manager, apiType runtime.Object) error {
	r.apiType = apiType

	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		Complete(r)
}

func (r *ScheduleCollector) setUnarchivedSchedule(req ctrl.Request, schedule v1alpha1.Schedule) error {
	archive := &core.Schedule{
		ScheduleMeta: core.ScheduleMeta{
			Namespace: req.Namespace,
			Name:      req.Name,
			Kind:      schedule.Kind,
			UID:       string(schedule.UID),
			Archived:  false,
		},
	}

	switch schedule.Spec.Type {
	case v1alpha1.ScheduleTypePodChaos:
		archive.Action = string(schedule.Spec.ScheduleItem.PodChaos.Action)
	case v1alpha1.ScheduleTypeNetworkChaos:
		archive.Action = string(schedule.Spec.ScheduleItem.NetworkChaos.Action)
	case v1alpha1.ScheduleTypeIOChaos:
		archive.Action = string(schedule.Spec.ScheduleItem.IOChaos.Action)
	case v1alpha1.ScheduleTypeTimeChaos, v1alpha1.ScheduleTypeKernelChaos, v1alpha1.ScheduleTypeStressChaos:
		archive.Action = ""
	case v1alpha1.ScheduleTypeDNSChaos:
		archive.Action = string(schedule.Spec.ScheduleItem.DNSChaos.Action)
	case v1alpha1.ScheduleTypeAwsChaos:
		archive.Action = string(schedule.Spec.ScheduleItem.AwsChaos.Action)
	case v1alpha1.ScheduleTypeGcpChaos:
		archive.Action = string(schedule.Spec.ScheduleItem.GcpChaos.Action)
	default:
		return errors.New("unsupported chaos type " + string(schedule.Spec.Type))
	}

	archive.StartTime = schedule.GetCreationTimestamp().Time
	if schedule.GetDeletionTimestamp() != nil {
		archive.FinishTime = schedule.GetDeletionTimestamp().Time
	}

	data, err := json.Marshal(schedule)
	if err != nil {
		r.Log.Error(err, "failed to marshal schedule", "kind", archive.Kind,
			"namespace", archive.Namespace, "name", archive.Name)
		return err
	}

	archive.Schedule = string(data)

	find, err := r.archive.FindByUID(context.Background(), string(schedule.UID))
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		r.Log.Error(err, "failed to find schedule", "UID", schedule.UID)
		return err
	}

	if find != nil {
		archive.ID = find.ID
		archive.CreatedAt = find.CreatedAt
		archive.UpdatedAt = find.UpdatedAt
	}

	if err := r.archive.Set(context.Background(), archive); err != nil {
		r.Log.Error(err, "failed to update schedule", "archive", archive)
		return err
	}

	return nil
}

func (r *ScheduleCollector) archiveSchedule(ns, name string) error {
	if err := r.archive.Archive(context.Background(), ns, name); err != nil {
		r.Log.Error(err, "failed to archive schedule", "namespace", ns, "name", name)
		return err
	}

	return nil
}
