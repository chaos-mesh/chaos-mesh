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

package podnotready

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

const (
	podNotReadyActionMsg = "pod notready duration %s"

	ChaosMeshInjectNotReady v1.PodConditionType = "ChaosMeshInjectNotReady"
)

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, reader client.Reader, log logr.Logger, recorder record.EventRecorder) *twophase.Reconciler {
	r := newReconciler(c, reader, log, recorder)
	return twophase.NewReconciler(r, r.Client, r.Reader, r.Log)
}

// NewCommonReconciler would create Reconciler for common package
func NewCommonReconciler(c client.Client, reader client.Reader, log logr.Logger, recorder record.EventRecorder) *common.Reconciler {
	r := newReconciler(c, reader, log, recorder)
	return common.NewReconciler(r, r.Client, r.Reader, r.Log)
}

func newReconciler(c client.Client, r client.Reader, log logr.Logger, recorder record.EventRecorder) *Reconciler {
	return &Reconciler{
		Client:        c,
		Reader:        r,
		EventRecorder: recorder,
		Log:           log,
	}
}

type Reconciler struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

// Object implements the reconciler.InnerReconciler.Object
func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.PodChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {

	podchaos, ok := chaos.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", chaos)
		return err
	}

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &podchaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}
	err = r.setAllPodsNotReady(ctx, pods, podchaos)
	if err != nil {
		return err
	}

	podchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(podchaos.Spec.Action),
		}
		if podchaos.Spec.Duration != nil {
			ps.Message = fmt.Sprintf(podNotReadyActionMsg, *podchaos.Spec.Duration)
		}
		podchaos.Status.Experiment.PodRecords = append(podchaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(podchaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

func (r *Reconciler) setAllPodsNotReady(ctx context.Context, pods []v1.Pod, podchaos *v1alpha1.PodChaos) error {
	r.Log.Info(" all update condition")
	for index := range pods {
		pod := &pods[index]
		if containsReadinessGate(pod) {
			continue
		}
		r.Log.Info(" into  update condition")
		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		podchaos.Finalizers = utils.InsertFinalizer(podchaos.Finalizers, key)

		clonePod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:            pod.Name,
				Labels:          pod.Labels,
				Annotations:     pod.Annotations,
				GenerateName:    pod.GenerateName,
				Finalizers:      pod.Finalizers,
				OwnerReferences: pod.OwnerReferences,
			},
		}
		if len(pod.Namespace) > 0 {
			clonePod.ObjectMeta.Namespace = pod.Namespace
		}
		clonePod.Spec = *pod.Spec.DeepCopy()
		injectReadinessGate(clonePod)
		if err := r.Delete(ctx, pod); err != nil {
			r.Log.Error(err, "unable to delete pod")
			return err
		}
		err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
			var pod v1.Pod
			if err = r.Client.Get(ctx, types.NamespacedName{
				Namespace: clonePod.Namespace,
				Name:      clonePod.Name,
			}, &pod); err != nil {
				if !k8serror.IsNotFound(err) {
					r.Log.Error(err, "get pod error")
					return false, nil
				}
			} else {
				return false, nil
			}

			if err := r.Create(ctx, clonePod); err != nil {
				r.Log.Error(err, "unable to create clonePod")
				return false, nil
			}

			return true, nil
		})

		if err != nil {
			r.Log.Error(err, "inject pod failed")
			return err
		}

		err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
			var pod v1.Pod
			err = r.Client.Get(ctx, types.NamespacedName{
				Namespace: clonePod.Namespace,
				Name:      clonePod.Name,
			}, &pod)

			if err != nil {
				r.Log.Error(err, "can't get pod")
				return false, nil
			}
			return containsReadinessGate(&pod), nil
		})

		if err != nil {
			r.Log.Error(err, "ensure pods status failed")
			return err
		}

		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(podchaos.Spec.Action),
		}
		if podchaos.Spec.Duration != nil {
			ps.Message = fmt.Sprintf(podNotReadyActionMsg, *podchaos.Spec.Duration)
		}

		podchaos.Status.Experiment.PodRecords = append(podchaos.Status.Experiment.PodRecords, ps)

	}
	return nil
}

// InjectReadinessGate injects ChaosMeshInjectNotReady into pod.spec.readinessGates
func injectReadinessGate(pod *v1.Pod) {
	for _, r := range pod.Spec.ReadinessGates {
		if r.ConditionType == ChaosMeshInjectNotReady {
			return
		}
	}
	pod.Spec.ReadinessGates = append(pod.Spec.ReadinessGates, v1.PodReadinessGate{ConditionType: ChaosMeshInjectNotReady})
}

func removeReadinessGate(pod *v1.Pod) {
	for i, r := range pod.Spec.ReadinessGates {
		if r.ConditionType == ChaosMeshInjectNotReady {
			pod.Spec.ReadinessGates = append(pod.Spec.ReadinessGates[:i], pod.Spec.ReadinessGates[i+1:]...)
			return
		}
	}
	return
}

func ensureSystemCondition(conditions []v1.PodCondition) bool {
	for _, c := range conditions {
		if c.Type == ChaosMeshInjectNotReady {
			return true
		}
	}
	return false
}

// Recover implements the reconciler.InnerReconciler.Recover
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, obj v1alpha1.InnerObject) error {

	podchaos, ok := obj.(*v1alpha1.PodChaos)
	if !ok {
		err := errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", obj)
		return err
	}

	if err := r.setPodReadyAndRecover(ctx, podchaos); err != nil {
		return err
	}

	r.Event(podchaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")
	return nil
}

func (r *Reconciler) setPodReadyAndRecover(ctx context.Context, podchaos *v1alpha1.PodChaos) error {
	var result error

	for _, key := range podchaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		var pod v1.Pod
		err = r.Client.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &pod)

		if err != nil {
			if !k8serror.IsNotFound(err) {
				result = multierror.Append(result, err)
				continue
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			podchaos.Finalizers = utils.RemoveFromFinalizer(podchaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod, podchaos)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		podchaos.Finalizers = utils.RemoveFromFinalizer(podchaos.Finalizers, key)
	}

	if podchaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", podchaos)
		podchaos.Finalizers = podchaos.Finalizers[:0]
		return nil
	}

	return result
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, podchaos *v1alpha1.PodChaos) error {
	r.Log.Info("Recovering", "namespace", pod.Namespace, "name", pod.Name)
	clonePod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pod.Name,
			Labels:          pod.Labels,
			Annotations:     pod.Annotations,
			GenerateName:    pod.GenerateName,
			Finalizers:      pod.Finalizers,
			OwnerReferences: pod.OwnerReferences,
		},
	}
	if len(pod.Namespace) > 0 {
		clonePod.ObjectMeta.Namespace = pod.Namespace
	}
	clonePod.Spec = *pod.Spec.DeepCopy()
	removeReadinessGate(clonePod)

	if err := r.Delete(ctx, pod); err != nil {
		r.Log.Error(err, "unable to delete pod")
		return err
	}
	if err := r.Create(ctx, clonePod); err != nil {
		r.Log.Error(err, "unable to create pod")
		return err
	}
	return nil
}

func containsReadinessGate(pod *v1.Pod) bool {
	for _, r := range pod.Spec.ReadinessGates {
		if r.ConditionType == ChaosMeshInjectNotReady {
			return true
		}
	}
	return false
}
