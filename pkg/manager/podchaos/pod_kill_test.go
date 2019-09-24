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

package podchaos

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	NAMESPACE  = metav1.NamespaceDefault
	IDENTIFIER = "chaos-operator-id"
	KIND       = "Pod"
	NAME       = "name"
)

func newPod(name string, status v1.PodPhase) v1.Pod {
	return v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: NAMESPACE,
			Labels: map[string]string{
				"chaos-operator/identifier": IDENTIFIER,
			},
		},
		Status: v1.PodStatus{
			Phase: status,
		},
	}
}

func generateNPods(namePrefix string, n int, status v1.PodPhase) []runtime.Object {
	var pods []runtime.Object
	for i := 0; i < n; i++ {
		pod := newPod(fmt.Sprintf("%s%d", namePrefix, i), status)
		pods = append(pods, &pod)
	}

	return pods
}

func generateNRunningPods(namePrefix string, n int) []runtime.Object {
	return generateNPods(namePrefix, n, v1.PodRunning)
}
