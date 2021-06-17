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

package task

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

const (
	PodMetadataVolumeName            = "podmetadata"
	PodMetadataAnnotationsVolumePath = ""
	PodMetadataMountPath             = "/var/run/chaos-mesh/"
)

func SpawnPodForTask(task v1alpha1.Task) (corev1.PodSpec, error) {
	deepCopiedContainer := task.Container.DeepCopy()
	if len(deepCopiedContainer.Resources.Limits) == 0 {
		deepCopiedContainer.Resources.Limits.Cpu().SetMilli(1000)
		deepCopiedContainer.Resources.Limits.Memory().Set(1000)
	}
	result := corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyNever,
		Volumes:       attachVolumes(task),
		Containers: []corev1.Container{
			*deepCopiedContainer,
		},
	}
	return result, nil
}

func attachVolumes(task v1alpha1.Task) []corev1.Volume {
	var result []corev1.Volume

	// TODO: downwards API and configmaps

	result = append(result, task.Volumes...)
	return result
}
