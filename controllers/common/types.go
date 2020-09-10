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
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/reconciler"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	// AnnotationCleanFinalizer key
	AnnotationCleanFinalizer = `chaos-mesh.chaos-mesh.org/cleanFinalizer`
	// AnnotationCleanFinalizerForced value
	AnnotationCleanFinalizerForced = `forced`
)

var log = ctrl.Log.WithName("controller")

//ControllerCfg is a global variable to keep the configuration for Chaos Controller
var ControllerCfg *config.ChaosControllerConfig

func init() {
	conf, err := config.EnvironChaosController()
	if err != nil {
		ctrl.SetLogger(zap.Logger(true))
		log.Error(err, "Chaos Controller: invalid environment configuration")
		os.Exit(1)
	}

	err = validate(&conf)
	if err != nil {
		ctrl.SetLogger(zap.Logger(true))
		log.Error(err, "Chaos Controller: invalid configuration")
		os.Exit(1)
	}

	ControllerCfg = &conf
}

func validate(config *config.ChaosControllerConfig) error {
	if !config.ClusterScoped {
		if strings.TrimSpace(config.TargetNamespace) == "" {
			return fmt.Errorf("no target namespace specified with namespace scoped mode")
		}
		if !isAllowedNamespaces(config.TargetNamespace, config.AllowedNamespaces, config.IgnoredNamespaces) {
			return fmt.Errorf("target namespace %s is not allowed with filter, please check config AllowedNamespaces and IgnoredNamespaces", config.TargetNamespace)
		}

		if config.TargetNamespace != config.WatcherConfig.TargetNamespace {
			return fmt.Errorf("K8sConfigMapWatcher config TargertNamespace is not same with controller-manager TargetNamespace. k8s configmap watcher: %s, controller manager: %s", config.WatcherConfig.TargetNamespace, config.TargetNamespace)
		}
	}

	return nil
}

// FIXME: duplicated with utils.IsAllowedNamespaces, it should considered with some dependency problems.
func isAllowedNamespaces(namespace, allowedNamespace, ignoredNamespace string) bool {
	if allowedNamespace != "" {
		matched, err := regexp.MatchString(allowedNamespace, namespace)
		if err != nil {
			return false
		}
		return matched
	}

	if ignoredNamespace != "" {
		matched, err := regexp.MatchString(ignoredNamespace, namespace)
		if err != nil {
			return false
		}
		return !matched
	}

	return true
}

// Reconciler for common chaos
type Reconciler struct {
	reconciler.InnerReconciler
	client.Client
	client.Reader
	Log logr.Logger
}

// NewReconciler would create Reconciler for common chaos
func NewReconciler(reconcile reconciler.InnerReconciler, c client.Client, r client.Reader, log logr.Logger) *Reconciler {
	return &Reconciler{
		InnerReconciler: reconcile,
		Client:          c,
		Reader:          r,
		Log:             log,
	}
}

// Reconcile the common chaos
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var err error

	r.Log.Info("Reconciling a common chaos", "name", req.Name, "namespace", req.Namespace)
	ctx := context.Background()

	chaos := r.Object()
	if err = r.Client.Get(ctx, req.NamespacedName, chaos); err != nil {
		r.Log.Error(err, "unable to get chaos")
		return ctrl.Result{}, err
	}

	status := chaos.GetStatus()

	if chaos.IsDeleted() {
		// This chaos was deleted
		r.Log.Info("Removing self")
		if err = r.Recover(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to recover chaos")
			return ctrl.Result{Requeue: true}, err
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished
	} else if chaos.IsPaused() {
		if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
			r.Log.Info("Pausing")

			if err = r.Recover(ctx, req, chaos); err != nil {
				r.Log.Error(err, "failed to pause chaos")
				return ctrl.Result{Requeue: true}, err
			}
			now := time.Now()
			status.Experiment.EndTime = &metav1.Time{
				Time: now,
			}
			if status.Experiment.StartTime != nil {
				status.Experiment.Duration = now.Sub(status.Experiment.StartTime.Time).String()
			}
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhasePaused
	} else if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
		r.Log.Info("The common chaos is already running", "name", req.Name, "namespace", req.Namespace)
		return ctrl.Result{}, nil
	} else {
		// Start chaos action
		r.Log.Info("Performing Action")

		if err = r.Apply(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to apply chaos action")

			status.Experiment.Phase = v1alpha1.ExperimentPhaseFailed

			updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.Update(ctx, chaos)
			})
			if updateError != nil {
				r.Log.Error(updateError, "unable to update chaos finalizers")
			}

			return ctrl.Result{Requeue: true}, err
		}
		status.Experiment.StartTime = &metav1.Time{
			Time: time.Now(),
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning
	}

	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaos status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
