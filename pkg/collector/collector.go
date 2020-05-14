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
	"github.com/go-logr/logr"
	"github.com/jinzhu/gorm"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/reconciler"
	"github.com/pingcap/chaos-mesh/pkg/core"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ChaosCollector represents a collector for Chaos Object.
type ChaosCollector struct {
	client.Client
	Log       logr.Logger
	apiType   runtime.Object
	archive   core.ExperimentStore
	event     core.EventStore
	podRecord core.PodRecordStore
}

// Reconcile reconciles a chaos collector.
func (r *ChaosCollector) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	if r.apiType == nil {
		r.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}
	ctx := context.Background()

	obj, ok := r.apiType.DeepCopyObject().(reconciler.InnerObject)
	if !ok {
		r.Log.Error(nil, "it's not a stateful object")
		return ctrl.Result{}, nil
	}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, nil
	}

	if err := r.recordEvent(req, obj); err != nil {
		r.Log.Error(err, "failed to record event")
		return ctrl.Result{}, nil
	}

	// TODO: archive experiment

	return ctrl.Result{}, nil
}

// Setup setups collectors by Manager.
func (r *ChaosCollector) Setup(mgr ctrl.Manager, apiType runtime.Object) error {
	r.apiType = apiType

	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		Complete(r)
}

func (r *ChaosCollector) recordEvent(req ctrl.Request, obj reconciler.InnerObject) error {
	status := obj.GetStatus()
	kind := obj.GetObjectKind().GroupVersionKind().Kind

	switch status.Experiment.Phase {
	case v1alpha1.ExperimentPhaseRunning:
		return r.createEvent(req, kind, status)
	case v1alpha1.ExperimentPhaseFinished, v1alpha1.ExperimentPhasePaused:
		return r.updateOrCreateEvent(req, kind, status)
	}

	return nil
}

func (r *ChaosCollector) createEvent(req ctrl.Request, kind string, status *v1alpha1.ChaosStatus) error {
	event := &core.Event{
		Experiment: req.Name,
		Namespace:  req.Namespace,
		Kind:       kind,
		StartTime:  &status.Experiment.StartTime.Time,
		FinishTime: nil,
	}

	if err := r.event.Create(context.Background(), event); err != nil {
		r.Log.Error(err, "failed to store event", "event", event)
		return err
	}

	for _, pod := range status.Experiment.Pods {
		podRecord := &core.PodRecord{
			EventID:   event.ID,
			PodIP:     pod.PodIP,
			PodName:   pod.Name,
			Namespace: pod.Namespace,
			Message:   pod.Message,
			Action:    pod.Action,
		}

		if err := r.podRecord.Create(context.Background(), podRecord); err != nil {
			r.Log.Error(err, "failed to store pod record", "podRecord", podRecord)
			return err
		}
	}

	return nil
}

func (r *ChaosCollector) updateOrCreateEvent(req ctrl.Request, kind string, status *v1alpha1.ChaosStatus) error {
	event := &core.Event{
		Experiment: req.Name,
		Namespace:  req.Namespace,
		Kind:       kind,
		StartTime:  &status.Experiment.StartTime.Time,
		FinishTime: &status.Experiment.EndTime.Time,
	}

	if _, err := r.event.FindByExperimentAndStartTime(
		context.Background(), event.Experiment, event.Namespace, event.StartTime); err != nil && gorm.IsRecordNotFoundError(err) {
		if err := r.createEvent(req, kind, status); err != nil {
			return err
		}
	}

	if err := r.event.Update(context.Background(), event); err != nil {
		r.Log.Error(err, "failed to update event", "event", event)
		return err
	}

	return nil
}
