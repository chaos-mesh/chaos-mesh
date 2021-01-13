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
	"strings"

	"github.com/hashicorp/go-multierror"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/iochaos/podiochaosmanager"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/finalizer"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

type endpoint struct {
	ctx.Context
}

func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.IoChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	iochaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		err := errors.New("chaos is not IOChaos")
		r.Log.Error(err, "chaos is not IOChaos", "chaos", chaos)
		return err
	}

	source := iochaos.Namespace + "/" + iochaos.Name
	m := podiochaosmanager.New(source, r.Log, r.Client)

	pods, err := selector.SelectAndFilterPods(ctx, r.Client, r.Reader, &iochaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
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

		if iochaos.Spec.ContainerName != nil &&
			len(strings.TrimSpace(*iochaos.Spec.ContainerName)) != 0 {
			t.SetContainer(*iochaos.Spec.ContainerName)
		}

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
		iochaos.Finalizers = finalizer.InsertFinalizer(iochaos.Finalizers, key)
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
	r.Event(iochaos, v1.EventTypeNormal, events.ChaosInjected, "")

	return nil
}

// Recover means the reconciler recovers the chaos action
func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	iochaos, ok := chaos.(*v1alpha1.IoChaos)
	if !ok {
		err := errors.New("chaos is not IoChaos")
		r.Log.Error(err, "chaos is not IoChaos", "chaos", chaos)
		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, iochaos); err != nil {
		return err
	}
	r.Event(iochaos, v1.EventTypeNormal, events.ChaosRecovered, "")
	return nil
}

func (r *endpoint) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.IoChaos) error {
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

		chaos.Finalizers = finalizer.RemoveFromFinalizer(chaos.Finalizers, response.Key.String())
	}
	r.Log.Info("After recovering", "finalizers", chaos.Finalizers)

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = make([]string, 0)
		return nil
	}

	return result
}

func init() {
	router.Register("iochaos", &v1alpha1.IoChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
