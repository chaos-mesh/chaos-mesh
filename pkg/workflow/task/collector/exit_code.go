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

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ExitCode string = "exitCode"

type ExitCodeCollector struct {
	kubeClient    client.Client
	namespace     string
	podName       string
	containerName string
}

func NewExitCodeCollector(kubeClient client.Client, namespace string, podName string, containerName string) *ExitCodeCollector {
	return &ExitCodeCollector{kubeClient: kubeClient, namespace: namespace, podName: podName, containerName: containerName}
}

func (it *ExitCodeCollector) CollectContext(ctx context.Context) (env map[string]interface{}, err error) {
	var pod corev1.Pod
	err = it.kubeClient.Get(ctx, types.NamespacedName{
		Namespace: it.namespace,
		Name:      it.podName,
	}, &pod)

	if apierrors.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var targetContainerStatus corev1.ContainerStatus
	found := false
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == it.containerName {
			targetContainerStatus = containerStatus
			found = true
			break
		}
	}

	if !found {
		return nil, errors.Errorf("no such contaienr called %s in pod %s/%s", it.containerName, pod.Namespace, pod.Name)
	}

	if targetContainerStatus.State.Terminated == nil {
		return nil, errors.Errorf("container %s in pod %s/%s is waiting or running, not in ternimated", it.containerName, pod.Namespace, pod.Name)
	}

	return map[string]interface{}{
		ExitCode: targetContainerStatus.State.Terminated.ExitCode,
	}, nil
}
