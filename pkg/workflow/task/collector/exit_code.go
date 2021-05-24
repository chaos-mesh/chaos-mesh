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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const ExitCode string = "exitCode"

type ExitCodeCollector struct {
	pod           corev1.Pod
	containerName string
}

func (it *ExitCodeCollector) CollectContext() (env map[string]interface{}, err error) {

	var targetContainerStatus corev1.ContainerStatus
	found := false
	for _, containerStatus := range it.pod.Status.ContainerStatuses {
		if containerStatus.Name == it.containerName {
			targetContainerStatus = containerStatus
			found = true
			break
		}
	}

	if !found {
		return nil, errors.Errorf("no such contaienr called %s in pod %s/%s", it.containerName, it.pod.Namespace, it.pod.Name)
	}

	if targetContainerStatus.State.Terminated == nil {
		return nil, errors.Errorf("container %s in pod %s/%s is waiting or running, not in ternimated", it.containerName, it.pod.Namespace, it.pod.Name)
	}

	return map[string]interface{}{
		ExitCode: targetContainerStatus.State.Terminated.ExitCode,
	}, nil
}
