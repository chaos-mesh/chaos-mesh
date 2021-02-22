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

package containerkill

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

const (
	containerKillActionMsg = "delete container %s"
)

type endpoint struct {
	ctx.Context
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, obj v1alpha1.InnerObject) error {
	var err error

	podchaos, ok := obj.(*v1alpha1.PodChaos)
	if !ok {
		err = errors.New("chaos is not PodChaos")
		r.Log.Error(err, "chaos is not PodChaos", "chaos", obj)
		return err
	}

	if podchaos.Spec.ContainerName == "" {
		r.Log.Error(nil, "the name of container is empty", "name", req.Name, "namespace", req.Namespace)
		return fmt.Errorf("podchaos[%s/%s] the name of container is empty", podchaos.Namespace, podchaos.Name)
	}

	pods, err := selector.SelectAndFilterPods(ctx, r.Client, r.Reader, &podchaos.Spec, config.ControllerCfg.ClusterScoped, config.ControllerCfg.TargetNamespace, config.ControllerCfg.AllowedNamespaces, config.ControllerCfg.IgnoredNamespaces)
	if err != nil {
		r.Log.Error(err, "fail to select and filter pods")
		return err
	}

	var podsNotHaveContainer []string
	for podIndex := range pods {
		pod := &pods[podIndex]
		haveContainer := false

		for _, item := range pod.Status.ContainerStatuses {
			if item.Name == podchaos.Spec.ContainerName {
				haveContainer = true
				break
			}
		}

		if !haveContainer {
			podsNotHaveContainer = append(podsNotHaveContainer, pod.Name)
			r.Log.Error(nil, fmt.Sprintf("the pod %s doesn't have container %s", pod.Name, podchaos.Spec.ContainerName))
		}
	}

	if len(podsNotHaveContainer) != 0 {
		return fmt.Errorf("the pod %v doesn't have container %s", podsNotHaveContainer, podchaos.Spec.ContainerName)
	}

	g := errgroup.Group{}
	for podIndex := range pods {
		pod := &pods[podIndex]

		for containerIndex := range pod.Status.ContainerStatuses {
			containerName := pod.Status.ContainerStatuses[containerIndex].Name
			containerID := pod.Status.ContainerStatuses[containerIndex].ContainerID

			if containerName == podchaos.Spec.ContainerName {
				g.Go(func() error {
					err = r.KillContainer(ctx, pod, containerID)
					if err != nil {
						r.Log.Error(err, fmt.Sprintf(
							"failed to kill container: %s, pod: %s, namespace: %s",
							containerName, pod.Name, pod.Namespace))
					}
					return err
				})
			}
		}
	}

	if err := g.Wait(); err != nil {
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
			Message:   fmt.Sprintf(containerKillActionMsg, podchaos.Spec.ContainerName),
		}

		podchaos.Status.Experiment.PodRecords = append(podchaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(podchaos, v1.EventTypeNormal, events.ChaosRecovered, "")
	return nil
}

// Recover implements the reconciler.InnerReconciler.Recover
func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, obj v1alpha1.InnerObject) error {
	return nil
}

// Object implements the reconciler.InnerReconciler.Object
func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.PodChaos{}
}

// KillContainer kills container according to containerID
// Use client in chaos-daemon
func (r *endpoint) KillContainer(ctx context.Context, pod *v1.Pod, containerID string) error {
	r.Log.Info("Try to kill container", "namespace", pod.Namespace, "podName", pod.Name, "containerID", containerID)

	pbClient, err := client.NewChaosDaemonClient(ctx, r.Client, pod, config.ControllerCfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	if _, err = pbClient.ContainerKill(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_KILL,
		},
		ContainerId: containerID,
	}); err != nil {
		r.Log.Error(err, "kill container error", "namespace", pod.Namespace, "podName", pod.Name, "containerID", containerID)
		return err
	}

	return nil
}

func init() {
	router.Register("podchaos", &v1alpha1.PodChaos{}, func(obj runtime.Object) bool {
		chaos, ok := obj.(*v1alpha1.PodChaos)
		if !ok {
			return false
		}

		return chaos.Spec.Action == v1alpha1.ContainerKillAction
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
