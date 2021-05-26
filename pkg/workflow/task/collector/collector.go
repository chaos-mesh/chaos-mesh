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

package collector

import (
	"context"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Collector is a set of tools for collecting parameters/context from user defined task
type Collector interface {
	CollectContext(ctx context.Context) (env map[string]interface{}, err error)
}

type ComposeCollector struct {
	collectors []Collector
}

func (it *ComposeCollector) CollectContext(ctx context.Context) (env map[string]interface{}, err error) {
	if len(it.collectors) == 0 {
		return nil, nil
	}
	if len(it.collectors) == 1 {
		return it.collectors[0].CollectContext(ctx)
	}

	result := make(map[string]interface{})
	for _, collector := range it.collectors {
		temp, err := collector.CollectContext(ctx)
		if err != nil {
			return nil, err
		}
		mapExtend(result, temp)
	}
	return result, nil
}

// mapExtend will merge another map into the origin map, value with duplicated key will be replaced.
// origin map should not be nil
func mapExtend(origin map[string]interface{}, another map[string]interface{}) {
	if origin == nil || another == nil {
		return
	}
	for k, v := range another {
		origin[k] = v
	}
}

func DefaultCollector(kubeClient client.Client, restConfig *rest.Config, namespace, podName, containerName string) Collector {
	return &ComposeCollector{collectors: []Collector{
		NewExitCodeCollector(kubeClient, namespace, podName, containerName),
		NewStdoutCollector(restConfig, namespace, podName, containerName),
	}}
}
