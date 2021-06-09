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

package podiochaos

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

// Reconciler applys podioworkchaos
type Reconciler struct {
	client.Client
	Recorder record.EventRecorder

	Log                      logr.Logger
	ChaosDaemonClientBuilder *chaosdaemon.ChaosDaemonClientBuilder
}

func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.TODO()

	obj := &v1alpha1.PodIOChaos{}

	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("chaos not found")
		} else {
			// TODO: handle this error
			r.Log.Error(err, "unable to get chaos")
		}
		return ctrl.Result{}, nil
	}

	if obj.ObjectMeta.Generation <= obj.Status.ObservedGeneration && obj.Status.FailedMessage == "" {
		r.Log.Info("the target pod has been up to date", "pod", obj.Namespace+"/"+obj.Name)
		return ctrl.Result{}, nil
	}

	r.Log.Info("updating io chaos", "pod", obj.Namespace+"/"+obj.Name, "spec", obj.Spec)

	pod := &v1.Pod{}

	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}, pod)
	if err != nil {
		r.Log.Error(err, "fail to find pod")
		return ctrl.Result{}, nil
	}

	failedMessage := ""
	observedGeneration := obj.ObjectMeta.Generation
	pid := obj.Status.Pid
	startTime := obj.Status.StartTime
	defer func() {
		if err != nil {
			failedMessage = err.Error()
		}

		updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			obj := &v1alpha1.PodIOChaos{}

			if err := r.Client.Get(context.TODO(), req.NamespacedName, obj); err != nil {
				r.Log.Error(err, "unable to get chaos")
				return err
			}

			obj.Status.FailedMessage = failedMessage
			obj.Status.ObservedGeneration = observedGeneration
			obj.Status.Pid = pid
			obj.Status.StartTime = startTime

			return r.Client.Status().Update(context.TODO(), obj)
		})

		if updateError != nil {
			r.Log.Error(updateError, "fail to update")
			r.Recorder.Eventf(obj, "Normal", "Failed", "Failed to update status: %s", updateError.Error())
		}
	}()

	pbClient, err := r.ChaosDaemonClientBuilder.Build(ctx, pod)
	if err != nil {
		r.Recorder.Event(obj, "Warning", "Failed", err.Error())
		return ctrl.Result{Requeue: true}, nil
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		err = fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
		r.Recorder.Event(obj, "Warning", "Failed", err.Error())
		return ctrl.Result{}, nil
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID
	if obj.Spec.Container != nil &&
		len(strings.TrimSpace(*obj.Spec.Container)) != 0 {
		containerID = ""
		for _, container := range pod.Status.ContainerStatuses {
			if container.Name == *obj.Spec.Container {
				containerID = container.ContainerID
				break
			}
		}
		if len(containerID) == 0 {
			err = fmt.Errorf("cannot find container with name %s", *obj.Spec.Container)
			r.Recorder.Event(obj, "Warning", "Failed", err.Error())
			return ctrl.Result{}, nil
		}
	}

	actions, err := json.Marshal(obj.Spec.Actions)
	if err != nil {
		r.Recorder.Event(obj, "Warning", "Failed", err.Error())
		return ctrl.Result{Requeue: true}, nil
	}
	input := string(actions)
	r.Log.Info("input with", "config", input)

	res, err := pbClient.ApplyIOChaos(ctx, &pb.ApplyIOChaosRequest{
		Actions:     input,
		Volume:      obj.Spec.VolumeMountPath,
		ContainerId: containerID,

		Instance:  obj.Status.Pid,
		StartTime: obj.Status.StartTime,
		EnterNS:   true,
	})
	if err != nil {
		r.Recorder.Event(obj, "Warning", "Failed", err.Error())
		return ctrl.Result{Requeue: true}, nil
	}

	startTime = res.StartTime
	pid = res.Instance

	return ctrl.Result{}, nil
}
