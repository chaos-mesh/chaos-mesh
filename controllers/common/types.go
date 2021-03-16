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

package common

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/container"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InnerObjectWithSelector interface {
	v1alpha1.InnerObject

	GetSelectorSpecs() map[string]interface{}
}

type ChaosImpl interface {
	Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error)
	Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error)
}

// Reconciler for common chaos
type Reconciler struct {
	Impl ChaosImpl

	// Object is used to mark the target type of this Reconciler
	Object InnerObjectWithSelector

	// Client is used to operate on the Kubernetes cluster
	client.Client
	client.Reader

	Log logr.Logger
}

// Reconcile the common chaos
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	obj := r.Object.DeepCopyObject().(InnerObjectWithSelector)

	if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("chaos not found")
		} else {
			// TODO: handle this error
			r.Log.Error(err, "unable to get chaos")
		}
		return ctrl.Result{}, nil
	}

	shouldUpdate := false

	status := obj.GetStatus()
	if status.Experiment.Records == nil {
		var records []*v1alpha1.Record
		// TODO: get selectors from obj
		for name, sel := range obj.GetSelectorSpecs() {
			selector := selector.New(selector.SelectorParams{
				PodSelector: pod.New(r.Client, r.Reader, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces),
				ContainerSelector: container.New(r.Client, r.Reader, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces),
			})
			targets, err := selector.Select(context.TODO(), sel)
			if err != nil {
				// TODO: handle this error
				r.Log.Error(err, "fail to select")
			}

			for _, target := range targets {
				records = append(records, &v1alpha1.Record{
					Id: target.Id(),
					SelectorKey: name,
					Phase: v1alpha1.NotInjected,
				})
			}
		}

		status.Experiment.Records = records
		shouldUpdate = true
		// TODO: dynamic upgrade the records when some of these pods/containers stopped
	}

	// TODO: seperate the defaulter logic to another place
	if status.Experiment.DesiredPhase == "" {
		status.Experiment.DesiredPhase = v1alpha1.RunningPhase
	}

	for index, record := range status.Experiment.Records {
		var err error
		if status.Experiment.DesiredPhase == v1alpha1.RunningPhase && record.Phase != v1alpha1.Injected {
			record.Phase, err = r.Impl.Apply(context.TODO(), index, status.Experiment.Records, obj)
			if err != nil {
				// TODO: handle this error
				r.Log.Error(err, "fail to apply chaos")
			}
		}
		if status.Experiment.DesiredPhase == v1alpha1.StoppedPhase && record.Phase != v1alpha1.NotInjected {
			record.Phase, err = r.Impl.Recover(context.TODO(), index, status.Experiment.Records, obj)
			if err != nil {
				// TODO: handle this error
				r.Log.Error(err, "fail to recover chaos")
			}
		}
	}

	if shouldUpdate {
		err := r.Client.Update(context.TODO(), obj)
		if err != nil {
			// TODO: handle this error
			// TODO: retry and update `status.Experiment.Records`
			r.Log.Error(err, "fail to update object", "obj", obj)
		}
	}
	return ctrl.Result{}, nil
}
