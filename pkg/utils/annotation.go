// Copyright 2019 PingCAP, Inc.
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

package utils

import (
	"fmt"

	"github.com/pingcap/chaos-operator/api/v1alpha1"
)

const (
	// AnnotationPrefix defines the prefix of annotation key for chaos-operator.
	AnnotationPrefix = "chaos-operator"
)

func GenAnnotationKeyForImage(pc *v1alpha1.PodChaos, containerName string) string {
	return fmt.Sprintf("%s-%s-%s-%s-image", AnnotationPrefix, pc.Name, pc.Spec.Action, containerName)
}

func GenAnnotationKeyForWebhook(prefix string, podName string) string {
	return fmt.Sprintf("%s-%s", prefix, podName)
}
