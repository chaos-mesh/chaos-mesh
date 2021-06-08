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

package annotation

import (
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

const (
	// AnnotationPrefix defines the prefix of annotation key for chaos-mesh.
	AnnotationPrefix = "chaos-mesh"
)

func GenKeyForImage(pc *v1alpha1.PodChaos, containerName string, isInit bool) string {
	if isInit {
		containerName += "-init"
	} else {
		containerName += "-normal"
	}
	imageKey := fmt.Sprintf("%s-%s-%s-%s-image", AnnotationPrefix, pc.Name, pc.Spec.Action, containerName)

	// name part of annotation must be no more than 63 characters.
	// If the key is too long, we just use containerName as the key of annotation.
	if len(imageKey) > 63 {
		imageKey = containerName
	}

	return imageKey
}

func GenKeyForWebhook(prefix string, podName string) string {
	return fmt.Sprintf("%s-%s", prefix, podName)
}
