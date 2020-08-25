// Copyright 2020 Chaos Mesh Authors.
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

package stresschaos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

const stressChaosMsg = "stress out pod"

// Reconciler is stresschaos reconciler
type Reconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

// Reconcile reconciles a StressChaos resource
func (r *Reconciler) Reconcile(req ctrl.Request, chaos *v1alpha1.StressChaos) (ctrl.Result, error) {
	r.Log.Info("Reconciling stresschaos")
	scheduler := chaos.GetScheduler()
	duration, err := chaos.GetDuration()
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get stresschaos[%s/%s]'s duration", chaos.Namespace, chaos.Name))
		return ctrl.Result{}, err
	}
	if scheduler == nil && duration == nil {
		return r.commonStressChaos(chaos, req)
	} else if scheduler != nil && duration != nil {
		return r.scheduleStressChaos(chaos, req)
	}

	// This should be ensured by admission webhook in the future
	r.Log.Error(fmt.Errorf("stresschaos[%s/%s] spec invalid", chaos.Namespace, chaos.Name), "scheduler and duration should be omitted or defined at the same time")
	return ctrl.Result{}, fmt.Errorf("invalid scheduler and duration")
}

func (r *Reconciler) commonStressChaos(stresschaos *v1alpha1.StressChaos, req ctrl.Request) (ctrl.Result, error) {
	cr := common.NewReconciler(r, r.Client, r.Log)
	return cr.Reconcile(req)
}

func (r *Reconciler) scheduleStressChaos(stresschaos *v1alpha1.StressChaos, req ctrl.Request) (ctrl.Result, error) {
	sr := twophase.NewReconciler(r, r.Client, r.Log)
	return sr.Reconcile(req)
}

// Apply applies stress-chaos
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	stresschaos, ok := chaos.(*v1alpha1.StressChaos)
	if !ok {
		err := errors.New("chaos is not stresschaos")
		r.Log.Error(err, "chaos is not StressChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, &stresschaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	stresschaos.Status.Instances = make(map[string]v1alpha1.StressInstance, len(pods))
	if err = r.applyAllPods(ctx, pods, stresschaos); err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}

	stresschaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Message:   stressChaosMsg,
		}

		stresschaos.Status.Experiment.PodRecords = append(stresschaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(stresschaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

// Recover means the reconciler recovers the chaos action
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	stresschaos, ok := chaos.(*v1alpha1.StressChaos)
	if !ok {
		err := errors.New("chaos is not StressChaos")
		r.Log.Error(err, "chaos is not StressChaos", "chaos", chaos)
		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, stresschaos); err != nil {
		return err
	}
	r.Event(stresschaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.StressChaos) error {
	var result error

	for _, key := range chaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		var pod v1.Pod
		err = r.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &pod)

		if err != nil {
			if !k8serror.IsNotFound(err) {
				result = multierror.Append(result, err)
				continue
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod, chaos)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
	}

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = chaos.Finalizers[:0]
		return nil
	}

	return result
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.StressChaos) error {
	r.Log.Info("Try to recover pod", "namespace", pod.Namespace, "name", pod.Name)
	daemonClient, err := utils.NewChaosDaemonClient(ctx, r.Client,
		pod, common.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer daemonClient.Close()
	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s/%s can't get the state of container", pod.Namespace, pod.Name)
	}
	instance, ok := chaos.Status.Instances[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)]
	if !ok {
		r.Log.Info("Pod seems already recovered", "pod", pod.UID)
		return nil
	}
	if _, err = daemonClient.CancelStressors(ctx, &pb.CancelStressRequest{
		Instance:  instance.UID,
		StartTime: instance.StartTime.UnixNano() / int64(time.Millisecond),
	}); err != nil {
		return err
	}
	delete(chaos.Status.Instances, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
	return nil
}

// Object would return the instance of chaos
func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.StressChaos{}
}

func (r *Reconciler) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.StressChaos) error {
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

func (r *Reconciler) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.StressChaos) error {
	r.Log.Info("Try to apply stress chaos", "namespace",
		pod.Namespace, "name", pod.Name)
	daemonClient, err := utils.NewChaosDaemonClient(ctx, r.Client,
		pod, common.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer daemonClient.Close()
	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	target := pod.Status.ContainerStatuses[0].ContainerID
	if chaos.Spec.ContainerName != nil &&
		len(strings.TrimSpace(*chaos.Spec.ContainerName)) != 0 {
		target = ""
		for _, container := range pod.Status.ContainerStatuses {
			if container.Name == *chaos.Spec.ContainerName {
				target = container.ContainerID
			}
		}
		if len(target) == 0 {
			return fmt.Errorf("cannot find container with name %s", *chaos.Spec.ContainerName)
		}
	}

	stressors := chaos.Spec.StressngStressors
	if len(stressors) == 0 {
		stressors, err = chaos.Spec.Stressors.Normalize()
		if err != nil {
			return err
		}
	}
	res, err := daemonClient.ExecStressors(ctx, &pb.ExecStressRequest{
		Scope:     pb.ExecStressRequest_CONTAINER,
		Target:    target,
		Stressors: stressors,
	})
	if err != nil {
		return err
	}
	chaos.Status.Instances[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)] = v1alpha1.StressInstance{
		UID: res.Instance,
		StartTime: &metav1.Time{
			Time: time.Unix(res.StartTime/1000, (res.StartTime%1000)*int64(time.Millisecond)),
		},
	}
	return nil
}
