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

package jvmchaos

import (
	"context"
	"errors"
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/finalizer"
	"github.com/chaos-mesh/chaos-mesh/pkg/jvm"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"

	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

const sandboxPort = 10086

type endpoint struct {
	ctx.Context
}

// Apply applies jvm-chaos
func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	jvmchaos, ok := chaos.(*v1alpha1.JVMChaos)
	if !ok {
		err := errors.New("chaos is not JVMChaos")
		r.Log.Error(err, "chaos is not JVMChaos", "chaos", chaos)
		return err
	}

	pods, err := selector.SelectAndFilterPods(ctx, r.Client, r.Reader, &jvmchaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	if err = r.applyAllPods(ctx, pods, jvmchaos); err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}

	jvmchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
		}

		jvmchaos.Status.Experiment.PodRecords = append(jvmchaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(jvmchaos, v1.EventTypeNormal, events.ChaosInjected, "")
	return nil
}

func (r *endpoint) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.JVMChaos) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			r.Log.Error(err, "get meta namespace key")
			return err
		}
		chaos.Finalizers = finalizer.InsertFinalizer(chaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, chaos)
		})
	}
	err := g.Wait()
	if err != nil {
		r.Log.Error(err, "g.Wait")
		return err
	}
	return nil
}

func (r *endpoint) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.JVMChaos) error {
	r.Log.Info("Try to apply jvm chaos", "namespace",
		pod.Namespace, "name", pod.Name)

	// TODO: Custom port may be required
	err := jvm.ActiveSandbox(pod.Status.PodIP, sandboxPort)
	if err != nil {
		return err
	}

	r.Log.Info("active sandbox", "pod", pod.Name)

	suid := genSUID(pod, chaos)
	jsonBytes, err := jvm.ToSandboxAction(suid, chaos)

	if err != nil {
		return err
	}
	// TODO: Custom port may be required
	err = jvm.InjectChaos(pod.Status.PodIP, sandboxPort, jsonBytes)
	if err != nil {
		return err
	}
	r.Log.Info("Inject JVM Chaos", "pod", pod.Name, "action", chaos.Spec.Action)
	return nil
}

func genSUID(pod *v1.Pod, chaos *v1alpha1.JVMChaos) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s",
		pod.Name,
		chaos.Spec.Action,
		chaos.Spec.Target,
		chaos.Name,
		chaos.Namespace)
}

// Recover means the reconciler recovers the chaos action
func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	jvmchaos, ok := chaos.(*v1alpha1.JVMChaos)
	if !ok {
		err := errors.New("chaos is not JVMChaos")
		r.Log.Error(err, "chaos is not JVMChaos", "chaos", chaos)
		return err
	}
	if err := r.cleanFinalizersAndRecover(ctx, jvmchaos); err != nil {
		return err
	}

	r.Event(jvmchaos, v1.EventTypeNormal, events.ChaosRecovered, "")

	return nil
}

func (r *endpoint) cleanFinalizersAndRecover(ctx context.Context, chaos *v1alpha1.JVMChaos) error {
	var result error

	for _, key := range chaos.Finalizers {
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
			chaos.Finalizers = finalizer.RemoveFromFinalizer(chaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod, chaos)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		chaos.Finalizers = finalizer.RemoveFromFinalizer(chaos.Finalizers, key)
	}

	if chaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", chaos)
		chaos.Finalizers = chaos.Finalizers[:0]
		return nil
	}

	return result
}

func (r *endpoint) recoverPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.JVMChaos) error {
	r.Log.Info("Try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	suid := genSUID(pod, chaos)
	jsonBytes, err := jvm.ToSandboxAction(suid, chaos)
	if err != nil {
		return err
	}

	// TODO: Custom port may be required
	err = jvm.RecoverChaos(pod.Status.PodIP, sandboxPort, jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

// Object would return the instance of chaos
func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.JVMChaos{}
}

func init() {
	router.Register("jvmchaos", &v1alpha1.JVMChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
