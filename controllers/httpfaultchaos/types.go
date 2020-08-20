package httpfaultchaos

import (
	"context"
	"errors"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	HttpFaultChaos, ok := chaos.(*v1alpha1.HttpFaultChaos)
	if !ok {
		err := errors.New("chaos is not HttpFaultChaos")
		r.Log.Error(err, "chaos is not HttpFaultChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, &HttpFaultChaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}
	if err = r.applyAllPods(ctx, pods, HttpFaultChaos); err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}
	return nil
}


func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	HttpFaultChaos, ok := chaos.(*v1alpha1.KernelChaos)
	if !ok {
		err := errors.New("chaos is not KernelChaos")
		r.Log.Error(err, "chaos is not KernelChaos", "chaos", chaos)
		return err
	}
	r.Event(HttpFaultChaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")
	return nil
}

func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.HttpFaultChaos{}
}

func (r *Reconciler) Reconcile(req ctrl.Request, chaos *v1alpha1.HttpFaultChaos) (ctrl.Result, error) {
	r.Log.Info("Reconciling HttpFaultChaos")
	duration, err := chaos.GetDuration()
	if err != nil {
		msg := fmt.Sprintf("unable to get iochaos[%s/%s]'s duration",
			req.Namespace, req.Name)
		r.Log.Error(err, msg)
		return ctrl.Result{}, err
	}

	if duration != 0 {
		return r.commonHttpFaultChaos(chaos, req)
	}
	err = fmt.Errorf("HttpFaultChaos[%s/%s] spec invalid", req.Namespace, req.Name)
	r.Log.Error(err, "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, err
}

func (r *Reconciler) commonHttpFaultChaos(HttpFaultChaos *v1alpha1.HttpFaultChaos, req ctrl.Request) (ctrl.Result, error) {
	cr := common.NewReconciler(r, r.Client, r.Log)
	return cr.Reconcile(req)
}

func (r *Reconciler) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.HttpFaultChaos) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		chaos.Finalizers = utils.InsertFinalizer(chaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, chaos)
		})
	}

	return g.Wait()
}

func (r *Reconciler) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.HttpFaultChaos) error {
	r.Log.Info("Try to inject kernel on pod", "namespace", pod.Namespace, "name", pod.Name)
	_, err := http.Get("http://" + pod.Status.PodIP + ":15000/delay?" + "open=true&duration=" + *chaos.Spec.Duration)
	return err
}
