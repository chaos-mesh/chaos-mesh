// Copyright 2021 Chaos Mesh Authors.
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

package container

import (
	"context"

	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

type SelectImpl struct {
	c client.Client
	r client.Reader

	pod.Option
}

type Container struct {
	v1.Pod
	ContainerName string
}

func (c *Container) Id() string {
	return c.Pod.Namespace + "/" + c.Pod.Name + "/" + c.ContainerName
}

func (impl *SelectImpl) Select(ctx context.Context, cs *v1alpha1.ContainerSelector) ([]*Container, error) {
	pods, err := pod.SelectAndFilterPods(ctx, impl.c, impl.r, &cs.PodSelector, impl.ClusterScoped, impl.TargetNamespace, impl.EnableFilterNamespace)
	if err != nil {
		return nil, err
	}

	containerNameMap := make(map[string]struct{})
	for _, name := range cs.ContainerNames {
		containerNameMap[name] = struct{}{}
	}

	var result []*Container
	for _, pod := range pods {
		if len(cs.ContainerNames) == 0 {
			result = append(result, &Container{
				Pod:           pod,
				ContainerName: pod.Spec.Containers[0].Name,
			})
			continue
		}

		for _, container := range pod.Spec.Containers {
			if _, ok := containerNameMap[container.Name]; ok {
				result = append(result, &Container{
					Pod:           pod,
					ContainerName: container.Name,
				})
			}
		}
	}

	return result, nil
}

type Params struct {
	fx.In

	Client client.Client
	Reader client.Reader `name:"no-cache"`
}

func New(params Params) *SelectImpl {
	return &SelectImpl{
		params.Client,
		params.Reader,
		pod.Option{
			ClusterScoped:         config.ControllerCfg.ClusterScoped,
			TargetNamespace:       config.ControllerCfg.TargetNamespace,
			EnableFilterNamespace: config.ControllerCfg.EnableFilterNamespace,
		},
	}
}
