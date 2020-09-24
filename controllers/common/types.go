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
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"

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

const emptyString = ""

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

	if config.WatcherConfig == nil {
		return fmt.Errorf("required WatcherConfig is missing")
	}

	if config.ClusterScoped != config.WatcherConfig.ClusterScoped {
		return fmt.Errorf("K8sConfigMapWatcher config ClusterScoped is not same with controller-manager ClusterScoped. k8s configmap watcher: %t, controller manager: %t", config.WatcherConfig.ClusterScoped, config.ClusterScoped)
	}

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
	chaosSourceTargetSpec := chaos.GetSourceTargetSpec()
	tgt, err := utils.ResolveTargets(ctx, r.Client, r.Reader, chaosSourceTargetSpec, ControllerCfg.ClusterScoped, ControllerCfg.AllowedNamespaces, ControllerCfg.IgnoredNamespaces, ControllerCfg.TargetNamespace)
	if err != nil {
		r.Log.Error(err, "failed to select and filter pods by chaos select spec")
		return ctrl.Result{}, err
	}

	if chaos.IsDeleted() {
		// This chaos was deleted
		r.Log.Info("Removing self")
		if err = r.Recover(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to recover chaos")
			updateFailedMessage(ctx, r, chaos, err.Error())
			return ctrl.Result{Requeue: true}, err
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished
		status.FailedMessage = emptyString
	} else if chaos.IsPaused() {
		if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
			r.Log.Info("Pausing")

			if err = r.Recover(ctx, req, chaos); err != nil {
				r.Log.Error(err, "failed to pause chaos")
				updateFailedMessage(ctx, r, chaos, err.Error())
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
		status.FailedMessage = emptyString
	} else if status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
		if !chaos.IsRenewed(tgt) {
			r.Log.Info("The common chaos is already running", "name", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, nil
		}

		r.Log.Info("Refreshing chaos target")

		if err = r.Recover(ctx, req, chaos); err != nil {
			r.Log.Error(err, "failed to recover chaos")
			updateFailedMessage(ctx, r, chaos, err.Error())
			return ctrl.Result{Requeue: true}, err
		}

		if err = r.Apply(ctx, req, chaos, tgt); err != nil {
			r.Log.Error(err, "failed to apply chaos action")
			updateFailedMessage(ctx, r, chaos, err.Error())

			status.Experiment.Phase = v1alpha1.ExperimentPhaseFailed

			updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.Update(ctx, chaos)
			})
			if updateError != nil {
				r.Log.Error(updateError, "unable to update chaos finalizers")
				updateFailedMessage(ctx, r, chaos, updateError.Error())
			}

			return ctrl.Result{Requeue: true}, err
		}

		r.Log.Info("The common chaos target has been renewed", "name", req.Name, "namespace", req.Namespace)
		return ctrl.Result{}, nil
	} else {
		// Start chaos action
		r.Log.Info("Performing Action")

		if err = r.Apply(ctx, req, chaos, tgt); err != nil {
			r.Log.Error(err, "failed to apply chaos action")
			updateFailedMessage(ctx, r, chaos, err.Error())

			status.Experiment.Phase = v1alpha1.ExperimentPhaseFailed

			updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return r.Update(ctx, chaos)
			})
			if updateError != nil {
				r.Log.Error(updateError, "unable to update chaos finalizers")
				updateFailedMessage(ctx, r, chaos, updateError.Error())
			}

			return ctrl.Result{Requeue: true}, err
		}
		status.Experiment.StartTime = &metav1.Time{
			Time: time.Now(),
		}
		status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning
		status.FailedMessage = emptyString
	}

	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaos status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func updateFailedMessage(
	ctx context.Context,
	r *Reconciler,
	chaos v1alpha1.InnerObject,
	err string,
) {
	status := chaos.GetStatus()
	status.FailedMessage = err
	if err := r.Update(ctx, chaos); err != nil {
		r.Log.Error(err, "unable to update chaos status")
	}
}
