// Copyright 2019 PingCAP, Inc.
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

package fs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/common"
	"github.com/pingcap/chaos-mesh/controllers/reconciler"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	fscli "github.com/pingcap/chaos-mesh/pkg/chaosfs/client"
	"github.com/pingcap/chaos-mesh/pkg/utils"
)

type Reconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

func newReconciler(c client.Client, log logr.Logger, req ctrl.Request) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client: c,
			Log:    log,
		},
		Client: c,
		Log:    log,
	}
}

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, log logr.Logger, req ctrl.Request) *twophase.Reconciler {
	r := newReconciler(c, log, req)
	return twophase.NewReconciler(r, r.Client, r.Log)
}

// NewCommonReconciler would create Reconciler for common package
func NewCommonReconciler(c client.Client, log logr.Logger, req ctrl.Request) *common.Reconciler {
	r := newReconciler(c, log, req)
	return common.NewReconciler(r, r.Client, r.Log)
}

// Object implements the reconciler.InnerReconciler.Object
func (r *Reconciler) Object() reconciler.InnerObject {
	return &v1alpha1.IoChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	iochaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		err := errors.New("chaos is not IoChaos")
		r.Log.Error(err, "chaos is not IoChaos", "chaos", chaos)
		return err
	}

	r.Recorder.Event(chaos, v1.EventTypeNormal, utils.EventChaosStarted, "")
	pods, err := utils.SelectAndGeneratePods(ctx, r.Client, &iochaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		r.Recorder.Eventf(chaos, v1.EventTypeWarning, utils.EventChaosFailed,
			"failed to select and generate pods:%s", err)
		return err
	}

	if err := r.injectAllPods(ctx, pods, iochaos); err != nil {
		return err
	}

	iochaos.Status.Experiment.StartTime = &metav1.Time{
		Time: time.Now(),
	}

	iochaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}
	iochaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseRunning

	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(iochaos.Spec.Action),
		}
		if iochaos.Spec.Duration != nil {
			ps.Message = genMessage(iochaos)
		}

		iochaos.Status.Experiment.Pods = append(iochaos.Status.Experiment.Pods, ps)
	}

	return nil
}

// Recover implements the reconciler.InnerReconciler.Recover
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	iochaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		err := errors.New("chaos is not IoChaos")
		r.Log.Error(err, "chaos is not IoChaos", "chaos", chaos)
		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, iochaos); err != nil {
		return err
	}

	iochaos.Status.Experiment.EndTime = &metav1.Time{
		Time: time.Now(),
	}

	iochaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, iochaos *v1alpha1.IoChaos) error {
	if len(iochaos.Finalizers) == 0 {
		return nil
	}

	for _, key := range iochaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return err
		}

		var pod v1.Pod
		err = r.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &pod)

		if err != nil {
			if !k8serror.IsNotFound(err) {
				return err
			}

			r.Log.Info("Pod not found", "namespace", ns, "names", name)
			iochaos.Finalizers = utils.RemoveFromFinalizer(iochaos.Finalizers, key)
			continue
		}

		if err := r.recoverPod(ctx, &pod, iochaos); err != nil {
			return err
		}

		iochaos.Finalizers = utils.RemoveFromFinalizer(iochaos.Finalizers, key)
	}

	return nil
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, iochaos *v1alpha1.IoChaos) error {
	r.Log.Info("Recovering", "namespace", pod.Namespace, "name", pod.Name)

	cctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	err := wait.PollUntil(2*time.Second, func() (bool, error) {
		if err := r.recoverInjectAction(ctx, pod, iochaos); err != nil {
			if utils.IsCaredNetError(err) {
				r.Log.Info("Recover I/O chaos action, network is not ok, retrying...",
					"namespace", pod.Namespace, "name", pod.Name)
				return false, nil
			}

			return false, err
		}

		r.Log.Info("Recover I/O chaos action successfully")

		return true, nil
	}, cctx.Done())

	if err != nil {
		r.Log.Error(err, "failed to recover I/O chaos action",
			"namespace", pod.Namespace, "name", pod.Name)
		return err
	}

	return nil
}

func (r *Reconciler) injectAllPods(ctx context.Context, pods []v1.Pod, iochaos *v1alpha1.IoChaos) error {
	g := errgroup.Group{}

	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}

		iochaos.Finalizers = utils.InsertFinalizer(iochaos.Finalizers, key)

		g.Go(func() error {
			return r.injectPod(ctx, pod, iochaos)
		})
	}

	return g.Wait()
}

func (r *Reconciler) injectPod(ctx context.Context, pod *v1.Pod, iochaos *v1alpha1.IoChaos) error {
	r.Log.Info("Inject I/O chaos action", "namespace", pod.Namespace, "name", pod.Name)

	cctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	err := wait.PollUntil(2*time.Second, func() (bool, error) {
		if err := r.injectAction(ctx, pod, iochaos); err != nil {
			if utils.IsCaredNetError(err) {
				r.Log.Info("Inject I/O chaos action, network is not ok, retrying...",
					"namespace", pod.Namespace, "name", pod.Name)
				return false, nil
			}

			return false, err
		}

		r.Log.Info("Inject I/O chaos action successfully")

		return true, nil
	}, cctx.Done())

	if err != nil {
		r.Log.Error(err, "failed to inject I/O chaos action",
			"namespace", pod.Namespace, "name", pod.Name)
		return err
	}

	return nil
}

func (r *Reconciler) injectAction(ctx context.Context, pod *v1.Pod, iochaos *v1alpha1.IoChaos) error {
	// TODO: move to api repo
	addr := iochaos.Spec.Addr
	if addr == "" {
		addr = v1alpha1.DefaultChaosfsAddr
	}

	addr = fmt.Sprintf("%s%s", pod.Status.PodIP, addr)

	cli, err := fscli.NewClient(addr)
	if err != nil {
		return err
	}

	req, err := genChaosfsRequest(iochaos)
	if err != nil {
		return err
	}

	if len(req.Methods) > 0 {
		_, err = cli.SetFault(ctx, req)
		return err
	}

	// inject fault to all methods if the the methods is empty.
	_, err = cli.SetFaultAll(ctx, req)
	return err
}

func (r *Reconciler) recoverInjectAction(ctx context.Context, pod *v1.Pod, iochaos *v1alpha1.IoChaos) error {
	addr := iochaos.Spec.Addr
	if addr == "" {
		addr = v1alpha1.DefaultChaosfsAddr
	}

	addr = fmt.Sprintf("%s%s", pod.Status.PodIP, addr)

	cli, err := fscli.NewClient(addr)
	if err != nil {
		return err
	}

	_, err = cli.RecoverAll(ctx, &empty.Empty{})
	return err
}
