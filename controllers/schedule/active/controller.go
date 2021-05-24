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

package active

import (
	"context"
	"reflect"
	"sort"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
)

type Reconciler struct {
	scheme *runtime.Scheme

	client.Client
	Log logr.Logger

	ActiveLister *utils.ActiveLister

	Recorder record.EventRecorder
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	schedule := &v1alpha1.Schedule{}
	err := r.Get(ctx, req.NamespacedName, schedule)
	if err != nil {
		if !k8sError.IsNotFound(err) {
			r.Log.Error(err, "unable to get chaos")
		}
		return ctrl.Result{}, nil
	}

	list, err := r.ActiveLister.ListActiveJobs(ctx, schedule)
	if err != nil {
		r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to list active jobs: %s", err.Error())
		return ctrl.Result{}, nil
	}

	active := []v1.ObjectReference{}
	items := reflect.ValueOf(list).Elem().FieldByName("Items")
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i).Addr().Interface().(runtime.Object)

		ref, err := reference.GetReference(r.scheme, item)
		if err != nil {
			r.Log.Error(err, "fail to get reference")
			r.Recorder.Eventf(schedule, "Warning", "Failed", "Failed to get reference from object: %s", err.Error())
			return ctrl.Result{}, nil
		}

		active = append(active, *ref)
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].Name < active[j].Name
	})
	if reflect.DeepEqual(active, schedule.Status.Active) {
		r.Log.Info("don't need to update active")
		return ctrl.Result{}, nil
	}

	updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		r.Log.Info("updating active", "active", active)
		schedule = schedule.DeepCopyObject().(*v1alpha1.Schedule)

		if err := r.Client.Get(ctx, req.NamespacedName, schedule); err != nil {
			r.Log.Error(err, "unable to get schedule")
			return err
		}

		schedule.Status.Active = active
		return r.Client.Update(ctx, schedule)
	})
	if updateError != nil {
		r.Log.Error(updateError, "fail to update")
		r.Recorder.Eventf(schedule, "Normal", "Failed", "Failed to update active: %s", updateError.Error())
		return ctrl.Result{}, nil
	}

	r.Recorder.Event(schedule, "Normal", "Updated", "Successfully update active of resource")
	return ctrl.Result{}, nil
}

type Objs struct {
	fx.In

	Objs []types.Object `group:"objs"`
}

func NewController(mgr ctrl.Manager, client client.Client, log logr.Logger, objs Objs, scheme *runtime.Scheme, lister *utils.ActiveLister) (types.Controller, error) {
	builder := builder.Default(mgr).
		For(&v1alpha1.Schedule{}).
		Named("schedule-active")

	for _, obj := range objs.Objs {
		// TODO: support workflow
		builder = builder.Owns(obj.Object)
	}

	builder.Complete(&Reconciler{
		scheme,
		client,
		log.WithName("schedule-active"),
		lister,
		mgr.GetEventRecorderFor("schedule-active"),
	})
	return "schedule-active", nil
}
