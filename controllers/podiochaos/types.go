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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/pkg/internalwatch"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	chaosDaemonClient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

// Reconciler applys podiochaos
type Reconciler struct {
	client.Client
	Log logr.Logger
}

type PodIOChaosReconcileContext struct {
	*Reconciler

	key types.NamespacedName

	context.Context

	updated bool
	chaos   *v1alpha1.PodIoChaos
}

func (r *Reconciler) NewContext(ctx context.Context, key types.NamespacedName) (*PodIOChaosReconcileContext, error) {
	chaos := &v1alpha1.PodIoChaos{}
	err := r.Client.Get(ctx, key, chaos)
	if err != nil {
		r.Log.Error(err, "fail to find podiochaos")
		return nil, err
	}

	internalwatch.Notify(chaos)

	return &PodIOChaosReconcileContext{
		key:        key,
		Context:    ctx,
		updated:    false,
		chaos:      chaos,
		Reconciler: r,
	}, nil
}

// Reconcile flushes io configuration on pod
func (r *Reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	// TODO: set the error information in the chaos status
	ctx, err := r.NewContext(context.TODO(), req.NamespacedName)
	if err != nil {
		r.Log.Error(err, "fail to construct reconciling context")
		return reconcile.Result{}, nil
	}

	err = ctx.Reconcile()
	if err != nil {
		r.Log.Error(err, "fail to reconcile podiochaos")
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

func (ctx *PodIOChaosReconcileContext) Reconcile() error {
	if ctx.chaos.Status.Sync {
		return nil
	}

	err := ctx.SyncIO()
	if err != nil {
		ctx.SetSync(false)
		ctx.SetFailedMessage(err.Error())
	} else {
		ctx.SetSync(true)
		ctx.SetFailedMessage("")
	}

	if ctx.updated {
		err := ctx.Update(ctx, ctx.chaos)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx *PodIOChaosReconcileContext) SetSync(sync bool) {
	if ctx.chaos.Status.Sync != sync {
		ctx.chaos.Status.Sync = sync
		ctx.updated = true
	}
}

func (ctx *PodIOChaosReconcileContext) SetPIDAndStartTime(pid int64, startTime int64) {
	if ctx.chaos.Spec.Pid != pid {
		ctx.chaos.Spec.Pid = pid
		ctx.updated = true
	}

	if ctx.chaos.Spec.StartTime != startTime {
		ctx.chaos.Spec.StartTime = startTime
		ctx.updated = true
	}
}

func (ctx *PodIOChaosReconcileContext) SetFailedMessage(failedMessage string) {
	if ctx.chaos.Status.FailedMessage != failedMessage {
		ctx.chaos.Status.FailedMessage = failedMessage
		ctx.updated = true
	}
}

func (ctx *PodIOChaosReconcileContext) SyncIO() error {
	pod := &v1.Pod{}

	err := ctx.Client.Get(ctx, types.NamespacedName{
		Name:      ctx.chaos.Name,
		Namespace: ctx.chaos.Namespace,
	}, pod)
	if err != nil {
		ctx.Log.Error(err, "fail to find pod")
		return err
	}

	pbClient, err := chaosDaemonClient.NewChaosDaemonClient(ctx, ctx, pod, config.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		err := fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
		ctx.Log.Error(err, "")
		return err
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID
	if ctx.chaos.Spec.Container != nil &&
		len(strings.TrimSpace(*ctx.chaos.Spec.Container)) != 0 {
		containerID = ""
		for _, container := range pod.Status.ContainerStatuses {
			if container.Name == *ctx.chaos.Spec.Container {
				containerID = container.ContainerID
				break
			}
		}
		if len(containerID) == 0 {
			err := fmt.Errorf("cannot find container with name %s", *ctx.chaos.Spec.Container)
			ctx.Log.Error(err, "")
			return err
		}
	}

	actions, err := json.Marshal(ctx.chaos.Spec.Actions)
	if err != nil {
		ctx.Log.Error(err, "fail to marshal actions")
		return nil
	}
	input := string(actions)
	ctx.Log.Info("input with", "config", input)

	res, err := pbClient.ApplyIoChaos(ctx, &pb.ApplyIoChaosRequest{
		Actions:     input,
		Volume:      ctx.chaos.Spec.VolumeMountPath,
		ContainerId: containerID,

		Instance:  ctx.chaos.Spec.Pid,
		StartTime: ctx.chaos.Spec.StartTime,
		EnterNS:   true,
	})
	if err != nil {
		ctx.Log.Error(err, "fail to apply iochaos")
		return nil
	}

	ctx.SetPIDAndStartTime(res.GetInstance(), res.GetStartTime())

	return nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.PodIoChaos{}).
		Complete()
	if err != nil {
		return err
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.PodIoChaos{}).
		Complete(r)

	return err
}
