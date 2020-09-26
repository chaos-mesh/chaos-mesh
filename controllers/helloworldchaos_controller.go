package controllers

import (
	"github.com/go-logr/logr"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HelloWorldChaosReconciler reconciles a HelloWorldChaos object
type HelloWorldChaosReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=chaos-mesh.org,resources=helloworldchaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chaos-mesh.org,resources=helloworldchaos/status,verbs=get;update;patch

func (r *HelloWorldChaosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("reconciler", "helloworldchaos")

	// the main logic of `HelloWorldChaos`, it prints a log `Hello World!` and returns nothing.
	logger.Info("Hello World!")

	return ctrl.Result{}, nil
}

func (r *HelloWorldChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// exports `HelloWorldChaos` object, which represents the yaml schema content the user applies.
	return ctrl.NewControllerManagedBy(mgr).
		For(&chaosmeshv1alpha1.HelloWorldChaos{}).
		Complete(r)
}
