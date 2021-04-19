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

package gc

import (
	"context"
	"reflect"
	"sort"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/go-logr/logr"
	"go.uber.org/fx"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	// In this controller, schedule could be out of date, as the reconcilation may be not caused by
	// an update on Schedule, but by a *Chaos.
	schedule := &v1alpha1.Schedule{}
	err := r.Get(ctx, req.NamespacedName, schedule)
	if err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, nil
	}

	kind, ok := v1alpha1.AllKinds()[string(schedule.Spec.Type)]
	if !ok {
		r.Log.Info("unknown kind", "kind", schedule.Spec.Type)
		r.Recorder.Eventf(schedule, "Warning", "Failed", "Unknown type: %s", schedule.Spec.Type)
		return ctrl.Result{}, nil
	}

	list := kind.ChaosList.DeepCopyObject()
	err = r.List(ctx, list, client.MatchingLabels{"managed-by": schedule.Name})
	if err != nil {
		r.Log.Error(err, "fail to list chaos")
		r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to list chaos: %s", err.Error())
		return ctrl.Result{}, nil
	}

	items := reflect.ValueOf(list).Elem().FieldByName("Items")
	statefulItems := []v1alpha1.StatefulObject{}
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i).Addr().Interface().(v1alpha1.StatefulObject)
		statefulItems = append(statefulItems, item)
	}

	sort.Slice(statefulItems, func(x, y int) bool {
		return statefulItems[x].GetObjectMeta().CreationTimestamp.Time.Before(statefulItems[y].GetObjectMeta().CreationTimestamp.Time)
	})

	// TODO: only delete finished chaos
	exceededHistory := len(statefulItems) - schedule.Spec.HistoryLimit
	if exceededHistory > 0 {
		for _, obj := range statefulItems[0:exceededHistory] {
			err := r.Client.Delete(ctx, obj)
			if err != nil && !k8sError.IsNotFound(err) {
				r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to delete: %s/%s", obj.GetObjectMeta().Namespace, obj.GetObjectMeta().Name)
			}
		}
	}

	return ctrl.Result{}, nil
}

type Objs struct {
	fx.In

	Objs []types.Object `group:"objs"`
}

func NewController(mgr ctrl.Manager, client client.Client, log logr.Logger, objs Objs, scheme *runtime.Scheme) (types.Controller, error) {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Schedule{}).
		Named("schedule-active")

	for _, obj := range objs.Objs {
		// TODO: support workflow
		builder.Owns(obj.Object)
	}

	builder.Complete(&Reconciler{
		client,
		log.WithName("schedule-gc"),
		mgr.GetEventRecorderFor("schedule-gc"),
	})
	return "schedule-gc", nil
}
