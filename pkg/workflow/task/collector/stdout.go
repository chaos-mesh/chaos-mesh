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
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const Stdout string = "stdout"

type StdoutCollector struct {
	restConfig    *rest.Config
	namespace     string
	podName       string
	containerName string
}

func NewStdoutCollector(restConfig *rest.Config, namespace string, podName string, containerName string) *StdoutCollector {
	return &StdoutCollector{restConfig: restConfig, namespace: namespace, podName: podName, containerName: containerName}
}

func (it *StdoutCollector) CollectContext(ctx context.Context) (env map[string]interface{}, err error) {
	client, err := kubernetes.NewForConfig(it.restConfig)
	if err != nil {
		return nil, err
	}

	request := client.CoreV1().Pods(it.namespace).GetLogs(it.podName, &v1.PodLogOptions{
		TypeMeta:  metav1.TypeMeta{},
		Container: it.containerName,
	}).Context(ctx)

	var bytes []byte
	if bytes, err = request.Do().Raw(); err != nil {
		return nil, err
	}
	stdout := strings.TrimSpace(string(bytes))

	return map[string]interface{}{Stdout: stdout}, nil
}
