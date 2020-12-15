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

package timechaos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	chaosdaemon "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

const timeChaosMsg = "time is shifted with %v"

// endpoint is time-chaos reconciler
type endpoint struct {
	ctx.Context
}

// Apply applies time-chaos
func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	timechaos, ok := chaos.(*v1alpha1.TimeChaos)
	if !ok {
		err := errors.New("chaos is not timechaos")
		r.Log.Error(err, "chaos is not TimeChaos", "chaos", chaos)
		return err
	}

	timechaos.SetDefaultValue()

	pods, err := utils.SelectAndFilterPods(ctx, r.Client, r.Reader, &timechaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)

	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}

	if err = r.applyAllPods(ctx, pods, timechaos); err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}

	timechaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Message:   fmt.Sprintf(timeChaosMsg, timechaos.Spec.TimeOffset),
		}

		timechaos.Status.Experiment.PodRecords = append(timechaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(timechaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

func (r *endpoint) recoverPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.TimeChaos) error {
	r.Log.Info("Try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	pbClient, err := utils.NewChaosDaemonClient(ctx, r.Client, pod, config.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	g := errgroup.Group{}
	expectedNames := make(map[string]bool)
	for _, name := range chaos.Spec.ContainerNames {
		expectedNames[name] = true
	}
	for index := range pod.Status.ContainerStatuses {
		container := pod.Status.ContainerStatuses[index]

		if len(expectedNames) == 0 || expectedNames[container.Name] {
			g.Go(func() error {
				err := r.recoverContainer(ctx, pbClient, container.ContainerID)

				if err != nil {
					r.Log.Error(err, "recover pod error", "namespace", pod.Namespace, "name", pod.Name)
				} else {
					r.Log.Info("Recover pod finished", "namespace", pod.Namespace, "name", pod.Name)
				}

				return err
			})
		}
	}

	return g.Wait()
}

func (r *endpoint) recoverContainer(ctx context.Context, client chaosdaemon.ChaosDaemonClient, containerID string) error {
	r.Log.Info("Try to recover time on container", "id", containerID)

	_, err := client.RecoverTimeOffset(ctx, &chaosdaemon.TimeRequest{
		ContainerId: containerID,
	})

	return err
}

// Object would return the instance of chaos
func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.TimeChaos{}
}

func (r *endpoint) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.TimeChaos) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		chaos.Finalizers = utils.InsertFinalizer(chaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, chaos)
		})
	}

	return g.Wait()
}

func (r *endpoint) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.TimeChaos) error {
	r.Log.Info("Try to shift time on pod", "namespace", pod.Namespace, "name", pod.Name)

	pbClient, err := utils.NewChaosDaemonClient(ctx, r.Client, pod, config.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	g := errgroup.Group{}
	expectedNames := make(map[string]bool)
	for _, name := range chaos.Spec.ContainerNames {
		expectedNames[name] = true
	}
	for index := range pod.Status.ContainerStatuses {
		container := pod.Status.ContainerStatuses[index]

		if len(expectedNames) == 0 || expectedNames[container.Name] {
			g.Go(func() error {
				return r.applyContainer(ctx, pbClient, container.ContainerID, chaos)
			})
		}
	}

	return g.Wait()
}

func (r *endpoint) applyContainer(ctx context.Context, client chaosdaemon.ChaosDaemonClient, containerID string, chaos *v1alpha1.TimeChaos) error {
	r.Log.Info("Try to shift time on container", "id", containerID)

	mask, err := utils.EncodeClkIds(chaos.Spec.ClockIds)
	if err != nil {
		return err
	}

	duration, err := time.ParseDuration(chaos.Spec.TimeOffset)
	if err != nil {
		return err
	}

	sec, nsec := secAndNSecFromDuration(duration)

	r.Log.Info("setting time shift", "mask", mask, "sec", sec, "nsec", nsec)
	_, err = client.SetTimeOffset(ctx, &chaosdaemon.TimeRequest{
		ContainerId: containerID,
		Sec:         sec,
		Nsec:        nsec,
		ClkIdsMask:  mask,
	})

	return err
}

func secAndNSecFromDuration(duration time.Duration) (sec int64, nsec int64) {
	sec = duration.Nanoseconds() / 1e9
	nsec = duration.Nanoseconds() - (sec * 1e9)

	return
}

func init() {
	router.Register("timechaos", &v1alpha1.TimeChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
