

package controllers

import (
	"github.com/go-logr/logr"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HttpFaultChaosReconciler reconciles aHttpFaultChaos object
type HttpFaultChaosReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=chaos-mesh.org,resources=httpfaultchaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chaos-mesh.org,resources=httpfaultchaos/status,verbs=get;update;patch

func (r *HttpFaultChaosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("reconciler", "httpfaultchaos")

	//  the main logic of `HttpFaultChaos`, it prints a log `Hello World!` and returns nothing.
	logger.Info("Hello World!")

	return ctrl.Result{}, nil
}

func (r *HttpFaultChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		//exports `HttpFaultChaos` object, which represents the yaml schema content the user applies.
		For(&chaosmeshv1alpha1.HttpFaultChaos{}).
		Complete(r)
}
