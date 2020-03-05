package controllers

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"

	chaosmeshv1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MemoryChaosReconciler reconciles a HelloWorldChaos object
type MemoryChaosReconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

// +kubebuilder:rbac:groups=pingcap.com,resources=memorychaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pingcap.com,resources=memorychaos/status,verbs=get;update;patch

func (r *MemoryChaosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("reconciler", "helloworldchaos")

	//  the main logic of `HelloWorldChaos`, it prints a log `Hello World!` and returns nothing.
	logger.Info("Hello World!")

	return ctrl.Result{}, nil
}

func (r *MemoryChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Info("Setting up MemoryChaos Controller")
	return ctrl.NewControllerManagedBy(mgr).
		//exports `MemoryChaos` object, which represents the yaml schema content the user applies.
		For(&chaosmeshv1alpha1.MemoryChaos{}).
		Complete(r)
}
