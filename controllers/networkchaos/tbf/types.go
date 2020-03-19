package tbf

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/common"
	"github.com/pingcap/chaos-mesh/controllers/reconciler"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	networkTbfActionMsg = "network tbf action duration %s"
)

// TbfSpec defines the interface to convert to a Netem protobuf
type TbfSpec interface {
	ToTbf() (*pb.Tbf, error)
}

type Reconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

// Object implements the reconciler.InnerReconciler.Object
func (r *Reconciler) Object() reconciler.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

func newReconciler(c client.Client, log logr.Logger, req ctrl.Request, recorder record.EventRecorder) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client:        c,
			EventRecorder: recorder,
			Log:           log,
		},
		Client: c,
		Log:    log,
	}
}

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, log logr.Logger, req ctrl.Request, recorder record.EventRecorder) *twophase.Reconciler {
	r := newReconciler(c, log, req, recorder)
	return twophase.NewReconciler(r, r.Client, r.Log)
}

// NewCommonReconciler would create Reconciler for common package
func NewCommonReconciler(c client.Client, log logr.Logger, req ctrl.Request, recorder record.EventRecorder) *common.Reconciler {
	r := newReconciler(c, log, req, recorder)
	return common.NewReconciler(r, r.Client, r.Log)
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, &networkchaos.Spec)

	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}

	err = r.applyAllPods(ctx, pods, networkchaos)
	if err != nil {
		return err
	}

	networkchaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}

	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(networkchaos.Spec.Action),
		}

		if networkchaos.Spec.Duration != nil {
			ps.Message = fmt.Sprintf(networkTbfActionMsg, *networkchaos.Spec.Duration)
		}

		networkchaos.Status.Experiment.Pods = append(networkchaos.Status.Experiment.Pods, ps)
	}
	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

func (r *Reconciler) applyAllPods(ctx context.Context, pods []v1.Pod, networkchaos *v1alpha1.NetworkChaos) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, networkchaos)
		})
	}

	return g.Wait()
}

func (r *Reconciler) applyPod(ctx context.Context, pod *v1.Pod, networkchaos *v1alpha1.NetworkChaos) error {
	r.Log.Info("Try to apply netem on pod", "namespace", pod.Namespace, "name", pod.Name)

	action := string(networkchaos.Spec.Action)
	action = strings.Title(action)

	spec, ok := reflect.Indirect(reflect.ValueOf(networkchaos.Spec)).FieldByName(action).Interface().(TbfSpec)
	if !ok {
		return fmt.Errorf("spec %s is not a NetemSpec", action)
	}

	pbClient, err := utils.NewChaosDaemonClient(ctx, r.Client, pod, os.Getenv("CHAOS_DAEMON_PORT"))
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	tbf, err := spec.ToTbf()
	if err != nil {
		return err
	}

	_, err = pbClient.SetTbf(ctx, &pb.TbfRequest{
		Tbf:         tbf,
		ContainerId: containerID,
	})

	return err
}
