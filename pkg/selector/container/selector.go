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
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1/selector"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Option struct {
	clusterScoped bool
	targetNamespace string
	allowedNamespaces string
	ignoredNamespaces string
}

type SelectImpl struct {
	c client.Client
	r client.Reader

	Option
}

type Container struct {
	v1.Pod
	ContainerName string
}

func (impl *SelectImpl) Select(ctx context.Context, cs *selector.ContainerSelector) ([]interface{}, error) {
	pods, err := pod.SelectAndFilterPods(ctx, impl.c, impl.r, &cs.PodSelector, impl.clusterScoped, impl.targetNamespace, impl.allowedNamespaces, impl.ignoredNamespaces)
	if err != nil {
		return nil, err
	}

	containerNameMap := make(map[string]struct{})
	for _, name := range cs.ContainerNames {
		containerNameMap[name] = struct{}{}
	}

	var result []interface{}
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			if _, ok := containerNameMap[container.Name]; ok {
				result = append(result, Container {
					Pod: pod,
					ContainerName: container.Name,
				})
			}
		}
	}

	return result, nil
}

func NewSelectImpl(c client.Client, r client.Reader, clusterScoped bool, targetNamespace, allowedNamespaces, ignoredNamespaces string) *SelectImpl {
	return &SelectImpl{
		c,
		r,
		Option {
			clusterScoped,
			targetNamespace,
			allowedNamespaces,
			ignoredNamespaces,
		},
	}
}
