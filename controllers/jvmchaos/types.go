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
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/jvm"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"

	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

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

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &jvmchaos.Spec)
	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	// TODO: applyAllPods
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
	r.Event(jvmchaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
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
		chaos.Finalizers = utils.InsertFinalizer(chaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod)
		})
	}
	err := g.Wait()
	if err != nil {
		r.Log.Error(err, "g.Wait")
		return err
	}
	return nil
}

func (r *endpoint) applyPod(ctx context.Context, pod *v1.Pod) error {
	r.Log.Info("Try to apply jvm chaos", "namespace",
		pod.Namespace, "name", pod.Name)
	// get pod id,and send http pod.Status.PodIP;
	// TODO: Custom ports may be required
	err := jvm.ActiveSandbox(pod.Status.PodIP, 10086)
	if err != nil {
		return err
	}

	r.Log.Info("active sandbox", "pod", pod.Name)

	rand.Seed(time.Now().Unix())
	//suid := fmt.Sprintf("%s-%d", pod.Name , rand.Int())
	suid := "podname-randid"
	delay := &Delay{
		Action:      "delay",
		Target:      "servlet",
		SUID:        suid,
		Time:        "10000",
		RequestPath: "/",
	}
	jsonBytes, err := json.Marshal(delay)
	err = jvm.InjectChaos(pod.Status.PodIP, 10086, jsonBytes)
	if err != nil {
		return err
	}
	r.Log.Info("delay servlet", "pod", pod.Name)
	return nil
}

type Delay struct {
	Action      string `json:"action"`
	Target      string `json:"target"`
	SUID        string `json:"suid"`
	Time        string `json:"time"`
	RequestPath string `json:"requestpath"`
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

	r.Event(jvmchaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")

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
			chaos.Finalizers = utils.RemoveFromFinalizer(chaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod)
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

func (r *endpoint) recoverPod(ctx context.Context, pod *v1.Pod) error {
	r.Log.Info("Try to recover pod", "namespace", pod.Namespace, "name", pod.Name)
	// TODO: Custom ports may be required
	suid := "podname-randid"
	delay := &Delay{
		Action:      "delay",
		Target:      "servlet",
		SUID:        suid,
		Time:        "10000",
		RequestPath: "/",
	}
	jsonBytes, err := json.Marshal(delay)
	err = jvm.RecoverChaos(pod.Status.PodIP, 10086, jsonBytes)
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
