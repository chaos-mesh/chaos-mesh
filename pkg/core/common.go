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

package core

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubeObjectDesc defines a simple kube object description which uses in apiserver.
type KubeObjectDesc struct {
	metav1.TypeMeta `json:",inline"`
	Meta            KubeObjectMeta `json:"metadata"`
	Spec            interface{}    `json:"spec"`
}

// KubeObjectMetadata extracts the required fields from metav1.ObjectMeta.
type KubeObjectMeta struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}
