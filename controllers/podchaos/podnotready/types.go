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
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
		if !containsReadinessGate(pod) {
			continue
		}
		r.Log.Info(" into  update condition")
		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		podchaos.Finalizers = utils.InsertFinalizer(podchaos.Finalizers, key)

		var newCondition = v1.PodCondition{
			Type:               ChaosMeshInjectNotReady,
			LastTransitionTime: metav1.Time{},
			Status:             v1.ConditionTrue,
			Reason:             "StartChaosMeshInjectNotReady",
		}
		setPodCondition(pod, newCondition)
		r.Log.Info(" update condition")
		if err := r.Status().Update(ctx, pod); err != nil {
			r.Log.Error(err, "unable to update condition")
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

func setPodCondition(pod *v1.Pod, condition v1.PodCondition) {
	for i, c := range pod.Status.Conditions {
		if c.Type == condition.Type {
			if c.Status != condition.Status {
				pod.Status.Conditions[i] = condition
			}
			return
		}
	}
	pod.Status.Conditions = append(pod.Status.Conditions, condition)
}

func resetPodCondition(pod *v1.Pod) {
	for i, c := range pod.Status.Conditions {
		if c.Type == ChaosMeshInjectNotReady {
			pod.Status.Conditions = append(pod.Status.Conditions[:i], pod.Status.Conditions[i+1:]...)
			return
		}
	}
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
	resetPodCondition(pod)
	if err := r.Update(ctx, pod); err != nil {
		r.Log.Error(err, "unable to update pod")
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
