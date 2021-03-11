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

type StatefulObjectWithSelector interface {
	v1alpha1.StatefulObject

	GetSelectorSpecs() map[string]interface{}
}

// Reconciler for common chaos
type Reconciler struct {
	// Object is used to mark the target type of this Reconciler
	Object StatefulObjectWithSelector

	// Client is used to operate on the Kubernetes cluster
	client.Client
	client.Reader

	Log logr.Logger
}

// Reconcile the common chaos
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	obj := r.Object.DeepCopyObject().(StatefulObjectWithSelector)

	if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("chaos not found")
		} else {
			r.Log.Error(err, "unable to get chaos")
		}
		return ctrl.Result{}, nil
	}

	status := obj.GetStatus()
	if status.Experiment.Records == nil {
		selectedTargets := make(map[string][]interface{})
		// TODO: get selectors from obj
		for name, sel := range obj.GetSelectorSpecs() {
			selector := selector.New(selector.SelectorParams{
				PodSelector: pod.New(r.Client, r.Reader, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces),
				ContainerSelector: container.New(r.Client, r.Reader, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces),
			})
			targets, err := selector.Select(context.TODO(), sel)
			if err != nil {
				r.Log.Error(err, "fail to select")
			}

			selectedTargets[name] = targets
		}

		for name, targets := range selectedTargets {
			for _, tgt := range targets {
				Apply(name, tgt, selectedTargets, obj)
			}
		}
	}
	return ctrl.Result{}, nil
}
