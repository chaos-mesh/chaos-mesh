/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"

	chaosoperatorv1alpha1 "github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/controllers/podchaos"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PodChaosReconciler reconciles a PodChaos object
type PodChaosReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=pingcap.com,resources=podchaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pingcap.com,resources=podchaos/status,verbs=get;update;patch

func (r *PodChaosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	logger := r.Log.WithValues("reconciler", "podchaos")

	reconciler := podchaos.Reconciler{
		Client: r.Client,
		Log:    logger,
	}

	return reconciler.Reconcile(req)
}

func (r *PodChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&chaosoperatorv1alpha1.PodChaos{}).
		Complete(r)
	if err != nil {
		return err
	}

	return r.SetupWebhookWithManager(mgr)
}

func (r *PodChaosReconciler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return nil
	// TODO: setup Webhook
	//return ctrl.NewWebhookManagedBy(mgr).
	//	For(&chaosoperatorv1alpha1.PodChaos{}).
	//	Complete()
}
