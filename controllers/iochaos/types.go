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

package iochaos

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/iochaos/podiochaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

type Reconciler struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.IoChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	iochaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		err := errors.New("chaos is not IoChaos")
		r.Log.Error(err, "chaos is not IoChaos", "chaos", chaos)
		return err
	}

	source := iochaos.Namespace + "/" + iochaos.Name
	m := podiochaosmanager.New(source, r.Log, r.Client)

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &iochaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}

	keyPodMap := make(map[types.NamespacedName]v1.Pod)
	for _, pod := range pods {
		keyPodMap[types.NamespacedName{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		}] = pod
	}

	r.Log.Info("applying iochaos", "iochaos", iochaos)

	for _, pod := range pods {
		t := m.WithInit(types.NamespacedName{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})

		// TODO: support chaos on multiple volume
		t.SetVolumePath(iochaos.Spec.VolumePath)
		t.Append(v1alpha1.IoChaosAction{
			Type: iochaos.Spec.Action,
			Filter: v1alpha1.Filter{
				Path:    iochaos.Spec.Path,
				Percent: iochaos.Spec.Percent,
				Methods: iochaos.Spec.Methods,
			},
			Faults: []v1alpha1.IoFault{
				{
					Errno:  iochaos.Spec.Errno,
					Weight: 1,
				},
			},
			Latency:          iochaos.Spec.Delay,
			AttrOverrideSpec: iochaos.Spec.Attr,
			Source:           m.Source,
		})

		key, err := cache.MetaNamespaceKeyFunc(&pod)
		if err != nil {
			return err
		}
		iochaos.Finalizers = utils.InsertFinalizer(iochaos.Finalizers, key)
	}
	r.Log.Info("commiting updates of podiochaos")
	responses := m.Commit(ctx)
	iochaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, keyErrorTuple := range responses {
		key := keyErrorTuple.Key
		err := keyErrorTuple.Err
		if err != nil {
			if err != podiochaosmanager.ErrPodNotFound && err != podiochaosmanager.ErrPodNotRunning {
				r.Log.Error(err, "fail to commit")
			} else {
				r.Log.Info("pod is not found or not running", "key", key)
			}
			return err
		}

		pod := keyPodMap[keyErrorTuple.Key]

		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(iochaos.Spec.Action),
		}

		iochaos.Status.Experiment.PodRecords = append(iochaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(iochaos, v1.EventTypeNormal, utils.EventChaosInjected, "")

	return nil
}

// Recover implements the reconciler.InnerReconciler.Recover
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	iochaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		err := errors.New("chaos is not IoChaos")
		r.Log.Error(err, "chaos is not IoChaos", "chaos", chaos)
		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, iochaos); err != nil {
		return err
	}
	r.Event(iochaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")
	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.IoChaos) error {
	var result error

	source := chaos.Namespace + "/" + chaos.Name

	m := podiochaosmanager.New(source, r.Log, r.Client)
	for _, key := range chaos.Finalizers {

		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		_ = m.WithInit(types.NamespacedName{
			Namespace: ns,
			Name:      name,
		})
	}
	responses := m.Commit(ctx)
	for _, response := range responses {
		key := response.Key
		err := response.Err
		// if pod not found or not running, directly return and giveup recover.
		if err != nil {
			if err != podiochaosmanager.ErrPodNotFound && err != podiochaosmanager.ErrPodNotRunning {
				r.Log.Error(err, "fail to commit", "key", key)

				result = multierror.Append(result, err)
				continue
			}

			r.Log.Info("pod is not found or not running", "key", key)
		}

		chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, response.Key.String())
	}
	r.Log.Info("After recovering", "finalizers", chaos.Finalizers)

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = make([]string, 0)
		return nil
	}

	return result
}

func (r *Reconciler) invalidActionResponse(iochaos *v1alpha1.IoChaos) (ctrl.Result, error) {
	r.Log.Error(nil, "unknown file system I/O layer", "action", iochaos.Spec.Action)
	return ctrl.Result{}, fmt.Errorf("unknown file system I/O layer")
}

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, r client.Reader, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *twophase.Reconciler {
	reconciler := newReconciler(c, r, log, req, recorder)
	return twophase.NewReconciler(reconciler, reconciler.Client, reconciler.Reader, reconciler.Log)
}

// NewCommonReconciler would create Reconciler for common package
func NewCommonReconciler(c client.Client, r client.Reader, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *common.Reconciler {
	reconciler := newReconciler(c, r, log, req, recorder)
	return common.NewReconciler(reconciler, reconciler.Client, reconciler.Reader, reconciler.Log)
}

func newReconciler(c client.Client, r client.Reader, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client:        c,
			Reader:        r,
			EventRecorder: recorder,
			Log:           log,
		},
		Client: c,
		Log:    log,
	}
}
