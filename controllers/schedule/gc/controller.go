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
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	corev1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/controllers"
)

type Reconciler struct {
	client.Client
	Log      logr.Logger
	Recorder recorder.ChaosRecorder

	ActiveLister *utils.ActiveLister
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	// In this controller, schedule could be out of date, as the reconcilation may be not caused by
	// an update on Schedule, but by a *Chaos.
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
		r.Recorder.Event(schedule, recorder.Failed{
			Activity: "list active jobs",
			Err:      err.Error(),
		})
		return ctrl.Result{}, nil
	}

	items := reflect.ValueOf(list).Elem().FieldByName("Items")
	metaItems := []v1alpha1.MetaObject{}
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i).Addr().Interface().(v1alpha1.MetaObject)
		metaItems = append(metaItems, item)
	}

	sort.Slice(metaItems, func(x, y int) bool {
		return metaItems[x].GetObjectMeta().CreationTimestamp.Time.Before(metaItems[y].GetObjectMeta().CreationTimestamp.Time)
	})

	exceededHistory := len(metaItems) - schedule.Spec.HistoryLimit
	requeuAfter := time.Duration(0)
	if exceededHistory > 0 {
		for _, obj := range metaItems[0:exceededHistory] {
			innerObj, ok := obj.(v1alpha1.InnerObject)
			if ok { // This is a chaos
				finished, untilStop := controller.IsChaosFinishedWithUntilStop(innerObj, time.Now())

				if !finished {
					if untilStop != 0 {
						if requeuAfter == 0 || requeuAfter > untilStop {
							requeuAfter = untilStop
						}

						r.Recorder.Event(schedule, recorder.ScheduleSkipRemoveHistory{
							RunningName: innerObj.GetChaos().Name,
						})
						continue
					}

					// hasn't finished, but untilStop is 0
					r.Log.Info("untilStop is 0 when the chaos has not finished")
				}
			} else { // A workflow
				if schedule.Spec.Type == v1alpha1.ScheduleTypeWorkflow {
					workflow, ok := obj.(*v1alpha1.Workflow)
					if ok {
						finished := controllers.WorkflowConditionEqualsTo(workflow.Status, v1alpha1.WorkflowConditionAccomplished, corev1.ConditionTrue)

						if !finished {
							r.Recorder.Event(schedule, recorder.ScheduleSkipRemoveHistory{
								RunningName: workflow.Name,
							})
							continue
						}
					}
				}
			}
			err := r.Client.Delete(ctx, obj)
			if err != nil && !k8sError.IsNotFound(err) {
				r.Recorder.Event(schedule, recorder.Failed{
					Activity: fmt.Sprintf("delete %s/%s", obj.GetObjectMeta().Namespace, obj.GetObjectMeta().Name),
					Err:      err.Error(),
				})
			}
		}
	}

	return ctrl.Result{
		RequeueAfter: requeuAfter,
	}, nil
}

type Objs struct {
	fx.In

	ScheduleObjs []types.Object `group:"schedule"`
	Objs         []types.Object `group:"objs"`
}

func NewController(mgr ctrl.Manager, client client.Client, log logr.Logger, objs Objs, scheme *runtime.Scheme, lister *utils.ActiveLister, recorderBuilder *recorder.RecorderBuilder) (types.Controller, error) {
	builder := builder.Default(mgr).
		For(&v1alpha1.Schedule{}).
		Named("schedule-gc")

	for _, obj := range objs.Objs {
		// TODO: support workflow
		builder.Owns(obj.Object)
	}

	builder = builder.Owns(&v1alpha1.Workflow{})

	builder.Complete(&Reconciler{
		client,
		log.WithName("schedule-gc"),
		recorderBuilder.Build("schedule-gc"),
		lister,
	})
	return "schedule-gc", nil
}
