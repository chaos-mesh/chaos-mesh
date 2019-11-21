// Copyright 2019 PingCAP, Inc.
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

package twophase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/utils"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
)

type InnerObject interface {
	runtime.Object

	IsDeleted() bool

	GetDuration() (time.Duration, error)

	GetNextStart() time.Time
	SetNextStart(time.Time)

	GetNextRecover() time.Time
	SetNextRecover(time.Time)

	GetScheduler() v1alpha1.SchedulerSpec
}

type InnerReconciler interface {
	Apply(ctx context.Context, req ctrl.Request, chaos InnerObject) error

	Recover(ctx context.Context, req ctrl.Request, chaos InnerObject) error

	Object() InnerObject
}

type Reconciler struct {
	InnerReconciler
	client.Client
	Log logr.Logger
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error
	now := time.Now()

	r.Log.Info("reconciling a two phase chaos")
	ctx := context.Background()

	chaos := r.Object()
	if err = r.Get(ctx, req.NamespacedName, chaos); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, utils.IgnoreNotFound(err)
	}

	duration, err := chaos.GetDuration()
	if err != nil {
		return ctrl.Result{}, err
	}

	ctx = context.Background()
	if chaos.IsDeleted() {
		// This chaos was deleted
		r.Log.Info("Removing self")
		err = r.Recover(ctx, req, chaos)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	} else if !chaos.GetNextRecover().IsZero() && chaos.GetNextRecover().Before(now) {
		// Start recover
		r.Log.Info("Recovering")

		err = r.Recover(ctx, req, chaos)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		chaos.SetNextRecover(time.Time{})
	} else if chaos.GetNextStart().Before(now) {
		nextStart, err := utils.NextTime(chaos.GetScheduler(), now)
		nextRecover := now.Add(duration)
		if nextStart.Before(nextRecover) {
			err := fmt.Errorf("nextRecover shouldn't be later than nextStart")
			r.Log.Error(err, "nextRecover is later than nextStart. Then recover can never be reached", "nextRecover", nextRecover, "nextStart", nextStart)
			return ctrl.Result{}, err
		}

		r.Log.Info("now chaos action:", "chaos", chaos)

		// Start failure action
		r.Log.Info("Performing Action")

		err = r.Apply(ctx, req, chaos)
		if err != nil {
			return ctrl.Result{}, err
		}

		chaos.SetNextStart(*nextStart)
		chaos.SetNextRecover(nextRecover)
	} else {
		nextTime := chaos.GetNextStart()

		if !chaos.GetNextRecover().IsZero() && chaos.GetNextRecover().Before(nextTime) {
			nextTime = chaos.GetNextRecover()
		}
		duration := nextTime.Sub(now)
		r.Log.Info("requeue request", "after", duration)

		return ctrl.Result{RequeueAfter: duration}, nil
	}

	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaosctl status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
