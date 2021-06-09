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
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/pkg/jvm"
)

const sandboxPort = 10086

type Impl struct {
	client.Client
	Log logr.Logger
}

// Apply applies jvm-chaos
func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	jvmchaos := obj.(*v1alpha1.JVMChaos)

	var pod v1.Pod
	err := impl.Client.Get(ctx, controller.ParseNamespacedName(records[index].Id), &pod)
	if err != nil {
		// TODO: handle this error
		return v1alpha1.NotInjected, err
	}

	impl.Log.Info("Try to apply jvm chaos", "namespace",
		pod.Namespace, "name", pod.Name)

	// TODO: Custom port may be required
	err = jvm.ActiveSandbox(pod.Status.PodIP, sandboxPort)
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	impl.Log.Info("active sandbox", "pod", pod.Name)

	suid := genSUID(&pod, jvmchaos)
	jsonBytes, err := jvm.ToSandboxAction(suid, jvmchaos)

	if err != nil {
		return v1alpha1.NotInjected, err
	}
	// TODO: Custom port may be required
	err = jvm.InjectChaos(pod.Status.PodIP, sandboxPort, jsonBytes)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	impl.Log.Info("Inject JVM Chaos", "pod", pod.Name, "action", jvmchaos.Spec.Action)

	return v1alpha1.Injected, nil
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
func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	jvmchaos := obj.(*v1alpha1.JVMChaos)

	var pod v1.Pod
	err := impl.Client.Get(ctx, controller.ParseNamespacedName(records[index].Id), &pod)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return v1alpha1.Injected, err
		}

		impl.Log.Info("Target pod has been deleted", "namespace", pod.Namespace, "name", pod.Name)
		return v1alpha1.NotInjected, nil

	}

	impl.Log.Info("Try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	suid := genSUID(&pod, jvmchaos)
	jsonBytes, err := jvm.ToSandboxAction(suid, jvmchaos)
	if err != nil {
		return v1alpha1.Injected, err
	}

	// TODO: Custom port may be required
	err = jvm.RecoverChaos(pod.Status.PodIP, sandboxPort, jsonBytes)

	if err != nil {
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil
}

// Object would return the instance of chaos

func NewImpl(c client.Client, log logr.Logger) *common.ChaosImplPair {
	return &common.ChaosImplPair{
		Name:   "jvmchaos",
		Object: &v1alpha1.JVMChaos{},
		Impl: &Impl{
			Client: c,
			Log:    log.WithName("jvmchaos"),
		},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
