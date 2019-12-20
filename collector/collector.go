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

package collector

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/api_interface"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	initLog    = ctrl.Log.WithName("setup")
	dataSource string
)

func init() {
	namespace, ok := os.LookupEnv("NAMESPACE")
	if !ok {
		initLog.Error(nil, "cannot find NAMESPACE")
		namespace = "chaos"
	}

	dataSource = fmt.Sprintf("root:@tcp(chaos-collector-database.%s:3306)/chaos_operator", namespace)
}

type ChaosCollector struct {
	client.Client
	Log            logr.Logger
	apiType        runtime.Object
	databaseClient *DatabaseClient
}

func (r *ChaosCollector) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	if r.apiType == nil {
		r.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}
	ctx := context.Background()

	obj, ok := r.apiType.DeepCopyObject().(api_interface.StatefulObject)
	if !ok {
		r.Log.Error(nil, "it's not a stateful object")
		return ctrl.Result{}, nil
	}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, nil
	}

	status := obj.GetStatus()

	if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
		event := Event{
			Name:      req.Name,
			Namespace: req.Namespace,
			StartTime: &status.Experiment.StartTime.Time,
			EndTime:   nil,
		}
		r.Log.Info("event started, save to database", "event", event)

		err := r.databaseClient.WriteEvent(event)
		if err != nil {
			r.Log.Error(err, "write event to database error")
			return ctrl.Result{}, nil
		}
	} else if status.Experiment.Phase == v1alpha1.ExperimentPhaseFinished {
		event := Event{
			Name:      req.Name,
			Namespace: req.Namespace,
			StartTime: &status.Experiment.StartTime.Time,
			EndTime:   &status.Experiment.EndTime.Time,
		}
		r.Log.Info("event finished, save to database", "event", event)

		err := r.databaseClient.UpdateEvent(event)
		if err != nil {
			r.Log.Error(err, "write event to database error")
			return ctrl.Result{}, nil
		}
	}
	return ctrl.Result{}, nil
}

func (r *ChaosCollector) Setup(mgr ctrl.Manager, apiType runtime.Object) error {
	r.apiType = apiType

	databaseClient, err := NewDatabaseClient(dataSource)
	if err != nil {
		r.Log.Error(err, "create database client failed")
		return nil
	}

	r.databaseClient = databaseClient
	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		Complete(r)
}
