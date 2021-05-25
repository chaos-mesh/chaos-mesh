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

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ScheduleCollector represents a collector for Schedule Object.
type ScheduleCollector struct {
	client.Client
	Log     logr.Logger
	apiType runtime.Object
	archive core.ScheduleStore
	event   core.EventStore
}

// Reconcile reconciles a Schedule collector.
func (r *ScheduleCollector) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	if r.apiType == nil {
		r.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}
	ctx := context.Background()

	obj, ok := r.apiType.DeepCopyObject().(v1alpha1.InnerObject)
	if !ok {
		r.Log.Error(nil, "it's not a stateful object")
		return ctrl.Result{}, nil
	}

	err := r.Get(ctx, req.NamespacedName, obj)
	if apierrors.IsNotFound(err) {
		if err = r.archiveSchedule(req.Namespace, req.Name); err != nil {
			r.Log.Error(err, "failed to archive experiment")
		}
		return ctrl.Result{}, nil
	}

	if err != nil {
		r.Log.Error(err, "failed to get chaos object", "request", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if obj.IsDeleted() {
		if err = r.archiveSchedule(req.Namespace, req.Name); err != nil {
			r.Log.Error(err, "failed to archive experiment")
		}
		return ctrl.Result{}, nil
	}

	if err := r.setUnarchivedSchedule(req, obj); err != nil {
		r.Log.Error(err, "failed to archive experiment")
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

func (r *ScheduleCollector) setUnarchivedSchedule(req ctrl.Request, obj v1alpha1.InnerObject) error {
	ctx := context.Background()

	schedule := &v1alpha1.Schedule{}
	err := r.Get(ctx, req.NamespacedName, schedule)
	if err != nil {
		r.Log.Error(err, "unable to get schedule")
		return nil
	}

	archive := &core.Schedule{
		ScheduleMeta: core.ScheduleMeta{
			Namespace: req.Namespace,
			Name:      req.Name,
			Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
			UID:       string(schedule.UID),
			Archived:  false,
		},
	}

	switch schedule.Spec.Type {
	case v1alpha1.TypePodChaos:
		archive.Action = string(schedule.Spec.PodChaos.Action)
	case v1alpha1.TypeNetworkChaos:
		archive.Action = string(schedule.Spec.NetworkChaos.Action)
	case v1alpha1.TypeIoChaos:
		archive.Action = string(schedule.Spec.IoChaos.Action)
	case v1alpha1.TypeTimeChaos, v1alpha1.TypeKernelChaos, v1alpha1.TypeStressChaos:
		archive.Action = ""
	case v1alpha1.TypeDNSChaos:
		archive.Action = string(schedule.Spec.DNSChaos.Action)
	case v1alpha1.TypeAwsChaos:
		archive.Action = string(schedule.Spec.AwsChaos.Action)
	case v1alpha1.TypeGcpChaos:
		archive.Action = string(schedule.Spec.GcpChaos.Action)
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
